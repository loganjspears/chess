package chess

import (
	"bufio"
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"io"
	"log"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
)

// Scanner is modeled on the bufio.Scanner type but
// instead of reading lines, it reads chess games
// from concatenated PGN files.  It is designed to
// replace GamesFromPGN in order to handle very large
// PGN database files such as https://database.lichess.org/.
type ScannerOpts struct {
	ExpandVariations bool
	MinELO           uint64
	MinTimeInSecs    uint64
}

var ScannerOptsDefault = ScannerOpts{
	ExpandVariations: false,
	MinELO:           0,
	MinTimeInSecs:    0,
}

type Scanner struct {
	games             []*Game
	scanBuf           []byte
	scanr             *bufio.Scanner
	err               error
	opts              ScannerOpts
	g                 *errgroup.Group
	ctx               context.Context
	numGamesProducers int32
	pgnChan           chan string
	// for efficiency we send/recv slices of games on the channel rather
	// than individual games
	gamesChan chan []*Game
}

type moveList struct {
	moveStrs []string
	outcome  Outcome
}

// NewScanner returns a new scanner with default options
func NewScanner(r io.Reader) *Scanner {
	return NewScannerWithOptions(r, ScannerOptsDefault)
}

// NewScanner returns a new scanner with explicit options
func NewScannerWithOptions(r io.Reader, o ScannerOpts) *Scanner {
	const IoBufferSize = 1024 * 1024 * 8 // in bytes
	const PgnBufferLen = 2048            // in number of pgn strings
	const GamesSliceBufferLen = 256      // in number of game _slices_

	ret := &Scanner{
		games:             make([]*Game, 0),
		scanBuf:           make([]byte, IoBufferSize),
		scanr:             bufio.NewScanner(r),
		err:               nil,
		opts:              o,
		numGamesProducers: int32(runtime.NumCPU()),
		pgnChan:           make(chan string, PgnBufferLen),
		gamesChan:         make(chan []*Game, GamesSliceBufferLen),
	}
	ret.scanr.Buffer(ret.scanBuf, cap(ret.scanBuf))

	// decode & read in the background so as to avoid starving the I/O
	// reading thread
	ret.g, ret.ctx = errgroup.WithContext(context.Background())
	ret.g.Go(ret.readPGNFiles)
	for ii := 0; int32(ii) < ret.numGamesProducers; ii++ {
		ret.g.Go(ret.processPGNs)
	}

	return ret
}

// Scan returns false if there was an error parsing
// a game or EOF was reached.  Running scan populates
// data for Next() and Err().
func (s *Scanner) Scan() bool {
	s.err = nil

	if len(s.games) > 0 {
		return true
	}

	var ok bool
	select {
	case s.games, ok = <-s.gamesChan:
		if !ok {
			// EOF
			return false
		}
	case <-s.ctx.Done():
		err := s.g.Wait()
		if err != nil {
			s.err = err
		}

		return false
	}

	return true
}

func (s *Scanner) readPGNFiles() error {
	defer close(s.pgnChan)

	var sb strings.Builder

	for {
		scan := s.scanr.Scan()
		if scan == false {
			return s.scanr.Err()
		}
		line := s.scanr.Text() + "\n"
		if strings.TrimSpace(line) == "" {
			continue
		} else {
			sb.WriteString(line)
		}

		if strings.HasPrefix(line, "1.") {
			select {
			case s.pgnChan <- sb.String():
			case <-s.ctx.Done():
				return s.ctx.Err()
			}

			sb.Reset()
		}
	}

	return nil
}

func (s *Scanner) processPGNs() error {
	defer func() {
		// last one out the door shuts the lights
		remaining := atomic.AddInt32(&s.numGamesProducers, -1)
		if remaining <= 0 {
			close(s.gamesChan)
		}
	}()

	for {
		select {
		case pgn, ok := <-s.pgnChan:
			if !ok {
				// no more pgns to process
				return nil
			}

			games, err := decodePGN(pgn, &s.opts)
			if err != nil {
				return err
			}

			if len(games) != 0 {
				select {
				case s.gamesChan <- games:
					games = make([]*Game, 0)
				case <-s.ctx.Done():
					return s.ctx.Err()
				}
			}
		case <-s.ctx.Done():
			return s.ctx.Err()
		}
	}

	return nil
}

// Next returns the game from the most recent Scan.
func (s *Scanner) Next() *Game {
	if len(s.games) == 0 {
		return nil
	}

	g := s.games[0]
	s.games = s.games[1:]

	return g
}

