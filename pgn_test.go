package chess

import (
	"os"
	"sort"
	"strings"
	"testing"
)

type pgnTest struct {
	PostPos *Position
	PGN     string
}

var (
	validPGNs = []pgnTest{
		{
			PostPos: unsafeFEN("4r3/6P1/2p2P1k/1p6/pP2p1R1/P1B5/2P2K2/3r4 b - - 0 45"),
			PGN: `[Event "?"]
            [Site "?"]
            [Date "1997.05.03"]
            [Round "1"]
            [White "Kasparov"]
            [Black "Deep-Blue"]
            [Result "1-0"]
            [WhiteElo "2795"]

            1. Nf3 d5 2. g3 Bg4 3. b3 Nd7 4. Bb2 e6 5. Bg2 Ngf6 6. O-O c6
            7. d3 Bd6 8. Nbd2 O-O 9. h3 Bh5 10. e3 h6 11. Qe1 Qa5 12. a3
            Bc7 13. Nh4 g5 14. Nhf3 e5 15. e4 Rfe8 16. Nh2 Qb6 17. Qc1 a5
            18. Re1 Bd6 19. Ndf1 dxe4 20. dxe4 Bc5 21. Ne3 Rad8 22. Nhf1 g4
            23. hxg4 Nxg4 24. f3 Nxe3 25. Nxe3 Be7 26. Kh1 Bg5 27. Re2 a4
            28. b4 f5 29. exf5 e4 30. f4 Bxe2 31. fxg5 Ne5 32. g6 Bf3 33. Bc3
            Qb5 34. Qf1 Qxf1+ 35. Rxf1 h5 36. Kg1 Kf8 37. Bh3 b5 38. Kf2 Kg7
            39. g4 Kh6 40. Rg1 hxg4 41. Bxg4 Bxg4 42. Nxg4+ Nxg4+ 43. Rxg4
            Rd5 44. f6 Rd1 45. g7 1-0`,
		},
		{
			PostPos: unsafeFEN("4r3/6P1/2p2P1k/1p6/pP2p1R1/P1B5/2P2K2/3r4 b - - 0 45"),
			PGN: `[Event "?"]
[Site "http://lichess.org/4HXJOtpN"]
[Date "1997.05.03"]
[White "Kasparov (2795)"]
[Black "Deep-Blue"]
[Result "1-0"]
[WhiteElo "?"]
[BlackElo "?"]
[PlyCount "89"]
[Variant "Standard"]
[TimeControl "-"]
[ECO "A07"]
[Opening "King's Indian Attack, General"]
[Termination "Normal"]
[Annotator "lichess.org"]

1. Nf3 d5 2. g3 { King's Indian Attack, General } Bg4 3. b3 Nd7 4. Bb2 e6 5. Bg2 Ngf6 6. O-O c6 7. d3 Bd6 8. Nbd2 O-O 9. h3 Bh5 10. e3 h6 11. Qe1 Qa5 12. a3 Bc7 13. Nh4 g5 14. Nhf3 e5 15. e4 Rfe8 16. Nh2 Qb6 17. Qc1 a5 18. Re1 Bd6 19. Ndf1 dxe4 20. dxe4 Bc5 21. Ne3 Rad8 22. Nhf1 g4 23. hxg4 Nxg4 24. f3 Nxe3 25. Nxe3 Be7 26. Kh1 Bg5 27. Re2 a4 28. b4 f5 29. exf5 e4 30. f4 Bxe2 31. fxg5 Ne5 32. g6 Bf3 33. Bc3 Qb5 34. Qf1 Qxf1+ 35. Rxf1 h5 36. Kg1 Kf8 37. Bh3 b5 38. Kf2 Kg7?! { (0.70 → 1.52) Inaccuracy. The best move was Rd6. } (38... Rd6 39. Re1 Re7 40. Rg1 Re8 41. g4 h4 42. g5 Kg8 43. Rf1 Kf8 44. Re1 Kg8 45. Rb1 Rdd8 46. Rf1 Kf8 47. Rg1 Kg8 48. Rb1) 39. g4 Kh6?! { (1.42 → 2.07) Inaccuracy. The best move was h4. } (39... h4 40. g5 Kf8 41. Bg2 Ng4+ 42. Nxg4 Bxg4 43. Ke3 h3 44. Rf4 hxg2 45. Rxg4 Rd1 46. Rxg2 Rf1 47. Rf2 Rg1 48. Bf6 Rh1 49. Kd4) 40. Rg1 hxg4 41. Bxg4 Bxg4 42. Nxg4+ Nxg4+ 43. Rxg4 Rd5 44. f6 Rd1?? { (1.60 → 8.36) Blunder. The best move was Rf5+. } (44... Rf5+ 45. Ke2 Rg8 46. g7 e3 47. Bd4 Kh7 48. Kxe3 Re8+ 49. Re4 Rxe4+ 50. Kxe4 Rf1 51. Ke5 Kg8 52. Ke6 Rf4 53. c3 Rf1 54. Ke7) 45. g7 { Black resigns } 1-0`,
		},
		{
			PostPos: unsafeFEN("2r2rk1/pp1bBpp1/2np4/2pp2p1/1bP5/1P4P1/P1QPPPBP/3R1RK1 b - - 0 3"),
			PGN: `[Event "?"]
			[Site "?"]
			[Date "????.??.??"]
			[Round "?"]
			[White "N.N."]
			[Black "N.N."]
			[Result "1-0"]
			[Annotator "T1R"]
			[SetUp "1"]
			[FEN "2r2rk1/pp1bqpp1/2nppn1p/2p3N1/1bP5/1PN3P1/PBQPPPBP/3R1RK1 w - - 0 1"]
			[PlyCount "5"]
			[EventType "swiss"]
			
			1. Nd5 exd5 (1... hxg5 2. Nxe7+ Nxe7) 2. Bxf6 hxg5 3. Bxe7 1-0`,
		},
	}
)

