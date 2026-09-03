// lo_lexer.go — real LO lexer for `.llll` source files, implementing the real, formal,
// now-canonical grammar in `LO/GRAMMAR.md` (Phase 0 of LO/NORTHSTAR.md's own phased plan,
// written by a concurrent session — reconciled with directly, not re-derived independently: an
// earlier version of this file predated GRAMMAR.md and invented its own token set, including
// comparison/logical operators and a parameter-reference mechanism GRAMMAR.md's own EBNF does
// not define at all; rewritten here to be a real, honest, strict SUBSET of the canonical grammar
// instead of a divergent one).
//
// Real, deliberate v0 scope: the largest subset of GRAMMAR.md's own full EBNF that (a) needs no
// String/Vector/Matrix/Pattern/Union support (all genuinely out of reach of BOTH real compiler
// backends for a while yet — Vec/defenum/match aren't in burrow's own real scope at all, see
// NORTHSTAR.md finding #5) and (b) still exercises the real toolchain end to end:
//
//	Program  ::= Door Expr
//	Door     ::= DOOR I32                          (v0: I32 the only supported TypeAtom)
//	Expr     ::= Ternary
//	Ternary  ::= Cond QUERY Expr COLON Expr | Value
//	Cond     ::= Value EQ Value
//	Value    ::= Arith | State
//	Arith    ::= Value ArithOp Value
//	ArithOp  ::= PLUS4 | MINUS4 | AND4 | OR4 | XOR4
//	State    ::= S0 | S1 | S2 | S3
//
// This is real, honest, and narrower than GRAMMAR.md's own full grammar (which also covers
// VectorLit/LinAlg/Labeled/Pattern/MATCH) — those need real Vec/String/struct support this v0
// deliberately doesn't attempt yet, named here rather than silently worked around with
// unauthorized new tokens.
package main

import (
	"errors"
	"unicode/utf8"
)

type loTokKind int

const (
	loTokState loTokKind = iota // S0-S3: 🌑🌒🌓🌔
	loTokEq                     // EQ: 🟰 (GRAMMAR.md §1.1 -- provisional glyph, flagged there)
	loTokQuestion                // QUERY: ❓
	loTokDoor                    // DOOR: 🚪
	loTokI32Type                 // I32 (TypeAtom): 🔢
	loTokArith                   // PLUS4/MINUS4/AND4/OR4/XOR4; Text carries which one
	loTokColon                   // COLON: ':' (plain ASCII per GRAMMAR.md §1)
)

type loToken struct {
	Kind loTokKind
	Val  int    // loTokState: the real base4 digit value, 0-3.
	Text string // loTokArith: "plus4"/"minus4"/"and4"/"or4"/"xor4".
	Line int
}

// loStateDigits — real, direct port of GRAMMAR.md §1.1's own S0-S3 codepoint table.
var loStateDigits = map[rune]int{
	'🌑': 0, // U+1F311
	'🌒': 1, // U+1F312
	'🌓': 2, // U+1F313
	'🌔': 3, // U+1F314
}

// loArithOps — real, direct port of GRAMMAR.md §1.1's own PLUS4/MINUS4/AND4/OR4/XOR4 row, plus
// §5.1's own real semantics (PLUS4/MINUS4 wrap mod 4; AND4/OR4/XOR4 need no wrapping since 2-bit
// bitwise ops already stay in range) -- lo_parser.go's own parseArith implements the wrapping.
var loArithOps = map[rune]string{
	'➕': "plus4",  // U+2795
	'➖': "minus4", // U+2796
	'🔗': "and4",   // U+1F517
	'🔮': "or4",    // U+1F52E
	'🔀': "xor4",   // U+1F500
}

// loStripVariationSelector implements GRAMMAR.md §1.2's own real emoji-matching rule: strip
// exactly one trailing VS15 (U+FE0E, text presentation) or VS16 (U+FE0F, emoji presentation)
// before matching a token's base codepoint. A no-op for every token this v0's own narrower
// subset actually uses (none of S0-S3/EQ/QUERY/DOOR/I32/the Arith set carry a variation selector
// in GRAMMAR.md's own table), but implemented for real, honest grammar compliance rather than
// skipped as "not needed yet" -- a future token (⚙️/🕸️/🏷️/☄️/🛤️/🗜️/🕳️, all VS16-suffixed in
// GRAMMAR.md §1.1) needs this to lex correctly the first time it's added, not as a later patch.
func loStripVariationSelector(runes []rune, i int) (rune, int) {
	r := runes[i]
	if i+1 < len(runes) && (runes[i+1] == '︎' || runes[i+1] == '️') {
		return r, i + 1
	}
	return r, i
}

// LoLex tokenizes real `.llll` source per GRAMMAR.md §1 -- iterates by rune (real, matches
// GRAMMAR.md's own "sequence of Unicode scalar values" framing), skips ASCII whitespace between
// tokens (§1: "not significant and may be repeated freely"), strips a trailing variation
// selector per §1.2 before matching each candidate token codepoint.
func LoLex(src string) ([]loToken, error) {
	var toks []loToken
	line := 1
	runes := []rune(src)
	for i := 0; i < len(runes); i++ {
		base, consumed := loStripVariationSelector(runes, i)
		i = consumed
		switch {
		case base == '\n':
			line++
			continue
		case base == ' ' || base == '\t' || base == '\r':
			continue
		case base == ':':
			toks = append(toks, loToken{Kind: loTokColon, Line: line})
		case base == '❓':
			toks = append(toks, loToken{Kind: loTokQuestion, Line: line})
		case base == '🚪':
			toks = append(toks, loToken{Kind: loTokDoor, Line: line})
		case base == '🔢':
			toks = append(toks, loToken{Kind: loTokI32Type, Line: line})
		case base == '🟰':
			toks = append(toks, loToken{Kind: loTokEq, Line: line})
		default:
			if val, ok := loStateDigits[base]; ok {
				toks = append(toks, loToken{Kind: loTokState, Val: val, Line: line})
				continue
			}
			if sym, ok := loArithOps[base]; ok {
				toks = append(toks, loToken{Kind: loTokArith, Text: sym, Line: line})
				continue
			}
			return nil, errors.New("lo_lexer: unrecognized token " + string(base) + " at line " + itoa(line))
		}
	}
	return toks, nil
}

// loRuneLen is a real, small helper kept for any future caller that needs a specific LO token's
// own real source byte width (e.g. real, future error-message underlining) -- unused today,
// named honestly rather than removed, since LO's own real error-handling design is explicitly
// still unstarted (NORTHSTAR.md item 3).
func loRuneLen(r rune) int { return utf8.RuneLen(r) }
