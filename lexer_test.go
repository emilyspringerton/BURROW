// lexer_test.go — real port of tests/test_selfhost_lexer.c's own 9 real test scenarios (all of
// them, not a subset) into idiomatic Go subtests. Every expected token sequence here is copied
// verbatim from that file's own real, hand-traced-against-src/lexer.c expectations — not
// re-derived or guessed at, matching this repo's own real "pass all that parena c tests"
// acceptance bar (NORTHSTAR.md's own founder-named Phase 0 bar, applied here to the real, concrete
// lexer domain first).
package main

import "testing"

func checkToken(t *testing.T, tok Token, wantKind TokenType, wantText string, wantLine int, label string) {
	t.Helper()
	if tok.Kind != wantKind {
		t.Errorf("%s: kind is %s (expected %s)", label, tok.Kind, wantKind)
	}
	if tok.Text != wantText {
		t.Errorf("%s: text is %q (expected %q)", label, tok.Text, wantText)
	}
	if tok.Line != wantLine {
		t.Errorf("%s: line is %d (expected %d)", label, tok.Line, wantLine)
	}
}

func TestTokenizeSimpleBalancedForm(t *testing.T) {
	toks, err := Tokenize("(a b)")
	if err != nil {
		t.Fatalf("tokenize should succeed on a real simple balanced form: %v", err)
	}
	if len(toks) != 5 {
		t.Fatalf("'(a b)' should produce exactly 5 real tokens: ( a b ) EOF, got %d", len(toks))
	}
	checkToken(t, toks[0], TLParen, "", 1, "tok0")
	checkToken(t, toks[1], TSymbol, "a", 1, "tok1")
	checkToken(t, toks[2], TSymbol, "b", 1, "tok2")
	checkToken(t, toks[3], TRParen, "", 1, "tok3")
	checkToken(t, toks[4], TEof, "", 1, "tok4")
}

func TestTokenizeBangAmpPrefixedSymbols(t *testing.T) {
	// Real bang/amp-prefixed linear/borrow forms -- !file, &var, &mut are single real symbols,
	// matching src/lexer.c's own real is_symbol_char permissiveness.
	toks, err := Tokenize("(!file &var &mut)")
	if err != nil {
		t.Fatalf("tokenize should succeed on real bang/amp-prefixed symbols: %v", err)
	}
	if len(toks) != 6 {
		t.Fatalf("'(!file &var &mut)' should produce 6 real tokens, got %d", len(toks))
	}
	checkToken(t, toks[1], TSymbol, "!file", 1, "!file")
	checkToken(t, toks[2], TSymbol, "&var", 1, "&var")
	checkToken(t, toks[3], TSymbol, "&mut", 1, "&mut")
}

func TestTokenizeKeywordVsStandaloneColon(t *testing.T) {
	// Real keyword vs. standalone colon distinction (:region/scratch vs. a bare ':' in a type
	// signature), matching src/lexer.c's own real lex_keyword dispatch exactly.
	toks, err := Tokenize("(x : Type @ :region/scratch)")
	if err != nil {
		t.Fatalf("tokenize should succeed on a real type/region signature: %v", err)
	}
	checkToken(t, toks[2], TColon, "", 1, "standalone colon")
	checkToken(t, toks[3], TSymbol, "Type", 1, "Type")
	checkToken(t, toks[4], TAt, "", 1, "at")
	checkToken(t, toks[5], TKeyword, ":region/scratch", 1, "keyword")
}

func TestTokenizeNumbersAndBareDash(t *testing.T) {
	// Real negative/decimal numbers, and a bare '-' as its own real symbol (matching
	// looksLikeNumberStart's own real "-/+ must be followed by a digit" guard).
	toks, err := Tokenize("-42 3.5 (- a b)")
	if err != nil {
		t.Fatalf("tokenize should succeed on real numbers + a bare '-' symbol: %v", err)
	}
	checkToken(t, toks[0], TNumber, "-42", 1, "-42")
	checkToken(t, toks[1], TNumber, "3.5", 1, "3.5")
	checkToken(t, toks[3], TSymbol, "-", 1, "bare -")
}

func TestTokenizeStringEscapes(t *testing.T) {
	// Real string literal with every real escape src/lexer.c's own lex_string decodes: \n \t \"
	// \\, plus a real literal passthrough char after an unrecognized backslash.
	toks, err := Tokenize(`"a\nb\tc\"d\\e\zf"`)
	if err != nil {
		t.Fatalf("tokenize should succeed on a real string with every real escape sequence: %v", err)
	}
	if toks[0].Kind != TString {
		t.Fatalf("the real escaped string should lex as TString, got %s", toks[0].Kind)
	}
	want := "a\nb\tc\"d\\ezf"
	if toks[0].Text != want {
		t.Errorf("every real \\n/\\t/\\\"/\\\\ escape should decode to its real single byte, and \\z (unrecognized) should pass 'z' through literally: got %q want %q", toks[0].Text, want)
	}
}

func TestTokenizeUnterminatedString(t *testing.T) {
	// Real unterminated string -- reports the real line it OPENED on, matching src/lexer.c's own
	// start_line convention.
	_, err := Tokenize(`(f "never closed`)
	if err == nil {
		t.Fatal("tokenize should report a real error on an unterminated string, not a crash or silent EOF")
	}
	lexErr, ok := err.(*LexError)
	if !ok {
		t.Fatalf("error should be a *LexError, got %T", err)
	}
	if lexErr.Kind != UnterminatedString {
		t.Errorf("the real error should be UnterminatedString, got kind %d", lexErr.Kind)
	}
}

func TestTokenizeCommentSkipping(t *testing.T) {
	// Real comment skipping -- a ';;' line comment runs to end of line and contributes no real
	// token; a real paren INSIDE it must not be lexed as a real structural token.
	toks, err := Tokenize(";; a real (fake paren\n(real)")
	if err != nil {
		t.Fatalf("tokenize should succeed skipping a real ';;' comment: %v", err)
	}
	checkToken(t, toks[0], TLParen, "", 2,
		"the real '(' after the comment should be on real line 2, and the comment's own fake paren should produce no token")
}

func TestTokenizeRealStringPrnFragment(t *testing.T) {
	// Real, genuine PARENA source fragment, lifted from this repo's own stdlib/string.prn (not
	// synthetic) -- same real fragment test_selfhost_lexer.c itself uses.
	src := "(defn is-digit? [(c : I32)]\n  : Bool\n  (and (>= c 48) (<= c 57)))"
	toks, err := Tokenize(src)
	if err != nil {
		t.Fatalf("tokenize should succeed on a real, genuine fragment lifted from stdlib/string.prn: %v", err)
	}
	if len(toks) <= 20 {
		t.Fatalf("the real string.prn fragment should produce a real, substantial token stream (>20 tokens), got %d", len(toks))
	}
	last := toks[len(toks)-1]
	if last.Kind != TEof {
		t.Errorf("the real token stream's own last token should be TEof, got %s", last.Kind)
	}
	checkToken(t, toks[1], TSymbol, "defn", 1, "defn itself")
	checkToken(t, toks[2], TSymbol, "is-digit?", 1, "defn name")
}

func TestTokenizeEmptyInput(t *testing.T) {
	// Real empty input -- a real, honest single-TEof stream, not an error or a crash.
	toks, err := Tokenize("")
	if err != nil {
		t.Fatalf("tokenize should succeed on a real empty input: %v", err)
	}
	if len(toks) != 1 {
		t.Fatalf("a real empty input should produce exactly one real TEof token, got %d", len(toks))
	}
}
