package chess

import (
	"strings"
	"unicode"
)

type engine struct{}

func (engine) CalcMoves(pos *Position, first bool) []*Move {
	// generate possible moves
	moves := standardMoves(pos, first)
	// return moves including castles
	return append(moves, castleMoves(pos)...)
}

func (engine) Status(pos *Position) Method {
	hasMove := false
	if pos.validMoves != nil {
		hasMove = len(pos.validMoves) > 0
	} else {
		hasMove = len(engine{}.CalcMoves(pos, true)) > 0
	}
	if !pos.inCheck && !hasMove {
		return Stalemate
	} else if pos.inCheck && !hasMove {
		return Checkmate
	}
	return NoMethod
}

var (
	promoPieceTypes = []PieceType{Queen, Rook, Bishop, Knight}
)

func standardMoves(pos *Position, first bool) []*Move {
	// compute allowed destination bitboard
	bbAllowed := ^pos.board.whiteSqs
	if pos.Turn() == Black {
		bbAllowed = ^pos.board.blackSqs
	}
	moves := []*Move{}
	// iterate through pieces to find possible moves
	for _, p := range allPieces {
		if pos.Turn() != p.Color() {
			continue
		}
		// iterate through possible starting squares for piece
		s1BB := pos.board.bbForPiece(p)
		if s1BB == 0 {
			continue
		}
		for s1 := 0; s1 < numOfSquaresInBoard; s1++ {
			if s1BB&bbForSquare(Square(s1)) == 0 {
				continue
			}
			// iterate through possible destination squares for piece
			s2BB := bbForPossibleMoves(pos, p.Type(), Square(s1)) & bbAllowed
			if s2BB == 0 {
				continue
			}
			for s2 := 0; s2 < numOfSquaresInBoard; s2++ {
				if s2BB&bbForSquare(Square(s2)) == 0 {
					continue
				}
				// add promotions if pawn on promo square
				if (p == WhitePawn && Square(s2).Rank() == Rank8) || (p == BlackPawn && Square(s2).Rank() == Rank1) {
					for _, pt := range promoPieceTypes {
						m := &Move{s1: Square(s1), s2: Square(s2), promo: pt}
						addTags(m, pos)
						// filter out moves that put king into check
						if !m.HasTag(inCheck) {
							moves = append(moves, m)
							if first {
								return moves
							}
						}
					}
				} else {
					m := &Move{s1: Square(s1), s2: Square(s2)}
					addTags(m, pos)
					// filter out moves that put king into check
					if !m.HasTag(inCheck) {
						moves = append(moves, m)
						if first {
							return moves
						}
					}
				}
			}
		}
	}
	return moves
}

func addTags(m *Move, pos *Position) {
	p := pos.board.Piece(m.s1)
	if pos.board.isOccupied(m.s2) && !m.HasTag(QueenSideCastle) && !m.HasTag(KingSideCastle) {
		m.addTag(Capture)
	} else if m.s2 == pos.enPassantSquare && p.Type() == Pawn {
		m.addTag(EnPassant)
	}
	// determine if in check after move (makes move invalid)
	cp := pos.copy()
	cp.board.update(m)
	if isInCheck(cp) {
		m.addTag(inCheck)
	}
	// determine if opponent in check after move
	cp.turn = cp.turn.Other()
	if isInCheck(cp) {
		m.addTag(Check)
	}
}

func isInCheck(pos *Position) bool {
	kingSq := pos.board.whiteKingSq
	if pos.Turn() == Black {
		kingSq = pos.board.blackKingSq
	}
	// king should only be missing in tests / examples
	if kingSq == NoSquare {
		return false
	}
	return squaresAreAttacked(pos, kingSq)
}

