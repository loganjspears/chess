package chess

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strconv"
	"strings"
)

// A Board represents a chess board and its relationship between squares and pieces.
type Board struct {
	bbWhiteKing   bitboard
	bbWhiteQueen  bitboard
	bbWhiteRook   bitboard
	bbWhiteBishop bitboard
	bbWhiteKnight bitboard
	bbWhitePawn   bitboard
	bbBlackKing   bitboard
	bbBlackQueen  bitboard
	bbBlackRook   bitboard
	bbBlackBishop bitboard
	bbBlackKnight bitboard
	bbBlackPawn   bitboard
	whiteSqs      bitboard
	blackSqs      bitboard
	emptySqs      bitboard
	whiteKingSq   Square
	blackKingSq   Square
}

// NewBoard returns a board from a square to piece mapping.
func NewBoard(m map[Square]Piece) *Board {
	b := &Board{}
	for _, p1 := range allPieces {
		bm := map[Square]bool{}
		for sq, p2 := range m {
			if p1 == p2 {
				bm[sq] = true
			}
		}
		bb := newBitboard(bm)
		b.setBBForPiece(p1, bb)
	}
	b.calcConvienceBBs(nil)
	return b
}

// SquareMap returns a mapping of squares to pieces.  A square is only added to the map if it is occupied.
func (b *Board) SquareMap() map[Square]Piece {
	m := map[Square]Piece{}
	for sq := 0; sq < numOfSquaresInBoard; sq++ {
		p := b.Piece(Square(sq))
		if p != NoPiece {
			m[Square(sq)] = p
		}
	}
	return m
}

// Rotate rotates the board 90 degrees clockwise.
func (b *Board) Rotate() *Board {
	return b.Flip(UpDown).Transpose()
}

// FlipDirection is the direction for the Board.Flip method
type FlipDirection int

const (
	// UpDown flips the board's rank values
	UpDown FlipDirection = iota
	// LeftRight flips the board's file values
	LeftRight
)

// Flip flips the board over the vertical or hoizontal
// center line.
func (b *Board) Flip(fd FlipDirection) *Board {
	m := map[Square]Piece{}
	for sq := 0; sq < numOfSquaresInBoard; sq++ {
		var mv Square
		switch fd {
		case UpDown:
			file := Square(sq).File()
			rank := Rank(7 - Square(sq).Rank())
			mv = getSquare(file, rank)
		case LeftRight:
			file := File(7 - Square(sq).File())
			rank := Square(sq).Rank()
			mv = getSquare(file, rank)
		}
		m[mv] = b.Piece(Square(sq))
	}
	return NewBoard(m)
}

// Transpose flips the board over the A8 to H1 diagonal.
func (b *Board) Transpose() *Board {
	m := map[Square]Piece{}
	for sq := 0; sq < numOfSquaresInBoard; sq++ {
		file := File(7 - Square(sq).Rank())
		rank := Rank(7 - Square(sq).File())
		mv := getSquare(file, rank)
		m[mv] = b.Piece(Square(sq))
	}
	return NewBoard(m)
}

// Draw returns visual representation of the board useful for debugging.
func (b *Board) Draw() string {
	s := "\n A B C D E F G H\n"
	for r := 7; r >= 0; r-- {
		s += Rank(r).String()
		for f := 0; f < numOfSquaresInRow; f++ {
			p := b.Piece(getSquare(File(f), Rank(r)))
			if p == NoPiece {
				s += "-"
			} else {
				s += p.String()
			}
			s += " "
		}
		s += "\n"
	}
	return s
}

// String implements the fmt.Stringer interface and returns
// a string in the FEN board format: rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR
func (b *Board) String() string {
	fen := ""
	for r := 7; r >= 0; r-- {
		for f := 0; f < numOfSquaresInRow; f++ {
			sq := getSquare(File(f), Rank(r))
			p := b.Piece(sq)
			if p != NoPiece {
				fen += p.getFENChar()
			} else {
				fen += "1"
			}
		}
		if r != 0 {
			fen += "/"
		}
	}
	for i := 8; i > 1; i-- {
		repeatStr := strings.Repeat("1", i)
		countStr := strconv.Itoa(i)
		fen = strings.Replace(fen, repeatStr, countStr, -1)
	}
	return fen
}