func TestValidPGNs(t *testing.T) {
	for _, test := range validPGNs {
		games, err := decodePGN(test.PGN, &ScannerOptsDefault)
		if err != nil {
			t.Fatalf("recieved unexpected pgn error %s", err.Error())
		}
		if len(games) != 1 {
			t.Fatalf("expected to parse 1 game but parsed %v", len(games))
		}
		if games[0].Position().String() != test.PostPos.String() {
			t.Fatalf("expected board to be \n%s\nFEN:%s\n but got \n%s\n\nFEN:%s\n",
				test.PostPos.board.Draw(), test.PostPos.String(),
				games[0].Position().board.Draw(), games[0].Position().String())
		}
	}
}

func BenchmarkPGN(b *testing.B) {
	pgn := `[Event "?"]
	[Site "?"]
	[Date "1997.05.03"]
	[Round "1"]
	[White "Kasparov"]
	[Black "Deep-Blue"]
	[Result "1-0"]
	[WhiteElo "2795"]

	1. Nf3 d5 2. g3 Bg4 3. b3 Nd7 4. Bb2 e6 5. Bg2 Ngf6 6. O-O c6
	7. d3 Bd6 8. Nbd2 O-O 9. h3 Bh5 10. e3 h6 11. Qe1 Qa5 12. a3
	Bc7 13. Nh4 g5 14. Nhf3 e5 15. e4 Rfe8 16. Nh2 Qb6 17. Qc1 a5
	18. Re1 Bd6 19. Ndf1 dxe4 20. dxe4 Bc5 21. Ne3 Rad8 22. Nhf1 g4
	23. hxg4 Nxg4 24. f3 Nxe3 25. Nxe3 Be7 26. Kh1 Bg5 27. Re2 a4
	28. b4 f5 29. exf5 e4 30. f4 Bxe2 31. fxg5 Ne5 32. g6 Bf3 33. Bc3
	Qb5 34. Qf1 Qxf1+ 35. Rxf1 h5 36. Kg1 Kf8 37. Bh3 b5 38. Kf2 Kg7
	39. g4 Kh6 40. Rg1 hxg4 41. Bxg4 Bxg4 42. Nxg4+ Nxg4+ 43. Rxg4
	Rd5 44. f6 Rd1 45. g7 1-0`

	for n := 0; n < b.N; n++ {
		opt, _ := PGN(strings.NewReader(pgn))
		NewGame(opt)
	}
}