func squaresAreAttacked(pos *Position, sqs ...Square) bool {
	otherColor := pos.Turn().Other()
	occ := ^pos.board.emptySqs
	for _, sq := range sqs {
		// hot path check to see if attack vector is possible
		s2BB := pos.board.blackSqs
		if pos.Turn() == Black {
			s2BB = pos.board.whiteSqs
		}
		if ((diaAttack(occ, sq)|hvAttack(occ, sq))&s2BB)|(bbKnightMoves[sq]&s2BB) == 0 {
			continue
		}
		// check queen attack vector
		queenBB := pos.board.bbForPiece(getPiece(Queen, otherColor))
		bb := (diaAttack(occ, sq) | hvAttack(occ, sq)) & queenBB
		if bb != 0 {
			return true
		}
		// check rook attack vector
		rookBB := pos.board.bbForPiece(getPiece(Rook, otherColor))
		bb = hvAttack(occ, sq) & rookBB
		if bb != 0 {
			return true
		}
		// check bishop attack vector
		bishopBB := pos.board.bbForPiece(getPiece(Bishop, otherColor))
		bb = diaAttack(occ, sq) & bishopBB
		if bb != 0 {
			return true
		}
		// check knight attack vector
		knightBB := pos.board.bbForPiece(getPiece(Knight, otherColor))
		bb = bbKnightMoves[sq] & knightBB
		if bb != 0 {
			return true
		}
		// check pawn attack vector
		if pos.Turn() == White {
			capRight := (pos.board.bbBlackPawn & ^bbFileH & ^bbRank1) << 7
			capLeft := (pos.board.bbBlackPawn & ^bbFileA & ^bbRank1) << 9
			bb = (capRight | capLeft) & bbForSquare(sq)
			if bb != 0 {
				return true
			}
		} else {
			capRight := (pos.board.bbWhitePawn & ^bbFileH & ^bbRank8) >> 9
			capLeft := (pos.board.bbWhitePawn & ^bbFileA & ^bbRank8) >> 7
			bb = (capRight | capLeft) & bbForSquare(sq)
			if bb != 0 {
				return true
			}
		}
		// check king attack vector
		kingBB := pos.board.bbForPiece(getPiece(King, otherColor))
		bb = bbKingMoves[sq] & kingBB
		if bb != 0 {
			return true
		}
	}
	return false
}

func bbForPossibleMoves(pos *Position, pt PieceType, sq Square) bitboard {
	switch pt {
	case King:
		return bbKingMoves[sq]
	case Queen:
		return diaAttack(^pos.board.emptySqs, sq) | hvAttack(^pos.board.emptySqs, sq)
	case Rook:
		return hvAttack(^pos.board.emptySqs, sq)
	case Bishop:
		return diaAttack(^pos.board.emptySqs, sq)
	case Knight:
		return bbKnightMoves[sq]
	case Pawn:
		return pawnMoves(pos, sq)
	}
	return bitboard(0)
}

func min(a, b File) File {
	if a < b {
		return a
	}
	return b
}

func max(a, b File) File {
	if a > b {
		return a
	}
	return b
}