// Err returns an error encountered during scanning.
// Typically this will be a PGN parsing error or an
// io.EOF.
func (s *Scanner) Err() error {
	return s.err
}

// GamesFromPGN returns all PGN decoding games from the
// reader.  It is designed to be used decoding multiple PGNs
// in the same file.  An error is returned if there is an
// issue parsing the PGNs.
// Deprecated: Use Scanner instead.
func GamesFromPGN(r io.Reader) ([]*Game, error) {
	games := []*Game{}
	current := ""
	count := 0
	totalCount := 0
	br := bufio.NewReader(r)
	for {
		line, err := br.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		if strings.TrimSpace(line) == "" {
			count++
		} else {
			current += line
		}
		if count == 2 {
			newGames, err := decodePGN(current, &ScannerOptsDefault)
			if err != nil {
				return nil, err
			}
			games = append(games, newGames[0])
			count = 0
			current = ""
			totalCount++
			log.Println("Processed game", totalCount)
		}
	}
	return games, nil
}

type multiDecoder []Decoder

func (a multiDecoder) Decode(pos *Position, s string) (*Move, error) {
	for _, d := range a {
		m, err := d.Decode(pos, s)
		if err == nil {
			return m, nil
		}
	}
	return nil, fmt.Errorf(`chess: failed to decode notation text "%s" for position %s`, s, pos)
}

func shouldFilterGame(pgn string, opts *ScannerOpts, tagPairs []*TagPair) (bool, error) {
	eloCount := 0
	foundTime := false

	for _, tp := range tagPairs {
		if opts.MinELO > 0 &&
			(strings.ToLower(tp.Key) == "whiteelo" ||
				strings.ToLower(tp.Key) == "blackelo") {
			elo, err := strconv.ParseUint(tp.Value, 10, 64)
			if err != nil {
				return false, fmt.Errorf("chess: could not parse ELO from tag %s: %v", err.Error(), tp.Key)
			}
			if elo < opts.MinELO {
				return true, nil
			}
			eloCount++
		}
		if opts.MinTimeInSecs > 0 && strings.ToLower(tp.Key) == "timecontrol" {
			// example formats:
			// 300+3 -> 5 minute game plus 3 seconds per move
			// 0+2 -> 0 minute game plus 2 seconds per move
			// 600+5 -> 10 minute game plus 5 seconds per move
			// "-" -> unlimited/correspondance
			if tp.Value == "-" {
				foundTime = true
				continue
			} // else
			timeFields := strings.Split(tp.Value, "+")
			if len(timeFields) == 0 {
				return false, fmt.Errorf("chess: could not parse TimeControl from value %s", tp.Value)
			}
			baseTime, err := strconv.ParseUint(timeFields[0], 10, 64)
			if err != nil {
				return false, fmt.Errorf("chess: could not parse TimeControl from value %s: %v", tp.Value, err)
			}
			if baseTime < opts.MinTimeInSecs {
				return true, nil
			}
			foundTime = true
		}
	}

	if opts.MinELO > 0 && eloCount != 2 {
		return true, nil
	}
	if opts.MinTimeInSecs > 0 && !foundTime {
		return true, nil
	}

	return false, nil
}

func decodePGN(pgn string, opts *ScannerOpts) ([]*Game, error) {
	tagPairs := getTagPairs(pgn)

	gameFuncs := []func(*Game){}
	for _, tp := range tagPairs {
		if strings.ToLower(tp.Key) == "fen" {
			fenFunc, err := FEN(tp.Value)
			if err != nil {
				return nil, fmt.Errorf("chess: pgn decode error %s on tag %s", err.Error(), tp.Key)
			}
			gameFuncs = append(gameFuncs, fenFunc)
			break
		}
	}

	shouldFilter, err := shouldFilterGame(pgn, opts, tagPairs)
	if err != nil {
		return nil, err
	}
	if shouldFilter {
		return nil, nil
	}

	moveLists, err := getMoveLists(pgn, opts.ExpandVariations)
	if err != nil {
		return nil, err
	}

	gameFuncs = append(gameFuncs, TagPairs(tagPairs))
	games := make([]*Game, 0)
	varName := "main line"

	for idx, moveList := range moveLists {
		if idx > 0 {
			varName = fmt.Sprintf("variation %v", idx)
		}
		g := NewGame(gameFuncs...)
		g.ignoreAutomaticDraws = true
		decoder := multiDecoder([]Decoder{AlgebraicNotation{}, LongAlgebraicNotation{}, UCINotation{}})
		for _, alg := range moveList.moveStrs {
			m, err := decoder.Decode(g.Position(), alg)
			if err != nil {
				return nil, fmt.Errorf("chess: pgn decode error %s on move %d of %v", err.Error(), g.Position().moveCount, varName)
			}
			if err := g.Move(m); err != nil {
				return nil, fmt.Errorf("chess: pgn invalid move error %s on move %d of %v", err.Error(), g.Position().moveCount, varName)
			}
		}
		g.outcome = moveList.outcome
		games = append(games, g)
	}

	return games, nil
}

