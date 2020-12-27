package chess

import (
	"crypto/md5"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Side represents a side of the board.
type Side int

const (
	// KingSide is the right side of the board from white's perspective.
	KingSide Side = iota + 1
	// QueenSide is the left side of the board from white's perspective.
	QueenSide
)

// CastleRights holds the state of both sides castling abilities.
type CastleRights string

// CanCastle returns true if the given color and side combination
// can castle, otherwise returns false.
func (cr CastleRights) CanCastle(board *Board, c Color, side Side) bool {
	if string(cr) == "-" {
		return false
	}
	char := "k"
	if side == QueenSide {
		char = "q"
	}
	if c == White {
		char = strings.ToUpper(char)
	}
	if strings.Contains(string(cr), char) {
		return true
	}
	// Check for file rights.
	for _, r := range string(cr) {
		if (c == White) != unicode.IsUpper(r) {
			continue
		}
		letter := strings.ToLower(fmt.Sprintf("%c", r))
		if letter == "q" || letter == "k" {
			continue
		}
		kingFile := board.whiteKingSq.File()
		if c != White {
			kingFile = board.blackKingSq.File()
		}
		kingFileLetter := fileChars[kingFile : kingFile+1]
		if (side == QueenSide) == (letter < kingFileLetter) {
			return true
		}
	}
	return false
}

// String implements the fmt.Stringer interface and returns
// a FEN compatible string.  Ex. KQq
func (cr CastleRights) String() string {
	return string(cr)
}

// Position represents the state of the game without reguard
// to its outcome.  Position is translatable to FEN notation.
type Position struct {
	board           *Board
	turn            Color
	castleRights    CastleRights
	enPassantSquare Square
	halfMoveClock   int
	moveCount       int
	inCheck         bool
	validMoves      []*Move
}

// Update returns a new position resulting from the given move.
// The move itself isn't validated, if validation is needed use
// Game's Move method.  This method is more performant for bots that
// rely on the ValidMoves because it skips redundant validation.
func (pos *Position) Update(m *Move) *Position {
	moveCount := pos.moveCount
	if pos.turn == Black {
		moveCount++
	}
	ncr := pos.updateCastleRights(m)
	p := pos.board.Piece(m.s1)
	halfMove := pos.halfMoveClock
	if p.Type() == Pawn || m.HasTag(Capture) {
		halfMove = 0
	} else {
		halfMove++
	}
	b := pos.board.copy()
	b.update(m)
	return &Position{
		board:           b,
		turn:            pos.turn.Other(),
		castleRights:    ncr,
		enPassantSquare: pos.updateEnPassantSquare(m),
		halfMoveClock:   halfMove,
		moveCount:       moveCount,
		inCheck:         m.HasTag(Check),
	}
}

// ValidMoves returns a list of valid moves for the position.
func (pos *Position) ValidMoves() []*Move {
	if pos.validMoves != nil {
		return append([]*Move(nil), pos.validMoves...)
	}
	pos.validMoves = engine{}.CalcMoves(pos, false)
	return append([]*Move(nil), pos.validMoves...)
}

// Status returns the position's status as one of the outcome methods.
// Possible returns values include Checkmate, Stalemate, and NoMethod.
func (pos *Position) Status() Method {
	return engine{}.Status(pos)
}

// Board returns the position's board.
func (pos *Position) Board() *Board {
	return pos.board
}

// Turn returns the color to move next.
func (pos *Position) Turn() Color {
	return pos.turn
}

// CastleRights returns the castling rights of the position.
func (pos *Position) CastleRights() CastleRights {
	return pos.castleRights
}

// String implements the fmt.Stringer interface and returns a
// string with the FEN format: rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1
func (pos *Position) String() string {
	b := pos.board.String()
	t := pos.turn.String()
	c := pos.castleRights.String()
	sq := "-"
	if pos.enPassantSquare != NoSquare {
		sq = pos.enPassantSquare.String()
	}
	return fmt.Sprintf("%s %s %s %s %d %d", b, t, c, sq, pos.halfMoveClock, pos.moveCount)
}

// Hash returns a unique hash of the position
func (pos *Position) Hash() [16]byte {
	sq := "-"
	if pos.enPassantSquare != NoSquare {
		sq = pos.enPassantSquare.String()
	}
	s := pos.turn.String() + ":" + pos.castleRights.String() + ":" + sq
	for _, p := range allPieces {
		bb := pos.board.bbForPiece(p)
		s += ":" + strconv.FormatUint(uint64(bb), 16)
	}
	return md5.Sum([]byte(s))
}

// MarshalText implements the encoding.TextMarshaler interface and
// encodes the position's FEN.
func (pos *Position) MarshalText() (text []byte, err error) {
	return []byte(pos.String()), nil
}

// UnmarshalText implements the encoding.TextUnarshaler interface and
// assumes the data is in the FEN format.
func (pos *Position) UnmarshalText(text []byte) error {
	cp, err := decodeFEN(string(text))
	if err != nil {
		return err
	}
	pos.board = cp.board
	pos.turn = cp.turn
	pos.castleRights = cp.castleRights
	pos.enPassantSquare = cp.enPassantSquare
	pos.halfMoveClock = cp.halfMoveClock
	pos.moveCount = cp.moveCount
	pos.inCheck = isInCheck(cp)
	return nil
}

func (pos *Position) copy() *Position {
	return &Position{
		board:           pos.board.copy(),
		turn:            pos.turn,
		castleRights:    pos.castleRights,
		enPassantSquare: pos.enPassantSquare,
		halfMoveClock:   pos.halfMoveClock,
		moveCount:       pos.moveCount,
		inCheck:         pos.inCheck,
	}
}

func (pos *Position) updateCastleRights(m *Move) CastleRights {
	cr := string(pos.castleRights)
	if cr == "-" {
		return pos.castleRights
	}
	p := pos.board.Piece(m.s1)
	if p == WhiteKing {
		new := ""
		for _, r := range cr {
			if unicode.IsLower(r) {
				new += string(r)
			}
		}
		cr = new
	} else if p == BlackKing {
		new := ""
		for _, r := range cr {
			if unicode.IsUpper(r) {
				new += string(r)
			}
		}
		cr = new
	} else {
		// Rook move or Capture
		// TODO: Should change back to KQkq format when the last outside non-castling rook moves away from rank.
		p2 := pos.board.Piece(m.s2)
		if p == WhiteRook || p2 == WhiteRook {
			sq := m.s1
			if p2 == WhiteRook {
				sq = m.s2
			}
			if sq.Rank() == Rank1 || sq == m.s1 && m.s2.Rank() == Rank1 {
				new := ""
				kingFile := pos.board.whiteKingSq.File()
				for _, r := range cr {
					keep := true
					if unicode.IsLower(r) {
					} else if string(r) == "K" {
						for file := FileH; file > kingFile; file -= 1 {
							if pos.board.bbWhiteRook.Occupied(getSquare(file, Rank1)) {
								if sq.File() == file {
									keep = false
								}
								if sq.Rank() != Rank1 && sq.File() > file {
									keep = false
									// No longer the outer, needs to change to A-H format.
									new += strings.ToUpper(fileChars[file : file+1])
								}
								break
							}
						}
					} else if string(r) == "Q" {
						for file := FileA; file < kingFile; file += 1 {
							if pos.board.bbWhiteRook.Occupied(getSquare(file, Rank1)) {
								if sq.File() == file {
									keep = false
								}
								if sq.Rank() != Rank1 && sq.File() < file {
									keep = false
									// No longer the outer, needs to change to A-H format.
									new += strings.ToUpper(fileChars[file : file+1])
								}
								break
							}
						}
					} else {
						if fileChars[sq.File():sq.File()+1] == strings.ToLower(string(r)) {
							keep = false
						}
					}
					if keep {
						new += string(r)
					}
				}
				cr = new
			}
		}
		if p == BlackRook || p2 == BlackRook {
			sq := m.s1
			if p2 == BlackRook {
				sq = m.s2
			}
			if sq.Rank() == Rank8 || sq == m.s1 && m.s2.Rank() == Rank8 {
				new := ""
				kingFile := pos.board.blackKingSq.File()
				for _, r := range cr {
					keep := true
					if unicode.IsUpper(r) {
					} else if string(r) == "k" {
						for file := FileH; file > kingFile; file -= 1 {
							if pos.board.bbBlackRook.Occupied(getSquare(file, Rank8)) {
								if sq.File() == file {
									keep = false
								}
								if sq.Rank() != Rank8 && sq.File() > file {
									keep = false
									// No longer the outer, needs to change to a-h format.
									new += fileChars[file : file+1]
								}
								break
							}
						}
					} else if string(r) == "q" {
						for file := FileA; file < kingFile; file += 1 {
							if pos.board.bbBlackRook.Occupied(getSquare(file, Rank8)) {
								if sq.File() == file {
									keep = false
								}
								if sq.Rank() != Rank8 && sq.File() < file {
									keep = false
									// No longer the outer, needs to change to a-h format.
									new += fileChars[file : file+1]
								}
								break
							}
						}
					} else {
						if fileChars[sq.File():sq.File()+1] == string(r) {
							keep = false
						}
					}
					if keep {
						new += string(r)
					}
				}
				cr = new
			}
		}
	}
	if cr == "" {
		cr = "-"
	}
	return CastleRights(cr)
}

func (pos *Position) updateEnPassantSquare(m *Move) Square {
	p := pos.board.Piece(m.s1)
	if p.Type() != Pawn {
		return NoSquare
	}
	if pos.turn == White &&
		(bbForSquare(m.s1)&bbRank2) != 0 &&
		(bbForSquare(m.s2)&bbRank4) != 0 {
		return Square(m.s2 - 8)
	} else if pos.turn == Black &&
		(bbForSquare(m.s1)&bbRank7) != 0 &&
		(bbForSquare(m.s2)&bbRank5) != 0 {
		return Square(m.s2 + 8)
	}
	return NoSquare
}

func (pos *Position) samePosition(pos2 *Position) bool {
	return pos.board.String() == pos2.board.String() &&
		pos.turn == pos2.turn &&
		pos.castleRights.String() == pos2.castleRights.String() &&
		pos.enPassantSquare == pos2.enPassantSquare
}
