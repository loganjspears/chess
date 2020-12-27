package chess

import (
	"log"
	"testing"
)

type moveTest struct {
	pos     *Position
	m       *Move
	postPos *Position
}

var (
	validMoves = []moveTest{
		// pawn moves
		{m: &Move{s1: E2, s2: E4}, pos: unsafeFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")},
		{m: &Move{s1: A2, s2: A3}, pos: unsafeFEN("8/8/8/8/8/8/P7/8 w - - 0 1")},
		{m: &Move{s1: A7, s2: A6}, pos: unsafeFEN("8/p7/8/8/8/8/8/8 b - - 0 1")},
		{m: &Move{s1: A7, s2: A5}, pos: unsafeFEN("8/p7/8/8/8/8/8/8 b - - 0 1")},
		{m: &Move{s1: C4, s2: B5}, pos: unsafeFEN("8/8/8/1p1p4/2P5/8/8/8 w - - 0 1")},
		{m: &Move{s1: C4, s2: D5}, pos: unsafeFEN("8/8/8/1p1p4/2P5/8/8/8 w - - 0 1")},
		{m: &Move{s1: C4, s2: C5}, pos: unsafeFEN("8/8/8/1p1p4/2P5/8/8/8 w - - 0 1")},
		{m: &Move{s1: C5, s2: B4}, pos: unsafeFEN("8/8/8/2p5/1P1P4/8/8/8 b - - 0 1")},
		{m: &Move{s1: C5, s2: D4}, pos: unsafeFEN("8/8/8/2p5/1P1P4/8/8/8 b - - 0 1")},
		{m: &Move{s1: C5, s2: C4}, pos: unsafeFEN("8/8/8/2p5/1P1P4/8/8/8 b - - 0 1")},
		{m: &Move{s1: A4, s2: B3}, pos: unsafeFEN("2r3k1/1q1nbppp/r3p3/3pP3/pPpP4/P1Q2N2/2RN1PPP/2R4K b - b3 0 23")},
		{m: &Move{s1: A2, s2: A1, promo: Queen}, pos: unsafeFEN("8/8/8/8/8/8/p7/8 b - - 0 1")},
		{m: &Move{s1: E7, s2: E6}, pos: unsafeFEN("r2qkbnr/pppnpppp/8/3p4/6b1/1P3NP1/PBPPPP1P/RN1QKB1R b KQkq - 2 4")},
		// knight moves
		{m: &Move{s1: E4, s2: F6}, pos: unsafeFEN("8/8/8/3pp3/4N3/8/5B2/8 w - - 0 1")},
		{m: &Move{s1: E4, s2: D6}, pos: unsafeFEN("8/8/8/3pp3/4N3/8/5B2/8 w - - 0 1")},
		{m: &Move{s1: E4, s2: G5}, pos: unsafeFEN("8/8/8/3pp3/4N3/8/5B2/8 w - - 0 1")},
		{m: &Move{s1: E4, s2: G3}, pos: unsafeFEN("8/8/8/3pp3/4N3/8/5B2/8 w - - 0 1")},
		{m: &Move{s1: E4, s2: D2}, pos: unsafeFEN("8/8/8/3pp3/4N3/8/5B2/8 w - - 0 1")},
		{m: &Move{s1: E4, s2: C3}, pos: unsafeFEN("8/8/8/3pp3/4N3/8/5B2/8 w - - 0 1")},
		{m: &Move{s1: E4, s2: C5}, pos: unsafeFEN("8/8/8/3pp3/4N3/8/5B2/8 w - - 0 1")},
		{m: &Move{s1: B8, s2: D7}, pos: unsafeFEN("rn1qkb1r/pp3ppp/2p1pn2/3p4/2PP4/2NQPN2/PP3PPP/R1B1K2R b KQkq - 0 7")},
		{m: &Move{s1: F6, s2: E4}, pos: unsafeFEN("r1b1k2r/ppp2ppp/2p2n2/4N3/4P3/2P5/PPP2PPP/R1BK3R b kq - 0 8")},
		// bishop moves
		{m: &Move{s1: E4, s2: H7}, pos: unsafeFEN("8/8/8/3pp3/4B3/5N2/8/8 w - - 0 1")},
		{m: &Move{s1: E4, s2: D5}, pos: unsafeFEN("8/8/8/3pp3/4B3/5N2/8/8 w - - 0 1")},
		{m: &Move{s1: E4, s2: B1}, pos: unsafeFEN("8/8/8/3pp3/4B3/5N2/8/8 w - - 0 1")},
		// rook moves
		{m: &Move{s1: B2, s2: B4}, pos: unsafeFEN("8/1p5b/4N3/4p3/8/8/1R6/1B6 w - - 0 1")},
		{m: &Move{s1: B2, s2: B7}, pos: unsafeFEN("8/1p5b/4N3/4p3/8/8/1R6/1B6 w - - 0 1")},
		{m: &Move{s1: B2, s2: A2}, pos: unsafeFEN("8/1p5b/4N3/4p3/8/8/1R6/1B6 w - - 0 1")},
		{m: &Move{s1: B2, s2: H2}, pos: unsafeFEN("8/1p5b/4N3/4p3/8/8/1R6/1B6 w - - 0 1")},
		{m: &Move{s1: E1, s2: E8}, pos: unsafeFEN("r3r1k1/p4p1p/3p4/1p4p1/2pP4/2P2P2/PP3P1P/R3RK2 w - g6 0 22")},
		// queen moves
		{m: &Move{s1: B2, s2: E5}, pos: unsafeFEN("8/1p5b/4N3/4p3/8/8/1Q6/1B6 w - - 0 1")},
		{m: &Move{s1: B2, s2: A1}, pos: unsafeFEN("8/1p5b/4N3/4p3/8/8/1Q6/1B6 w - - 0 1")},
		{m: &Move{s1: B2, s2: A2}, pos: unsafeFEN("8/1p5b/4N3/4p3/8/8/1Q6/1B6 w - - 0 1")},
		{m: &Move{s1: B2, s2: H2}, pos: unsafeFEN("8/1p5b/4N3/4p3/8/8/1Q6/1B6 w - - 0 1")},
		{m: &Move{s1: D8, s2: D1}, pos: unsafeFEN("r1bqk2r/ppp2ppp/2p2n2/4N3/4P3/2P5/PPP2PPP/R1BQK2R b KQkq - 0 7")},
		// king moves
		{m: &Move{s1: E4, s2: E5}, pos: unsafeFEN("5r2/8/8/8/4K3/8/8/8 w - - 0 1")},
		{m: &Move{s1: E4, s2: E3}, pos: unsafeFEN("5r2/8/8/8/4K3/8/8/8 w - - 0 1")},
		{m: &Move{s1: E4, s2: D3}, pos: unsafeFEN("5r2/8/8/8/4K3/8/8/8 w - - 0 1")},
		{m: &Move{s1: E4, s2: D4}, pos: unsafeFEN("5r2/8/8/8/4K3/8/8/8 w - - 0 1")},
		{m: &Move{s1: E4, s2: D5}, pos: unsafeFEN("5r2/8/8/8/4K3/8/8/8 w - - 0 1")},
		{m: &Move{s1: E4, s2: E5}, pos: unsafeFEN("5r2/8/8/8/4K3/8/8/8 w - - 0 1")},
		// castleing
		{m: &Move{s1: E1, s2: G1, tags: KingSideCastle}, pos: unsafeFEN("r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1")},
		{m: &Move{s1: E1, s2: C1, tags: QueenSideCastle}, pos: unsafeFEN("r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1")},
		{m: &Move{s1: E8, s2: G8, tags: KingSideCastle}, pos: unsafeFEN("r3k2r/8/8/8/8/8/8/R3K2R b KQkq - 0 1")},
		{m: &Move{s1: E8, s2: C8, tags: QueenSideCastle}, pos: unsafeFEN("r3k2r/8/8/8/8/8/8/R3K2R b KQkq - 0 1")},
		{m: &Move{s1: E1, s2: H1}, pos: unsafeFEN("r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1")},
		{m: &Move{s1: E1, s2: A1}, pos: unsafeFEN("r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1")},
		{m: &Move{s1: E8, s2: H8}, pos: unsafeFEN("r3k2r/8/8/8/8/8/8/R3K2R b KQkq - 0 1")},
		{m: &Move{s1: E8, s2: A8}, pos: unsafeFEN("r3k2r/8/8/8/8/8/8/R3K2R b KQkq - 0 1")},
		// king moving in front of enemy pawn http://en.lichess.org/4HXJOtpN#75
		{m: &Move{s1: F8, s2: G7}, pos: unsafeFEN("3rrk2/8/2p3P1/1p2nP1p/pP2p3/P1B1NbPB/2P2K2/5R2 b - - 1 38")},
	}

	invalidMoves = []moveTest{
		// out of turn moves
		{m: &Move{s1: E7, s2: E5}, pos: unsafeFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")},
		{m: &Move{s1: E2, s2: E4}, pos: unsafeFEN("rnbqkbnr/1ppppppp/p7/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1")},
		// pawn moves
		{m: &Move{s1: E2, s2: D3}, pos: unsafeFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")},
		{m: &Move{s1: E2, s2: F3}, pos: unsafeFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")},
		{m: &Move{s1: E2, s2: E5}, pos: unsafeFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")},
		{m: &Move{s1: A2, s2: A1}, pos: unsafeFEN("8/8/8/8/8/8/p7/8 b - - 0 1")},
		{m: &Move{s1: E6, s2: E5}, pos: unsafeFEN(`2b1r3/2k2p1B/p2np3/4B3/8/5N2/PP1K1PPP/3R4 b - - 2 1`)},
		{m: &Move{s1: H7, s2: H5}, pos: unsafeFEN(`2bqkbnr/rpppp2p/2n2p2/p5pB/5P2/4P3/PPPP2PP/RNBQK1NR b KQk - 4 6`)},
		// knight moves
		{m: &Move{s1: E4, s2: F2}, pos: unsafeFEN("8/8/8/3pp3/4N3/8/5B2/8 w - - 0 1")},
		{m: &Move{s1: E4, s2: F3}, pos: unsafeFEN("8/8/8/3pp3/4N3/8/5B2/8 w - - 0 1")},
		// bishop moves
		{m: &Move{s1: E4, s2: C6}, pos: unsafeFEN("8/8/8/3pp3/4B3/5N2/8/8 w - - 0 1")},
		{m: &Move{s1: E4, s2: E5}, pos: unsafeFEN("8/8/8/3pp3/4B3/5N2/8/8 w - - 0 1")},
		{m: &Move{s1: E4, s2: E4}, pos: unsafeFEN("8/8/8/3pp3/4B3/5N2/8/8 w - - 0 1")},
		{m: &Move{s1: E4, s2: F3}, pos: unsafeFEN("8/8/8/3pp3/4B3/5N2/8/8 w - - 0 1")},
		// rook moves
		{m: &Move{s1: B2, s2: B1}, pos: unsafeFEN("8/1p5b/4N3/4p3/8/8/1R6/1B6 w - - 0 1")},
		{m: &Move{s1: B2, s2: C3}, pos: unsafeFEN("8/1p5b/4N3/4p3/8/8/1R6/1B6 w - - 0 1")},
		{m: &Move{s1: B2, s2: B8}, pos: unsafeFEN("8/1p5b/4N3/4p3/8/8/1R6/1B6 w - - 0 1")},
		{m: &Move{s1: B2, s2: G7}, pos: unsafeFEN("8/1p5b/4N3/4p3/8/8/1R6/1B6 w - - 0 1")},
		// queen moves
		{m: &Move{s1: B2, s2: B1}, pos: unsafeFEN("8/1p5b/4N3/4p3/8/8/1Q6/1B6 w - - 0 1")},
		{m: &Move{s1: B2, s2: C4}, pos: unsafeFEN("8/1p5b/4N3/4p3/8/8/1Q6/1B6 w - - 0 1")},
		{m: &Move{s1: B2, s2: B8}, pos: unsafeFEN("8/1p5b/4N3/4p3/8/8/1Q6/1B6 w - - 0 1")},
		{m: &Move{s1: B2, s2: G7}, pos: unsafeFEN("8/1p5b/4N3/4p3/8/8/1Q6/1B6 w - - 0 1")},
		// king moves
		{m: &Move{s1: E4, s2: F3}, pos: unsafeFEN("5r2/8/8/8/4K3/8/8/8 w - - 0 1")},
		{m: &Move{s1: E4, s2: F4}, pos: unsafeFEN("5r2/8/8/8/4K3/8/8/8 w - - 0 1")},
		{m: &Move{s1: E4, s2: F5}, pos: unsafeFEN("5r2/8/8/8/4K3/8/8/8 w - - 0 1")},
		// castleing
		{m: &Move{s1: E1, s2: B1}, pos: unsafeFEN("r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1")},
		{m: &Move{s1: E8, s2: B8}, pos: unsafeFEN("r3k2r/8/8/8/8/8/8/R3K2R b KQkq - 0 1")},
		{m: &Move{s1: E1, s2: C1, tags: QueenSideCastle}, pos: unsafeFEN("r3k2r/8/8/8/8/8/8/R2QK2R w KQkq - 0 1")},
		{m: &Move{s1: E1, s2: C1, tags: QueenSideCastle}, pos: unsafeFEN("2r1k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1")},
		{m: &Move{s1: E1, s2: C1, tags: QueenSideCastle}, pos: unsafeFEN("3rk2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1")},
		{m: &Move{s1: E1, s2: G1, tags: KingSideCastle}, pos: unsafeFEN("r3k2r/8/8/8/8/8/8/R3K2R w Qkq - 0 1")},
		{m: &Move{s1: E1, s2: C1, tags: QueenSideCastle}, pos: unsafeFEN("r3k2r/8/8/8/8/8/8/R3K2R w Kkq - 0 1")},
		{m: &Move{s1: E1, s2: A1}, pos: unsafeFEN("r3k2r/8/8/8/8/8/8/R2QK2R w KQkq - 0 1")},
		{m: &Move{s1: E1, s2: A1}, pos: unsafeFEN("2r1k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1")},
		{m: &Move{s1: E1, s2: A1}, pos: unsafeFEN("3rk2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1")},
		{m: &Move{s1: E1, s2: H1}, pos: unsafeFEN("r3k2r/8/8/8/8/8/8/R3K2R w Qkq - 0 1")},
		{m: &Move{s1: E1, s2: A1}, pos: unsafeFEN("r3k2r/8/8/8/8/8/8/R3K2R w Kkq - 0 1")},
		// invalid promotion for non-pawn move
		{m: &Move{s1: B8, s2: D7, promo: Pawn}, pos: unsafeFEN("rn1qkb1r/pp3ppp/2p1pn2/3p4/2PP4/2NQPN2/PP3PPP/R1B1K2R b KQkq - 0 7")},
		// en passant on doubled pawn file http://en.lichess.org/TnRtrHxf#24
		{m: &Move{s1: E3, s2: F6}, pos: unsafeFEN("r1b2rk1/pp2b1pp/1qn1p3/3pPp2/1P1P4/P2BPN2/6PP/RN1Q1RK1 w - f6 0 13")},
		// can't move piece out of pin (even if checking enemy king) http://en.lichess.org/JCRBhXH7#62
		{m: &Move{s1: E1, s2: E7}, pos: unsafeFEN("4R3/1r1k2pp/p1p5/1pP5/8/8/1PP3PP/2K1Rr2 w - - 5 32")},
		// invalid one up pawn capture
		{m: &Move{s1: E6, s2: E5}, pos: unsafeFEN(`2b1r3/2k2p1B/p2np3/4B3/8/5N2/PP1K1PPP/3R4 b - - 2 1`)},
		// invalid two up pawn capture
		{m: &Move{s1: H7, s2: H5}, pos: unsafeFEN(`2bqkbnr/rpppp2p/2n2p2/p5pB/5P2/4P3/PPPP2PP/RNBQK1NR b KQk - 4 6`)},
		// invalid pawn move d5e4
		{m: &Move{s1: D5, s2: E4}, pos: unsafeFEN(`rnbqkbnr/pp2pppp/8/2pp4/3P4/4PN2/PPP2PPP/RNBQKB1R b KQkq - 0 3`)},
	}

	positionUpdates = []moveTest{
		{
			m:       &Move{s1: E2, s2: E4},
			pos:     unsafeFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"),
			postPos: unsafeFEN("rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1"),
		},
		{
			m:       &Move{s1: E1, s2: G1, tags: KingSideCastle},
			pos:     unsafeFEN("r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1"),
			postPos: unsafeFEN("r3k2r/8/8/8/8/8/8/R4RK1 b kq - 1 1"),
		},
		{
			m:       &Move{s1: A4, s2: B3, tags: EnPassant},
			pos:     unsafeFEN("2r3k1/1q1nbppp/r3p3/3pP3/pPpP4/P1Q2N2/2RN1PPP/2R4K b - b3 0 23"),
			postPos: unsafeFEN("2r3k1/1q1nbppp/r3p3/3pP3/11pP4/PpQ2N2/2RN1PPP/2R4K w - - 0 24"),
		},
		{
			m:       &Move{s1: E1, s2: G1, tags: KingSideCastle},
			pos:     unsafeFEN("r2qk2r/pp1n1ppp/2pbpn2/3p4/2PP4/1PNQPN2/P4PPP/R1B1K2R w KQkq - 1 9"),
			postPos: unsafeFEN("r2qk2r/pp1n1ppp/2pbpn2/3p4/2PP4/1PNQPN2/P4PPP/R1B2RK1 b kq - 2 9"),
		},
	}
)

func unsafeFEN(s string) *Position {
	pos, err := decodeFEN(s)
	if err != nil {
		log.Fatal(err)
	}
	return pos
}

func TestValidMoves(t *testing.T) {
	for _, mt := range validMoves {
		if !moveIsValid(mt.pos, mt.m, false) {
			log.Println(mt.pos.String())
			log.Println(mt.pos.board.Draw())
			log.Println(mt.pos.ValidMoves())
			log.Println("In Check:", squaresAreAttacked(mt.pos, mt.pos.board.whiteKingSq))
			// log.Println("In Check:", mt.pos.inCheck())
			mt.pos.turn = mt.pos.turn.Other()
			t.Fatalf("expected move %s to be valid", mt.m)
		}
	}
}

func TestInvalidMoves(t *testing.T) {
	for _, mt := range invalidMoves {
		if moveIsValid(mt.pos, mt.m, false) {
			log.Println(mt.pos.String())
			log.Println(mt.pos.board.Draw())
			t.Fatalf("expected move %s to be invalid", mt.m)
		}
	}
}

func TestPositionUpdates(t *testing.T) {
	for _, mt := range positionUpdates {
		if !moveIsValid(mt.pos, mt.m, true) {
			log.Println(mt.pos.String())
			log.Println(mt.pos.board.Draw())
			log.Println(mt.pos.ValidMoves())
			t.Fatalf("expected move %s %v to be valid", mt.m, mt.m.tags)
		}

		postPos := mt.pos.Update(mt.m)
		if postPos.String() != mt.postPos.String() {
			t.Fatalf("starting from board \n%s%s\n after move %s\n expected board to be %s\n%s\n but was %s\n%s\n",
				mt.pos.String(),
				mt.pos.board.Draw(),
				mt.m.String(),
				mt.postPos.String(),
				mt.postPos.board.Draw(),
				postPos.String(),
				postPos.board.Draw(),
			)
		}
	}
}

type perfTest struct {
	pos           *Position
	nodesPerDepth []int
}

/* https://www.chessprogramming.org/Perft_Results */
var perfResults = []perfTest{
	{pos: unsafeFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"), nodesPerDepth: []int{
		20, 400, 8902, 197281,
		// 4865609, 119060324, 3195901860, 84998978956, 2439530234167, 69352859712417
	}},
	{pos: unsafeFEN("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1"), nodesPerDepth: []int{
		48, 2039, 97862,
		// 4085603, 193690690
	}},
	{pos: unsafeFEN("8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1"), nodesPerDepth: []int{
		14, 191, 2812, 43238, 674624,
		// 11030083, 178633661
	}},
	{pos: unsafeFEN("r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1"), nodesPerDepth: []int{
		6, 264, 9467, 422333,
		// 15833292, 706045033
	}},
	{pos: unsafeFEN("r2q1rk1/pP1p2pp/Q4n2/bbp1p3/Np6/1B3NBn/pPPP1PPP/R3K2R b KQ - 0 1"), nodesPerDepth: []int{
		6, 264, 9467, 422333,
		// 15833292, 706045033
	}},
	{pos: unsafeFEN("rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8"), nodesPerDepth: []int{
		44, 1486, 62379,
		// 2103487, 89941194
	}},
	{pos: unsafeFEN("r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10"), nodesPerDepth: []int{
		46, 2079, 89890,
		// 3894594, 164075551, 6923051137, 287188994746, 11923589843526, 490154852788714
	}},
	// FRC/960 specific positions.
	{pos: unsafeFEN("bqrkrbnn/ppp1pppp/8/8/8/8/PPP1PPPP/BQRKRBNN w CKeq - 0 1"), nodesPerDepth: []int{
		19, 342, 6987, 142308,
		// 3294156, 75460468,
	}},
	{pos: unsafeFEN("qn1rkrbb/pp1p1ppp/2p1p3/3n4/4P2P/2NP4/PPP2PP1/Q1NRKRBB w FDfd - 1 9"), nodesPerDepth: []int{
		24, 585, 14769, 356950,
		// 9482310, 233468620
	}},
	{pos: unsafeFEN("nbnrbk2/p1pppp1p/1p3qr1/6p1/1B1P4/1N6/PPP1PPPP/1BNR1RKQ w d - 2 9"), nodesPerDepth: []int{
		30, 796, 22780, 687302,
		// 20120565, 641832725
	}},
	{pos: unsafeFEN("nrbbnk1r/pp2pppq/8/2pp3p/3P2P1/1N6/PPP1PP1P/1RBBNKQR w HBhb - 0 9"), nodesPerDepth: []int{
		29, 1036, 31344,
		// 1139166, 35627310, 1310683359
	}},
	{pos: unsafeFEN("2rkqbnr/p1pppppp/2b5/1pn5/1P3P1Q/2B5/P1PPP1PP/1NRK1BNR w HChc - 3 9"), nodesPerDepth: []int{
		33, 904, 30111,
		// 840025, 28194726, 801757709
	}},
	{pos: unsafeFEN("1qrkrnbb/1p1p1ppp/pnp1p3/8/3PP3/P6P/1PP2PP1/NQRKRNBB w ECec - 0 9"), nodesPerDepth: []int{
		24, 688, 17342, 511444,
		// 13322502, 403441498
	}},
	{pos: unsafeFEN("qbrnnkbr/1p2pp1p/p1p3p1/3p4/6P1/P1N4P/1PPPPP2/QBR1NKBR w HChc - 0 9"), nodesPerDepth: []int{
		26, 701, 18930, 521377,
		// 14733245, 416881799
	}},
	{pos: unsafeFEN("r1nbbkrq/1ppp2pp/2n2p2/p3p3/5P2/1N4BP/PPPPP1P1/RN1B1KRQ w GAga - 0 9"), nodesPerDepth: []int{
		25, 774, 20141, 618805,
		// 16718577, 515864053
	}},
	{pos: unsafeFEN("rnb2bqr/ppkpppp1/3n3p/2p5/6PP/2N2P2/PPPPP3/R1BKNBQR w HA - 2 9"), nodesPerDepth: []int{
		30, 647, 20365, 467780,
		// 15115531, 369257622
	}},
	{pos: unsafeFEN("b1rkqrnb/p1ppp1pp/1p1n4/5p2/5P2/PN5P/1PPPP1P1/BR1KQRNB w FBf - 0 9"), nodesPerDepth: []int{
		23, 688, 17259, 531592,
		// 14228372, 451842354
	}},
	{pos: unsafeFEN("bbrkqn1r/1pppppp1/5n2/p7/1PP2P1p/7N/P2PP1PP/BBRKQN1R w HChc - 1 9"), nodesPerDepth: []int{
		36, 963, 35291,
		// 973839, 35907489, 1034223364
	}},
	{pos: unsafeFEN("1rkb1rbn/p1pp1ppp/3np3/1p6/4qP2/3NB3/PPPPPRPP/QRKB3N w Bfb - 0 9"), nodesPerDepth: []int{
		22, 923, 22585, 914106,
		// 24049880, 957218571
	}},
	{pos: unsafeFEN("rknr1bnq/p2pp1pp/1p3p2/2p4b/6PP/2P2N2/PP1PPP2/RKNRBB1Q w DAda - 1 9"), nodesPerDepth: []int{
		25, 502, 13150, 279098,
		// 7824941, 175766730
	}},
	{pos: unsafeFEN("1kbrnqnb/r1ppppp1/8/pp5p/8/1P1NP3/P1PP1PPP/RKB1RQNB w Ad - 2 9"), nodesPerDepth: []int{
		26, 618, 17305, 442643,
		// 13112297, 357030697
	}},
	{pos: unsafeFEN("r1bkrn1q/ppbppppp/5n2/2p5/3P4/P6N/1PP1PPPP/RBBKRNQ1 w KQkq - 3 99"), nodesPerDepth: []int{
		27, 822, 22551, 678880,
		// 19115128, 578210135
	}},
	{pos: unsafeFEN("qn1rkrbb/pp1p1ppp/2p1p3/3n4/4P2P/2NP4/PPP2PP1/Q1NRKRBB w KQkq - 1 9"), nodesPerDepth: []int{
		24, 585, 14769, 356950,
		// 9482310, 233468620
	}},
	{pos: unsafeFEN("qnr1bkrb/pppp2pp/3np3/5p2/8/P2P2P1/NPP1PP1P/QN1RBKRB w KQk - 3 9"), nodesPerDepth: []int{
		33, 823, 26895, 713420,
		// 23114629, 646390782
	}},
	{pos: unsafeFEN("bnrbk1qn/1pppprpp/8/p4p1P/6P1/3P4/PPP1PP2/BNRBKRQN w KQq - 0 9"), nodesPerDepth: []int{
		22, 459, 11447, 268157,
		// 7371098, 190583454
	}},
	{pos: unsafeFEN("1rbnkbrq/pppppp2/n5pp/2P5/P7/4N3/1P1PPPPP/RNB1KBRQ w KQk - 2 9"), nodesPerDepth: []int{
		23, 574, 14146, 391413,
		// 10203438, 301874034
	}},
	{pos: unsafeFEN("rbbk1rnq/pppp1pp1/4p2p/8/3P2n1/4BN1P/PPP1PPP1/RB1K1RNQ w KQkq - 3 9"), nodesPerDepth: []int{
		26, 628, 16151, 411995,
		// 11237919, 300314373
	}},
}

func TestPerfResults(t *testing.T) {
	for _, perf := range perfResults {
		countMoves(t, perf.pos, []*Position{perf.pos}, perf.nodesPerDepth, len(perf.nodesPerDepth))
	}
}

func countMoves(t *testing.T, originalPosition *Position, positions []*Position, nodesPerDepth []int, maxDepth int) {
	if len(nodesPerDepth) == 0 {
		return
	}
	depth := maxDepth - len(nodesPerDepth) + 1
	expNodes := nodesPerDepth[0]
	newPositions := make([]*Position, 0)
	for _, pos := range positions {
		for _, move := range pos.ValidMoves() {
			newPos := pos.Update(move)
			newPositions = append(newPositions, newPos)
		}
	}
	gotNodes := len(newPositions)
	if expNodes != gotNodes {
		t.Errorf("Depth: %d Expected: %d Got: %d", depth, expNodes, gotNodes)
		t.Log("##############################")
		t.Log("# Original position info")
		t.Log("###")
		t.Log(originalPosition.String())
		t.Log(originalPosition.board.Draw())
		t.Log("##############################")
		t.Log("# Details in JSONL (http://jsonlines.org)")
		t.Log("###")
		for _, pos := range positions {
			t.Logf(`{"position": "%s", "moves": %d}`, pos.String(), len(pos.ValidMoves()))
		}
	}
	countMoves(t, originalPosition, newPositions, nodesPerDepth[1:], maxDepth)
}

func BenchmarkValidMoves(b *testing.B) {
	pos := unsafeFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		pos.ValidMoves()
		pos.validMoves = nil
	}
}

func moveIsValid(pos *Position, m *Move, useTags bool) bool {
	for _, move := range pos.ValidMoves() {
		// Support comparison of legacy and 960 style castling move formats.
		if move.HasTag(QueenSideCastle) || move.HasTag(KingSideCastle) {
			castle_flags := QueenSideCastle | KingSideCastle
			if (move.tags&castle_flags) == (m.tags&castle_flags) && move.s1 == m.s1 {
				if m.tags != move.tags {
					return false
				}
				return true
			}
		}
		if move.s1 == m.s1 && move.s2 == m.s2 && move.promo == m.promo {
			if useTags {
				if m.tags != move.tags {
					return false
				}
			}
			return true
		}
	}
	return false
}