func encodePGN(g *Game) string {
	s := ""
	for _, tag := range g.tagPairs {
		s += fmt.Sprintf("[%s \"%s\"]\n", tag.Key, tag.Value)
	}
	if s != "" {
		s += "\n"
	}
	for i, move := range g.moves {
		pos := g.positions[i]
		txt := g.notation.Encode(pos, move)
		if i%2 == 0 {
			s += fmt.Sprintf("%d.%s", (i/2)+1, txt)
		} else {
			s += fmt.Sprintf(" %s ", txt)
		}
	}
	s += " " + string(g.outcome)
	return s
}

var (
	tagPairRegex = regexp.MustCompile(`\[(.*)\s\"(.*)\"\]`)
)

func getTagPairs(pgn string) []*TagPair {
	tagPairs := []*TagPair{}
	matches := tagPairRegex.FindAllString(pgn, -1)
	for _, m := range matches {
		results := tagPairRegex.FindStringSubmatch(m)
		if len(results) == 3 {
			pair := &TagPair{
				Key:   results[1],
				Value: results[2],
			}
			tagPairs = append(tagPairs, pair)
		}
	}
	return tagPairs
}

var (
	moveNumRegex = regexp.MustCompile(`(?:\d+\.+)?(.*)`)
)

func getMoveLists(pgn string, expandVariations bool) ([]moveList, error) {
	// remove comments
	text := removeSection("{", "}", pgn)
	// remove tag pairs
	text = removeSection(`\[`, `\]`, text)
	// remove line breaks
	text = strings.Replace(text, "\n", " ", -1)

	var err error
	if !expandVariations {
		text, err = removeVariations(text)
		if err != nil {
			return nil, err
		}
	} else {
		text = strings.ReplaceAll(text, "(", "( ")
		text = strings.ReplaceAll(text, ")", " )")
	}

	list := strings.Split(text, " ")

	moveListIdx := 0
	moveListIdxStack := make([]int, 0)
	moveLists := make([]moveList, 0)
	moveLists = append(moveLists, moveList{moveStrs: make([]string, 0)})

	for _, move := range list {
		move = strings.TrimSpace(move)
		switch move {
		case string(NoOutcome), string(WhiteWon), string(BlackWon), string(Draw):
			moveLists[moveListIdx].outcome = Outcome(move)
		case "":
		case "(":
			// begin new variation
			moveListIdxStack = append(moveListIdxStack, moveListIdx)
			newIdx := len(moveLists)
			numMoves := len(moveLists[moveListIdx].moveStrs) - 1
			moveLists = append(moveLists, moveList{moveStrs: make([]string, numMoves)})
			copy(moveLists[newIdx].moveStrs, moveLists[moveListIdx].moveStrs)
			moveListIdx = newIdx

		case ")":
			// end current variation
			stackSize := len(moveListIdxStack)
			if stackSize == 0 {
				return nil, fmt.Errorf("Failed to parse variation")
			}
			moveListIdx = moveListIdxStack[stackSize-1]
			moveListIdxStack = moveListIdxStack[:stackSize-1]
		default:
			results := moveNumRegex.FindStringSubmatch(move)
			if len(results) == 2 && results[1] != "" {
				moveLists[moveListIdx].moveStrs = append(moveLists[moveListIdx].moveStrs, results[1])
			}
		}
	}
	return moveLists, nil
}

func removeSection(leftChar, rightChar, s string) string {
	r := regexp.MustCompile(leftChar + ".*?" + rightChar)
	for {
		i := r.FindStringIndex(s)
		if i == nil {
			return s
		}
		s = s[0:i[0]] + s[i[1]:]
	}
}

func removeVariations(s string) (string, error) {
	var ret strings.Builder
	variationDepth := 0

	// cannot just use regexp since variations may be nested
	for _, c := range s {
		if c == '(' {
			variationDepth++
			continue
		}
		if c == ')' {
			if variationDepth <= 0 {
				return "", fmt.Errorf("Mismatched parenthesis in variation: %v", s)
			}
			variationDepth--
			continue
		}
		if variationDepth == 0 {
			_, err := ret.WriteRune(c)
			if err != nil {
				return "", err
			}
		}
	}

	return ret.String(), nil
}
