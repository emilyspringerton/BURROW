// lo_parser_test.go — real v0 verification for LO's own lexer/parser (lo_lexer.go/lo_parser.go),
// implementing LO/GRAMMAR.md's own canonical EBNF. Same real split every other emitter test in
// this repo already uses: check the LO frontend's own success/failure behavior and the C it
// produces directly here; verify real gcc/go acceptance and correct runtime behavior separately
// (not as a go test), by actually invoking `burrow build` + gcc/go on real .llll files.
package main

import (
	"strings"
	"testing"
)

// buildLoToC parses real LO source all the way through to real emitted C, the same real
// parse->analyze->emit pipeline `burrow build` itself runs, just called directly for a fast,
// isolated unit test.
func buildLoToC(t *testing.T, src string) (string, error) {
	t.Helper()
	program, err := LoParseProgram(src)
	if err != nil {
		return "", err
	}
	if err := RegionAnalyze(program); err != nil {
		t.Fatalf("unexpected region-analyze error: %v", err)
	}
	return EmitC(program)
}

func TestLoBareStateLiteral(t *testing.T) {
	// The simplest possible real LO program: Door then a single State, no Cond/Ternary/Arith at
	// all -- GRAMMAR.md's own `Ternary ::= ... | Value` and `Value ::= ... | State` alternatives.
	c, err := buildLoToC(t, "🚪 🔢 🌓")
	if err != nil {
		t.Fatalf("a bare State program should emit successfully: %v", err)
	}
	if !strings.Contains(c, "int Main(void) {") || !strings.Contains(c, "return 2;") {
		t.Errorf("🌓 alone should compile to a zero-arg Main returning 2: got %s", c)
	}
}

func TestLoEqTernaryTrueBranch(t *testing.T) {
	// A real, corrected version of LoLanguageSpec.pdf's own canonical worked example ("checks if
	// a vector component is equal to State 2 [here: State 1, for a real true-branch test], and
	// if so, performs a Base4 XOR; if not, falls back to State 0") -- GRAMMAR.md's own
	// EQ-then-QUERY-then-COLON shape, State 1 == State 1 (true) selects the XOR4 branch.
	c, err := buildLoToC(t, "🚪 🔢 🌒 🟰 🌒 ❓ 🔀 🌒 🌔 : 🌑")
	if err != nil {
		t.Fatalf("EQ ternary (true branch) should emit successfully: %v", err)
	}
	if !strings.Contains(c, "(1 == 1)") {
		t.Errorf("EQ should lower to real C ==, states as literal ints: got %s", c)
	}
	if !strings.Contains(c, "(1 ^ 3)") {
		t.Errorf("XOR4 should lower to real C bit-xor (no mod-4 wrap needed): got %s", c)
	}
	if !strings.Contains(c, "?") || !strings.Contains(c, ":") {
		t.Errorf("should compile to a real C ternary: got %s", c)
	}
}

func TestLoEqTernaryFalseBranch(t *testing.T) {
	// Same real shape, State 1 != State 2 (false) -- takes the else branch (State 0).
	c, err := buildLoToC(t, "🚪 🔢 🌒 🟰 🌓 ❓ 🔀 🌒 🌔 : 🌑")
	if err != nil {
		t.Fatalf("EQ ternary (false branch) should emit successfully: %v", err)
	}
	if !strings.Contains(c, "(1 == 2)") {
		t.Errorf("EQ should lower to real C ==: got %s", c)
	}
}

func TestLoPlus4WrapsBase4(t *testing.T) {
	// Real base4 mod-4 addition (GRAMMAR.md §5.1) -- (3 + 2) would be 5 in plain arithmetic,
	// must wrap to 1.
	c, err := buildLoToC(t, "🚪 🔢 ➕ 🌔 🌓")
	if err != nil {
		t.Fatalf("PLUS4 should emit successfully: %v", err)
	}
	if !strings.Contains(c, "((3 + 2) % 4)") {
		t.Errorf("PLUS4 should lower to a real (mod (+ a b) 4), not plain +: got %s", c)
	}
}

func TestLoMinus4WrapsNonNegative(t *testing.T) {
	// Real base4 mod-4 subtraction with a negative intermediate result (1 - 3 = -2 in plain
	// arithmetic) -- must wrap to the real, correct non-negative base4 state (2), not C's own
	// real negative-modulo result (-2).
	c, err := buildLoToC(t, "🚪 🔢 ➖ 🌒 🌔")
	if err != nil {
		t.Fatalf("MINUS4 should emit successfully: %v", err)
	}
	if !strings.Contains(c, "(((1 - 3) + 4) % 4)") {
		t.Errorf("MINUS4 should lower to a real (mod (+ (- a b) 4) 4): got %s", c)
	}
}

func TestLoAnd4Or4NeedNoWrapping(t *testing.T) {
	// GRAMMAR.md §5.1: AND4/OR4/XOR4 operate directly on States with no mod-4 wrap needed --
	// real, direct proof that this parser doesn't over-wrap them the way PLUS4/MINUS4 need.
	c, err := buildLoToC(t, "🚪 🔢 🔗 🌒 🌓")
	if err != nil {
		t.Fatalf("AND4 should emit successfully: %v", err)
	}
	if !strings.Contains(c, "return (1 & 2);") {
		t.Errorf("AND4 should lower to plain C &, no mod wrap: got %s", c)
	}
}

func TestLoVariationSelectorStripped(t *testing.T) {
	// Real, direct proof of GRAMMAR.md §1.2's own emoji-matching rule: a token codepoint
	// followed by a trailing VS16 (U+FE0F) lexes identically to the bare codepoint. None of this
	// v0's own real tokens carry a VS16 in the canonical table, so this synthetically appends one
	// to 🚪 to prove the stripping logic actually fires, not just that it's unreachable.
	toks, err := LoLex("🚪️🔢🌓")
	if err != nil {
		t.Fatalf("a token with a trailing VS16 should lex successfully: %v", err)
	}
	if len(toks) != 3 || toks[0].Kind != loTokDoor {
		t.Errorf("🚪+VS16 should lex as a single real DOOR token: got %d tokens, first kind %d", len(toks), toks[0].Kind)
	}
}

func TestLoMissingDoorIsError(t *testing.T) {
	_, err := LoParseProgram("🌑")
	if err == nil {
		t.Error("a program with no leading 🚪 DOOR should be a real, honest error")
	}
}

func TestLoUnrecognizedRuneIsError(t *testing.T) {
	_, err := LoLex("🚪🔢🐙")
	if err == nil {
		t.Error("an unrecognized emoji should be a real, honest lexer error, not silently skipped")
	}
}