func TestScanner(t *testing.T) {
	f, err := os.OpenFile("./assets/scan_test.pgn", os.O_RDONLY, 0600)
	if err != nil {
		t.Fatalf("Failed to open scanner test pgn: %v", err)
	}
	defer f.Close()

	scanner := NewScanner(f)

	expectedFinalFENs := []string{"r1bqkb1r/2pp1ppp/p1n2n2/1p2p3/4P3/1B1P1N2/PPP2PPP/RNBQK2R b KQkq - 0 6", "rnbqkb1r/pp2pppp/5n2/3p4/2PP4/2N5/PP3PPP/R1BQKBNR b KQkq - 2 5"}
	actualFinalFENs := []string{}

	for scanner.Scan() {
		g := scanner.Next()
		if g == nil {
			t.Fatalf("Scan.Next() failed to return game")
		}
		actualFinalFENs = append(actualFinalFENs, g.Position().String())
	}

	if scanner.Err() != nil {
		t.Fatalf("Unexpected non-nil Err(): %v", scanner.Err())
	}

	if len(actualFinalFENs) != len(expectedFinalFENs) {
		t.Fatalf("Expecting %v games but only parsed %v games", len(expectedFinalFENs), len(actualFinalFENs))
	}

	sort.Strings(expectedFinalFENs)
	sort.Strings(actualFinalFENs)

	for idx, _ := range expectedFinalFENs {
		if expectedFinalFENs[idx] != actualFinalFENs[idx] {
			t.Fatalf("Mismatched final FEN idx %v expected %v got %v", idx, expectedFinalFENs[idx], actualFinalFENs[idx])
		}
	}
}

func TestScannerWithOpts(t *testing.T) {
	f, err := os.OpenFile("./assets/scan_test.pgn", os.O_RDONLY, 0600)
	if err != nil {
		t.Fatalf("Failed to open scanner test pgn: %v", err)
	}
	defer f.Close()

	scanner := NewScannerWithOptions(f, ScannerOpts{ExpandVariations: true})

	expectedFinalFENs := []string{"r1bqkb1r/2pp1ppp/p1n2n2/1p2p3/4P3/1B1P1N2/PPP2PPP/RNBQK2R b KQkq - 0 6",
		"r1bqkb1r/pppp1ppp/2n2n2/1B2p3/4P3/5N2/PPPP1PPP/RNBQ1RK1 b kq - 5 4",
		"r2qkbnr/pppb1ppp/2np4/1B6/3NP3/8/PPP2PPP/RNBQK2R w KQkq - 1 6",
		"r1bqkbnr/2pp1ppp/p7/1p2p3/3NP3/1B6/PPPP1PPP/RNBQK2R b KQkq - 0 6",
		"rnbqkb1r/pp2pppp/5n2/3p4/2PP4/2N5/PP3PPP/R1BQKBNR b KQkq - 2 5",
		"rn1qkbnr/pp3ppp/4p3/2ppPb2/3P4/5N2/PPP1BPPP/RNBQK2R w KQkq - 0 6",
		"rn1qkbnr/pp2ppp1/2p5/3pP2p/3P3P/3b4/PPP2PP1/RNBQK1NR w KQkq - 0 6",
		"rn1qkbnr/pp3ppp/2p1p1b1/3pP3/3P2P1/2N5/PPP2P1P/R1BQKBNR w KQkq - 1 6"}
	actualFinalFENs := []string{}

	for scanner.Scan() {
		g := scanner.Next()
		if g == nil {
			t.Fatalf("Scan.Next() failed to return game")
		}
		actualFinalFENs = append(actualFinalFENs, g.Position().String())
	}

	if scanner.Err() != nil {
		t.Fatalf("Unexpected non-nil Err(): %v", scanner.Err())
	}

	if len(actualFinalFENs) != len(expectedFinalFENs) {
		t.Fatalf("Expecting %v games but only parsed %v games", len(expectedFinalFENs), len(actualFinalFENs))
	}

	sort.Strings(expectedFinalFENs)
	sort.Strings(actualFinalFENs)

	for idx, _ := range expectedFinalFENs {
		if expectedFinalFENs[idx] != actualFinalFENs[idx] {
			t.Fatalf("Mismatched final FEN idx %v expected %v got %v", idx, expectedFinalFENs[idx], actualFinalFENs[idx])
		}
	}
}