// Piece returns the piece for the given square.
func (b *Board) Piece(sq Square) Piece {
	for _, p := range allPieces {
		bb := b.bbForPiece(p)
		if bb.Occupied(sq) {
			return p
		}
	}
	return NoPiece
}

// MarshalText implements the encoding.TextMarshaler interface and returns
// a string in the FEN board format: rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR
func (b *Board) MarshalText() (text []byte, err error) {
	return []byte(b.String()), nil
}

// UnmarshalText implements the encoding.TextUnarshaler interface and takes
// a string in the FEN board format: rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR
func (b *Board) UnmarshalText(text []byte) error {
	cp, err := fenBoard(string(text))
	if err != nil {
		return err
	}
	*b = *cp
	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface and returns
// the bitboard representations as a array of bytes.  Bitboads are encoded
// in the following order: WhiteKing, WhiteQueen, WhiteRook, WhiteBishop, WhiteKnight
// WhitePawn, BlackKing, BlackQueen, BlackRook, BlackBishop, BlackKnight, BlackPawn
func (b *Board) MarshalBinary() (data []byte, err error) {
	bbs := []bitboard{b.bbWhiteKing, b.bbWhiteQueen, b.bbWhiteRook, b.bbWhiteBishop, b.bbWhiteKnight, b.bbWhitePawn,
		b.bbBlackKing, b.bbBlackQueen, b.bbBlackRook, b.bbBlackBishop, b.bbBlackKnight, b.bbBlackPawn}
	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, bbs)
	return buf.Bytes(), err
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface and parses
// the bitboard representations as a array of bytes.  Bitboads are decoded
// in the following order: WhiteKing, WhiteQueen, WhiteRook, WhiteBishop, WhiteKnight
// WhitePawn, BlackKing, BlackQueen, BlackRook, BlackBishop, BlackKnight, BlackPawn
func (b *Board) UnmarshalBinary(data []byte) error {
	if len(data) != 96 {
		return errors.New("chess: invalid number of bytes for board unmarshal binary")
	}
	b.bbWhiteKing = bitboard(binary.BigEndian.Uint64(data[:8]))
	b.bbWhiteQueen = bitboard(binary.BigEndian.Uint64(data[8:16]))
	b.bbWhiteRook = bitboard(binary.BigEndian.Uint64(data[16:24]))
	b.bbWhiteBishop = bitboard(binary.BigEndian.Uint64(data[24:32]))
	b.bbWhiteKnight = bitboard(binary.BigEndian.Uint64(data[32:40]))
	b.bbWhitePawn = bitboard(binary.BigEndian.Uint64(data[40:48]))
	b.bbBlackKing = bitboard(binary.BigEndian.Uint64(data[48:56]))
	b.bbBlackQueen = bitboard(binary.BigEndian.Uint64(data[56:64]))
	b.bbBlackRook = bitboard(binary.BigEndian.Uint64(data[64:72]))
	b.bbBlackBishop = bitboard(binary.BigEndian.Uint64(data[72:80]))
	b.bbBlackKnight = bitboard(binary.BigEndian.Uint64(data[80:88]))
	b.bbBlackPawn = bitboard(binary.BigEndian.Uint64(data[88:96]))
	b.calcConvienceBBs(nil)
	return nil
}

func (b *Board) update(m *Move) {
	p1 := b.Piece(m.s1)
	s1BB := bbForSquare(m.s1)
	s2BB := bbForSquare(m.s2)

	legacyCastle := false
	if m.HasTag(KingSideCastle) || m.HasTag(QueenSideCastle) {
		// We support both legacy and 960 style moves.
		if p1.Color() == White {
			if b.bbWhiteRook.Occupied(m.s2) {
				// Clear out rook, it'll be placed back later.
				b.bbWhiteRook = b.bbWhiteRook & ^s2BB
				// Correct the king destination.
				if m.HasTag(QueenSideCastle) {
					s2BB = bbForSquare(getSquare(FileC, m.s2.Rank()))
				} else {
					s2BB = bbForSquare(getSquare(FileG, m.s2.Rank()))
				}
			} else {
				legacyCastle = true
			}
		} else {
			if b.bbBlackRook.Occupied(m.s2) {
				// Clear out rook, it'll be placed back later.
				b.bbBlackRook = b.bbBlackRook & ^s2BB
				// Correct the king destination.
				if m.HasTag(QueenSideCastle) {
					s2BB = bbForSquare(getSquare(FileC, m.s2.Rank()))
				} else {
					s2BB = bbForSquare(getSquare(FileG, m.s2.Rank()))
				}
			} else {
				legacyCastle = true
			}
		}
	}

	// move s1 piece to s2
	for _, p := range allPieces {
		bb := b.bbForPiece(p)
		// remove what was at s2
		b.setBBForPiece(p, bb & ^s2BB)
		// move what was at s1 to s2
		if bb.Occupied(m.s1) {
			bb = b.bbForPiece(p)
			b.setBBForPiece(p, (bb & ^s1BB)|s2BB)
		}
	}
	// check promotion
	if m.promo != NoPieceType {
		newPiece := getPiece(m.promo, p1.Color())
		// remove pawn
		bbPawn := b.bbForPiece(p1)
		b.setBBForPiece(p1, bbPawn & ^s2BB)
		// add promo piece
		bbPromo := b.bbForPiece(newPiece)
		b.setBBForPiece(newPiece, bbPromo|s2BB)
	}
	// remove captured en passant piece
	if m.HasTag(EnPassant) {
		if p1.Color() == White {
			b.bbBlackPawn = ^(bbForSquare(m.s2) << 8) & b.bbBlackPawn
		} else {
			b.bbWhitePawn = ^(bbForSquare(m.s2) >> 8) & b.bbWhitePawn
		}
	}
	if legacyCastle {
		// move rook for castle
		if p1.Color() == White && m.HasTag(KingSideCastle) {
			b.bbWhiteRook = (b.bbWhiteRook & ^bbForSquare(H1) | bbForSquare(F1))
		} else if p1.Color() == White && m.HasTag(QueenSideCastle) {
			b.bbWhiteRook = (b.bbWhiteRook & ^bbForSquare(A1)) | bbForSquare(D1)
		} else if p1.Color() == Black && m.HasTag(KingSideCastle) {
			b.bbBlackRook = (b.bbBlackRook & ^bbForSquare(H8) | bbForSquare(F8))
		} else if p1.Color() == Black && m.HasTag(QueenSideCastle) {
			b.bbBlackRook = (b.bbBlackRook & ^bbForSquare(A8)) | bbForSquare(D8)
		}
	} else {
		// put back rook for castle
		if p1.Color() == White && m.HasTag(KingSideCastle) {
			b.bbWhiteRook = (b.bbWhiteRook | bbForSquare(F1))
		} else if p1.Color() == White && m.HasTag(QueenSideCastle) {
			b.bbWhiteRook = (b.bbWhiteRook | bbForSquare(D1))
		} else if p1.Color() == Black && m.HasTag(KingSideCastle) {
			b.bbBlackRook = (b.bbBlackRook | bbForSquare(F8))
		} else if p1.Color() == Black && m.HasTag(QueenSideCastle) {
			b.bbBlackRook = (b.bbBlackRook | bbForSquare(D8))
		}
	}
	b.calcConvienceBBs(m)
}

func (b *Board) calcConvienceBBs(m *Move) {
	whiteSqs := b.bbWhiteKing | b.bbWhiteQueen | b.bbWhiteRook | b.bbWhiteBishop | b.bbWhiteKnight | b.bbWhitePawn
	blackSqs := b.bbBlackKing | b.bbBlackQueen | b.bbBlackRook | b.bbBlackBishop | b.bbBlackKnight | b.bbBlackPawn
	emptySqs := ^(whiteSqs | blackSqs)
	b.whiteSqs = whiteSqs
	b.blackSqs = blackSqs
	b.emptySqs = emptySqs
	// In non-legacy castling, s2 is not the destination for the king.
	if m == nil || m.HasTag(QueenSideCastle) || m.HasTag(KingSideCastle) {
		b.whiteKingSq = NoSquare
		b.blackKingSq = NoSquare

		for sq := 0; sq < numOfSquaresInBoard; sq++ {
			sqr := Square(sq)
			if b.bbWhiteKing.Occupied(sqr) {
				b.whiteKingSq = sqr
			} else if b.bbBlackKing.Occupied(sqr) {
				b.blackKingSq = sqr
			}
		}
	} else if m.s1 == b.whiteKingSq {
		b.whiteKingSq = m.s2
	} else if m.s1 == b.blackKingSq {
		b.blackKingSq = m.s2
	}
}

func (b *Board) copy() *Board {
	return &Board{
		whiteSqs:      b.whiteSqs,
		blackSqs:      b.blackSqs,
		emptySqs:      b.emptySqs,
		whiteKingSq:   b.whiteKingSq,
		blackKingSq:   b.blackKingSq,
		bbWhiteKing:   b.bbWhiteKing,
		bbWhiteQueen:  b.bbWhiteQueen,
		bbWhiteRook:   b.bbWhiteRook,
		bbWhiteBishop: b.bbWhiteBishop,
		bbWhiteKnight: b.bbWhiteKnight,
		bbWhitePawn:   b.bbWhitePawn,
		bbBlackKing:   b.bbBlackKing,
		bbBlackQueen:  b.bbBlackQueen,
		bbBlackRook:   b.bbBlackRook,
		bbBlackBishop: b.bbBlackBishop,
		bbBlackKnight: b.bbBlackKnight,
		bbBlackPawn:   b.bbBlackPawn,
	}
}

func (b *Board) isOccupied(sq Square) bool {
	return !b.emptySqs.Occupied(sq)
}

func (b *Board) hasSufficientMaterial() bool {
	// queen, rook, or pawn exist
	if (b.bbWhiteQueen | b.bbWhiteRook | b.bbWhitePawn |
		b.bbBlackQueen | b.bbBlackRook | b.bbBlackPawn) > 0 {
		return true
	}
	// if king is missing then it is a test
	if b.bbWhiteKing == 0 || b.bbBlackKing == 0 {
		return true
	}
	count := map[PieceType]int{}
	pieceMap := b.SquareMap()
	for _, p := range pieceMap {
		count[p.Type()]++
	}
	// 	king versus king
	if count[Bishop] == 0 && count[Knight] == 0 {
		return false
	}
	// king and bishop versus king
	if count[Bishop] == 1 && count[Knight] == 0 {
		return false
	}
	// king and knight versus king
	if count[Bishop] == 0 && count[Knight] == 1 {
		return false
	}
	// king and bishop(s) versus king and bishop(s) with the bishops on the same colour.
	if count[Knight] == 0 {
		whiteCount := 0
		blackCount := 0
		for sq, p := range pieceMap {
			if p.Type() == Bishop {
				switch sq.color() {
				case White:
					whiteCount++
				case Black:
					blackCount++
				}
			}
		}
		if whiteCount == 0 || blackCount == 0 {
			return false
		}
	}
	return true
}

func (b *Board) bbForPiece(p Piece) bitboard {
	switch p {
	case WhiteKing:
		return b.bbWhiteKing
	case WhiteQueen:
		return b.bbWhiteQueen
	case WhiteRook:
		return b.bbWhiteRook
	case WhiteBishop:
		return b.bbWhiteBishop
	case WhiteKnight:
		return b.bbWhiteKnight
	case WhitePawn:
		return b.bbWhitePawn
	case BlackKing:
		return b.bbBlackKing
	case BlackQueen:
		return b.bbBlackQueen
	case BlackRook:
		return b.bbBlackRook
	case BlackBishop:
		return b.bbBlackBishop
	case BlackKnight:
		return b.bbBlackKnight
	case BlackPawn:
		return b.bbBlackPawn
	}
	return bitboard(0)
}

func (b *Board) setBBForPiece(p Piece, bb bitboard) {
	switch p {
	case WhiteKing:
		b.bbWhiteKing = bb
	case WhiteQueen:
		b.bbWhiteQueen = bb
	case WhiteRook:
		b.bbWhiteRook = bb
	case WhiteBishop:
		b.bbWhiteBishop = bb
	case WhiteKnight:
		b.bbWhiteKnight = bb
	case WhitePawn:
		b.bbWhitePawn = bb
	case BlackKing:
		b.bbBlackKing = bb
	case BlackQueen:
		b.bbBlackQueen = bb
	case BlackRook:
		b.bbBlackRook = bb
	case BlackBishop:
		b.bbBlackBishop = bb
	case BlackKnight:
		b.bbBlackKnight = bb
	case BlackPawn:
		b.bbBlackPawn = bb
	default:
		panic("invalid piece")
	}
}