// TODO can calc isInCheck twice
func castleMoves(pos *Position) []*Move {
	moves := []*Move{}
	kingSide := pos.castleRights.CanCastle(pos.Board(), pos.Turn(), KingSide)
	queenSide := pos.castleRights.CanCastle(pos.Board(), pos.Turn(), QueenSide)
	if !kingSide && !queenSide || pos.inCheck {
		return moves
	}
	kingFile := pos.board.whiteKingSq.File()
	kingRank := Rank1
	if pos.Turn() != White {
		kingFile = pos.board.blackKingSq.File()
		kingRank = Rank8
	}
	for _, r := range string(pos.castleRights) {
		if (pos.Turn() == White) != unicode.IsUpper(r) {
			continue
		}
		rookFile := FileA
		if pos.turn == White {
			if string(r) == "K" {
				for file := FileH; file > kingFile; file -= 1 {
					if pos.board.bbWhiteRook.Occupied(getSquare(file, kingRank)) {
						rookFile = file
						break
					}
				}
			} else if string(r) == "Q" {
				for file := FileA; file < kingFile; file += 1 {
					if pos.board.bbWhiteRook.Occupied(getSquare(file, kingRank)) {
						rookFile = file
						break
					}
				}
			} else {
				rookFile = File(int(FileA) + strings.Index(fileChars, strings.ToLower(string(r))))
			}
		} else {
			if string(r) == "k" {
				for file := FileH; file > kingFile; file -= 1 {
					if pos.board.bbBlackRook.Occupied(getSquare(file, kingRank)) {
						rookFile = file
						break
					}
				}
			} else if string(r) == "q" {
				for file := FileA; file < kingFile; file += 1 {
					if pos.board.bbBlackRook.Occupied(getSquare(file, kingRank)) {
						rookFile = file
						break
					}
				}
			} else {
				rookFile = File(int(FileA) + strings.Index(fileChars, string(r)))
			}
		}
		if rookFile < kingFile {
			// Validate Queen side legal.
			ocupied := ^pos.board.emptySqs
			count := 0
			lower := min(rookFile, FileC)
			higher := max(kingFile, FileD)
			for file := lower; file <= higher; file += 1 {
				if ocupied&bbForSquare(getSquare(file, kingRank)) != 0 {
					count++
				}
			}
			if count > 2 {
				continue
			}
			legal := true
			lower = min(kingFile, FileC)
			higher = max(kingFile, FileC)
			for file := lower; file <= higher; file += 1 {
				if file == kingFile {
					continue
				}
				if squaresAreAttacked(pos, getSquare(file, kingRank)) {
					legal = false
					break
				}
			}
			if !legal {
				continue
			}
			m := &Move{s1: getSquare(kingFile, kingRank), s2: getSquare(rookFile, kingRank)}
			m.addTag(QueenSideCastle)
			addTags(m, pos)
			// filter out moves that put king into check
			if !m.HasTag(inCheck) {
				moves = append(moves, m)
			}
		} else {
			// Validate King side legal.
			ocupied := ^pos.board.emptySqs
			count := 0
			lower := min(kingFile, FileF)
			higher := max(rookFile, FileG)
			for file := lower; file <= higher; file += 1 {
				if ocupied&bbForSquare(getSquare(file, kingRank)) != 0 {
					count++
				}
			}
			if count > 2 {
				continue
			}
			legal := true
			lower = min(kingFile, FileG)
			higher = max(kingFile, FileG)
			for file := lower; file <= higher; file += 1 {
				if file == kingFile {
					continue
				}
				if squaresAreAttacked(pos, getSquare(file, kingRank)) {
					legal = false
					break
				}
			}
			if !legal {
				continue
			}
			m := &Move{s1: getSquare(kingFile, kingRank), s2: getSquare(rookFile, kingRank)}
			m.addTag(KingSideCastle)
			addTags(m, pos)
			// filter out moves that put king into check
			if !m.HasTag(inCheck) {
				moves = append(moves, m)
			}
		}
	}
	return moves
}

func pawnMoves(pos *Position, sq Square) bitboard {
	bb := bbForSquare(sq)
	var bbEnPassant bitboard
	if pos.enPassantSquare != NoSquare {
		bbEnPassant = bbForSquare(pos.enPassantSquare)
	}
	if pos.Turn() == White {
		capRight := ((bb & ^bbFileH & ^bbRank8) >> 9) & (pos.board.blackSqs | bbEnPassant)
		capLeft := ((bb & ^bbFileA & ^bbRank8) >> 7) & (pos.board.blackSqs | bbEnPassant)
		upOne := ((bb & ^bbRank8) >> 8) & pos.board.emptySqs
		upTwo := ((upOne & bbRank3) >> 8) & pos.board.emptySqs
		return capRight | capLeft | upOne | upTwo
	}
	capRight := ((bb & ^bbFileH & ^bbRank1) << 7) & (pos.board.whiteSqs | bbEnPassant)
	capLeft := ((bb & ^bbFileA & ^bbRank1) << 9) & (pos.board.whiteSqs | bbEnPassant)
	upOne := ((bb & ^bbRank1) << 8) & pos.board.emptySqs
	upTwo := ((upOne & bbRank6) << 8) & pos.board.emptySqs
	return capRight | capLeft | upOne | upTwo
}