type filttestcase struct {
	opts         ScannerOpts
	expectedECOs []string
}

func runOneScanFilterTestcase(t *testing.T, tcase *filttestcase) {
	f, err := os.OpenFile("./assets/scan_filter_test.pgn", os.O_RDONLY, 0600)
	if err != nil {
		t.Fatalf("Failed to open scanner filter test pgn: %v", err)
	}
	defer f.Close()

	scanner := NewScannerWithOptions(f, tcase.opts)

	actualECOs := []string{}
	var g *Game
	for scanner.Scan() {
		g = scanner.Next()
		if g == nil {
			t.Fatalf("Scan.Next() failed to return game")
		}
		tagPair := g.GetTagPair("ECO")
		if tagPair == nil {
			t.Fatalf("Scan failed to parse ECO tag")
		}
		actualECOs = append(actualECOs, tagPair.Value)
	}

	if scanner.Err() != nil {
		t.Fatalf("Unexpected non-nil Err(): %v", scanner.Err())
	}

	if len(actualECOs) != len(tcase.expectedECOs) {
		t.Fatalf("Expecting %v games but only parsed %v games with opts %v", len(tcase.expectedECOs), len(actualECOs), tcase.opts)
	}

	sort.Strings(actualECOs)
	sort.Strings(tcase.expectedECOs)

	for idx, _ := range tcase.expectedECOs {
		if tcase.expectedECOs[idx] != actualECOs[idx] {
			t.Fatalf("Mismatched ECO idx %v expected %v got %v", idx, tcase.expectedECOs[idx], actualECOs[idx])
		}
	}
}

func TestScannerWithFilters(t *testing.T) {
	cases := []filttestcase{
		{opts: ScannerOpts{
			ExpandVariations: true,
			MinTimeInSecs:    0,
			MinELO:           0},
			expectedECOs: []string{"C50", "C40", "A13", "B10"}},
		{opts: ScannerOpts{
			ExpandVariations: false,
			MinTimeInSecs:    0,
			MinELO:           0},
			expectedECOs: []string{"C50", "C40", "A13", "B10"}},
		{opts: ScannerOpts{
			ExpandVariations: true,
			MinTimeInSecs:    300,
			MinELO:           0},
			expectedECOs: []string{"C50", "A13"}},
		{opts: ScannerOpts{
			ExpandVariations: false,
			MinTimeInSecs:    300,
			MinELO:           0},
			expectedECOs: []string{"C50", "A13"}},
		{opts: ScannerOpts{
			ExpandVariations: true,
			MinTimeInSecs:    0,
			MinELO:           1800},
			expectedECOs: []string{"C50", "B10"}},
		{opts: ScannerOpts{
			ExpandVariations: false,
			MinTimeInSecs:    0,
			MinELO:           1800},
			expectedECOs: []string{"C50", "B10"}},
		{opts: ScannerOpts{
			ExpandVariations: true,
			MinTimeInSecs:    300,
			MinELO:           1800},
			expectedECOs: []string{"C50"}},
		{opts: ScannerOpts{
			ExpandVariations: false,
			MinTimeInSecs:    300,
			MinELO:           1800},
			expectedECOs: []string{"C50"}},
		{opts: ScannerOpts{
			ExpandVariations: false,
			MinTimeInSecs:    99999,
			MinELO:           4000},
			expectedECOs: []string{}},
	}

	for _, tcase := range cases {
		runOneScanFilterTestcase(t, &tcase)
	}
}
