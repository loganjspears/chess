package chess

import (
	"strconv"
	"strings"
)

// bitboard is a board representation encoded in an unsigned 64-bit integer.  The
// 64 board positions begin with A1 as the most significant bit and H8 as the least.
type bitboard uint64

func newBitboard(m map[Square]bool) bitboard {
	var bb uint64
	for sq := 0; sq < numOfSquaresInBoard; sq++ {
		bb <<= 1
		if m[Square(sq)] {
			bb |= 1
		}
	}
	return bitboard(bb)
}

func (b bitboard) Mapping() map[Square]bool {
	m := map[Square]bool{}
	for sq := 0; sq < numOfSquaresInBoard; sq++ {
		if b&bbForSquare(Square(sq)) > 0 {
			m[Square(sq)] = true
		}
	}
	return m
}

// String returns a 64 character string of 1s and 0s starting with the most significant bit.
func (b bitboard) String() string {
	s := strconv.FormatUint(uint64(b), 2)
	return strings.Repeat("0", numOfSquaresInBoard-len(s)) + s
}

// Draw returns visual representation of the bitboard useful for debugging.
func (b bitboard) Draw() string {
	s := "\n A B C D E F G H\n"
	for r := 7; r >= 0; r-- {
		s += Rank(r).String()
		for f := 0; f < numOfSquaresInRow; f++ {
			sq := getSquare(File(f), Rank(r))
			if b.Occupied(sq) {
				s += "1"
			} else {
				s += "0"
			}
			s += " "
		}
		s += "\n"
	}
	return s
}