func diaAttack(occupied bitboard, sq Square) bitboard {
	pos := bbForSquare(sq)
	dMask := bbDiagonals[sq]
	adMask := bbAntiDiagonals[sq]
	return linearAttack(occupied, pos, dMask) | linearAttack(occupied, pos, adMask)
}

func hvAttack(occupied bitboard, sq Square) bitboard {
	pos := bbForSquare(sq)
	rankMask := bbRanks[Square(sq).Rank()]
	fileMask := bbFiles[Square(sq).File()]
	return linearAttack(occupied, pos, rankMask) | linearAttack(occupied, pos, fileMask)
}

func linearAttack(occupied, pos, mask bitboard) bitboard {
	oInMask := occupied & mask
	return ((oInMask - 2*pos) ^ (oInMask.Reverse() - 2*pos.Reverse()).Reverse()) & mask
}

const (
	bbFileA bitboard = 9259542123273814144
	bbFileB bitboard = 4629771061636907072
	bbFileC bitboard = 2314885530818453536
	bbFileD bitboard = 1157442765409226768
	bbFileE bitboard = 578721382704613384
	bbFileF bitboard = 289360691352306692
	bbFileG bitboard = 144680345676153346
	bbFileH bitboard = 72340172838076673

	bbRank1 bitboard = 18374686479671623680
	bbRank2 bitboard = 71776119061217280
	bbRank3 bitboard = 280375465082880
	bbRank4 bitboard = 1095216660480
	bbRank5 bitboard = 4278190080
	bbRank6 bitboard = 16711680
	bbRank7 bitboard = 65280
	bbRank8 bitboard = 255
)

// TODO make method on Square
func bbForSquare(sq Square) bitboard {
	return bbSquares[sq]
}

var (
	bbFiles = [8]bitboard{bbFileA, bbFileB, bbFileC, bbFileD, bbFileE, bbFileF, bbFileG, bbFileH}
	bbRanks = [8]bitboard{bbRank1, bbRank2, bbRank3, bbRank4, bbRank5, bbRank6, bbRank7, bbRank8}

	bbDiagonals = [64]bitboard{9241421688590303745, 4620710844295151872, 2310355422147575808, 1155177711073755136, 577588855528488960, 288794425616760832, 144396663052566528, 72057594037927936, 36099303471055874, 9241421688590303745, 4620710844295151872, 2310355422147575808, 1155177711073755136, 577588855528488960, 288794425616760832, 144396663052566528, 141012904183812, 36099303471055874, 9241421688590303745, 4620710844295151872, 2310355422147575808, 1155177711073755136, 577588855528488960, 288794425616760832, 550831656968, 141012904183812, 36099303471055874, 9241421688590303745, 4620710844295151872, 2310355422147575808, 1155177711073755136, 577588855528488960, 2151686160, 550831656968, 141012904183812, 36099303471055874, 9241421688590303745, 4620710844295151872, 2310355422147575808, 1155177711073755136, 8405024, 2151686160, 550831656968, 141012904183812, 36099303471055874, 9241421688590303745, 4620710844295151872, 2310355422147575808, 32832, 8405024, 2151686160, 550831656968, 141012904183812, 36099303471055874, 9241421688590303745, 4620710844295151872, 128, 32832, 8405024, 2151686160, 550831656968, 141012904183812, 36099303471055874, 9241421688590303745}

	bbAntiDiagonals = [64]bitboard{9223372036854775808, 4647714815446351872, 2323998145211531264, 1161999622361579520, 580999813328273408, 290499906672525312, 145249953336295424, 72624976668147840, 4647714815446351872, 2323998145211531264, 1161999622361579520, 580999813328273408, 290499906672525312, 145249953336295424, 72624976668147840, 283691315109952, 2323998145211531264, 1161999622361579520, 580999813328273408, 290499906672525312, 145249953336295424, 72624976668147840, 283691315109952, 1108169199648, 1161999622361579520, 580999813328273408, 290499906672525312, 145249953336295424, 72624976668147840, 283691315109952, 1108169199648, 4328785936, 580999813328273408, 290499906672525312, 145249953336295424, 72624976668147840, 283691315109952, 1108169199648, 4328785936, 16909320, 290499906672525312, 145249953336295424, 72624976668147840, 283691315109952, 1108169199648, 4328785936, 16909320, 66052, 145249953336295424, 72624976668147840, 283691315109952, 1108169199648, 4328785936, 16909320, 66052, 258, 72624976668147840, 283691315109952, 1108169199648, 4328785936, 16909320, 66052, 258, 1}

	bbKnightMoves = [64]bitboard{9077567998918656, 4679521487814656, 38368557762871296, 19184278881435648, 9592139440717824, 4796069720358912, 2257297371824128, 1128098930098176, 2305878468463689728, 1152939783987658752, 9799982666336960512, 4899991333168480256, 2449995666584240128, 1224997833292120064, 576469569871282176, 288234782788157440, 4620693356194824192, 11533718717099671552, 5802888705324613632, 2901444352662306816, 1450722176331153408, 725361088165576704, 362539804446949376, 145241105196122112, 18049583422636032, 45053588738670592, 22667534005174272, 11333767002587136, 5666883501293568, 2833441750646784, 1416171111120896, 567348067172352, 70506185244672, 175990581010432, 88545054707712, 44272527353856, 22136263676928, 11068131838464, 5531918402816, 2216203387392, 275414786112, 687463207072, 345879119952, 172939559976, 86469779988, 43234889994, 21609056261, 8657044482, 1075839008, 2685403152, 1351090312, 675545156, 337772578, 168886289, 84410376, 33816580, 4202496, 10489856, 5277696, 2638848, 1319424, 659712, 329728, 132096}

	bbBishopMoves = [64]bitboard{18049651735527937, 45053622886727936, 22667548931719168, 11334324221640704, 5667164249915392, 2833579985862656, 1416240237150208, 567382630219904, 4611756524879479810, 11529391036782871041, 5764696068147249408, 2882348036221108224, 1441174018118909952, 720587009051099136, 360293502378066048, 144117404414255168, 2323857683139004420, 1197958188344280066, 9822351133174399489, 4911175566595588352, 2455587783297826816, 1227793891648880768, 577868148797087808, 288793334762704928, 1161999073681608712, 581140276476643332, 326598935265674242, 9386671504487645697, 4693335752243822976, 2310639079102947392, 1155178802063085600, 577588851267340304, 580999811184992272, 290500455356698632, 145390965166737412, 108724279602332802, 9241705379636978241, 4620711952330133792, 2310355426409252880, 1155177711057110024, 290499906664153120, 145249955479592976, 72625527495610504, 424704217196612, 36100411639206946, 9241421692918565393, 4620710844311799048, 2310355422147510788, 145249953336262720, 72624976676520096, 283693466779728, 1659000848424, 141017232965652, 36099303487963146, 9241421688590368773, 4620710844295151618, 72624976668147712, 283691315142656, 1108177604608, 6480472064, 550848566272, 141012904249856, 36099303471056128, 9241421688590303744}

	bbRookMoves = [64]bitboard{9187484529235886208, 13781085504453754944, 16077885992062689312, 17226286235867156496, 17800486357769390088, 18087586418720506884, 18231136449196065282, 18302911464433844481, 9259260648297103488, 4665518383679160384, 2368647251370188832, 1220211685215703056, 645993902138460168, 358885010599838724, 215330564830528002, 143553341945872641, 9259541023762186368, 4629910699613634624, 2315095537539358752, 1157687956502220816, 578984165983651848, 289632270724367364, 144956323094725122, 72618349279904001, 9259542118978846848, 4629771607097753664, 2314886351157207072, 1157443723186933776, 578722409201797128, 289361752209228804, 144681423712944642, 72341259464802561, 9259542123257036928, 4629771063767613504, 2314885534022901792, 1157442769150545936, 578721386714368008, 289360695496279044, 144680349887234562, 72340177082712321, 9259542123273748608, 4629771061645230144, 2314885530830970912, 1157442765423841296, 578721382720276488, 289360691368494084, 144680345692602882, 72340172854657281, 9259542123273813888, 4629771061636939584, 2314885530818502432, 1157442765409283856, 578721382704674568, 289360691352369924, 144680345676217602, 72340172838141441, 9259542123273814143, 4629771061636907199, 2314885530818453727, 1157442765409226991, 578721382704613623, 289360691352306939, 144680345676153597, 72340172838076926}

	bbQueenMoves = [64]bitboard{9205534180971414145, 13826139127340482880, 16100553540994408480, 17237620560088797200, 17806153522019305480, 18090419998706369540, 18232552689433215490, 18303478847064064385, 13871017173176583298, 16194909420462031425, 8133343319517438240, 4102559721436811280, 2087167920257370120, 1079472019650937860, 575624067208594050, 287670746360127809, 11583398706901190788, 5827868887957914690, 12137446670713758241, 6068863523097809168, 3034571949281478664, 1517426162373248132, 722824471891812930, 361411684042608929, 10421541192660455560, 5210911883574396996, 2641485286422881314, 10544115227674579473, 5272058161445620104, 2600000831312176196, 1299860225776030242, 649930110732142865, 9840541934442029200, 4920271519124312136, 2460276499189639204, 1266167048752878738, 9820426766351346249, 4910072647826412836, 2455035776296487442, 1227517888139822345, 9550042029937901728, 4775021017124823120, 2387511058326581416, 1157867469641037908, 614821794359483434, 9530782384287059477, 4765391190004401930, 2382695595002168069, 9404792076610076608, 4702396038313459680, 2315169224285282160, 1157444424410132280, 578862399937640220, 325459994840333070, 9386102034266586375, 4693051017133293059, 9332167099941961855, 4630054752952049855, 2314886638996058335, 1157442771889699055, 578721933553179895, 289501704256556795, 180779649147209725, 9313761861428380670}

	bbKingMoves = [64]bitboard{4665729213955833856, 11592265440851656704, 5796132720425828352, 2898066360212914176, 1449033180106457088, 724516590053228544, 362258295026614272, 144959613005987840, 13853283560024178688, 16186183351374184448, 8093091675687092224, 4046545837843546112, 2023272918921773056, 1011636459460886528, 505818229730443264, 216739030602088448, 54114388906344448, 63227278716305408, 31613639358152704, 15806819679076352, 7903409839538176, 3951704919769088, 1975852459884544, 846636838289408, 211384331665408, 246981557485568, 123490778742784, 61745389371392, 30872694685696, 15436347342848, 7718173671424, 3307175149568, 825720045568, 964771708928, 482385854464, 241192927232, 120596463616, 60298231808, 30149115904, 12918652928, 3225468928, 3768639488, 1884319744, 942159872, 471079936, 235539968, 117769984, 50463488, 12599488, 14721248, 7360624, 3680312, 1840156, 920078, 460039, 197123, 49216, 57504, 28752, 14376, 7188, 3594, 1797, 770}

	bbSquares = [64]bitboard{}
)

func init() {
	for sq := 0; sq < 64; sq++ {
		bbSquares[sq] = bitboard(uint64(1) << (uint8(63) - uint8(sq)))
	}
}
