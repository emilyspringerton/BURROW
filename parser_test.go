// parser_test.go — real port of tests/test_lexer_parser.c's own real test scenarios (every
// balanced and imbalanced case, matching DoD domain 1's own required verification method: "Unit
// tests on balanced and imbalanced S-expressions") into idiomatic Go subtests. Every expectation
// here is copied verbatim from that file's own real assertions, not re-derived.
package main

import "testing"

func TestParseSimpleBalancedListWithVectorChild(t *testing.T) {
	n, err := ParseProgram("(defn foo [] 1)")
	if err != nil {
		t.Fatalf("simple balanced list with a vector child should parse: %v", err)
	}
	if len(n.Children) != 1 {
		t.Fatalf("top-level program should have exactly one form, got %d", len(n.Children))
	}
	defn := n.Children[0]
	if defn.Type != NodeList {
		t.Errorf("the one form should be a list, got %s", defn.Type)
	}
	if len(defn.Children) != 4 {
		t.Fatalf("(defn foo [] 1) should have 4 children: defn, foo, [], 1, got %d", len(defn.Children))
	}
	if defn.Children[0].Type != NodeSymbol || defn.Children[0].Text != "defn" {
		t.Errorf("first child should be the symbol 'defn', got %s %q", defn.Children[0].Type, defn.Children[0].Text)
	}
	if defn.Children[2].Type != NodeVec || len(defn.Children[2].Children) != 0 {
		t.Errorf("third child should be an empty vector []")
	}
}

func TestParseRealScratchToBufferExample(t *testing.T) {
	src := "(defn load-config [(buf-arena : Arena @ :region/buffer)]\n" +
		"  (with-arena [scratch :region/scratch 1024]\n" +
		"    (let [temp-str (alloc scratch String \"config.json\")]\n" +
		"      temp-str)))"
	if _, err := ParseProgram(src); err != nil {
		t.Fatalf("real VS0 test.prn-style scratch-to-buffer promotion example should parse: %v", err)
	}
}

func TestParseMapLiteral(t *testing.T) {
	n, err := ParseProgram("{:field1 val1 :field2 val2}")
	if err != nil {
		t.Fatalf("map literal should parse: %v", err)
	}
	m := n.Children[0]
	if m.Type != NodeMap {
		t.Errorf("top form should be a map, got %s", m.Type)
	}
	if len(m.Children) != 4 {
		t.Fatalf("map should have 4 flat children (2 keyword/value pairs), got %d", len(m.Children))
	}
	if m.Children[0].Type != NodeKeyword || m.Children[0].Text != ":field1" {
		t.Errorf("first map child should be the keyword :field1, got %s %q", m.Children[0].Type, m.Children[0].Text)
	}
}

func TestParseBangAmpPrefixedSymbols(t *testing.T) {
	n, err := ParseProgram("!file &var &mut")
	if err != nil {
		t.Fatalf("bang/amp-prefixed symbols (linear/borrow forms) should parse as symbols: %v", err)
	}
	checkNodeAtom(t, n.Children[0], NodeSymbol, "!file", "!file should be one SYMBOL token, not a separate operator + symbol")
	checkNodeAtom(t, n.Children[1], NodeSymbol, "&var", "&var should be one SYMBOL token")
	checkNodeAtom(t, n.Children[2], NodeSymbol, "&mut", "&mut should be one SYMBOL token")
}

func checkNodeAtom(t *testing.T, n *Node, wantType NodeType, wantText, label string) {
	t.Helper()
	if n.Type != wantType || n.Text != wantText {
		t.Errorf("%s (got %s %q)", label, n.Type, n.Text)
	}
}

func TestParseLineCommentsSkipped(t *testing.T) {
	n, err := ParseProgram("; a comment\n(foo 1) ;; trailing comment")
	if err != nil {
		t.Fatalf("line comments should be skipped, both ; and ;; forms: %v", err)
	}
	if len(n.Children) != 1 {
		t.Errorf("comment-only line should contribute no form, got %d top-level forms", len(n.Children))
	}
}

func TestParseStandaloneColonAndAt(t *testing.T) {
	n, err := ParseProgram("(x : Type @ :region/scratch)")
	if err != nil {
		t.Fatalf("standalone ':' and '@' tokens should parse distinctly from keywords: %v", err)
	}
	form := n.Children[0]
	if form.Children[1].Type != NodeColon {
		t.Errorf("standalone ':' should parse as NodeColon, not a keyword, got %s", form.Children[1].Type)
	}
	if form.Children[3].Type != NodeAt {
		t.Errorf("'@' should parse as NodeAt, got %s", form.Children[3].Type)
	}
	checkNodeAtom(t, form.Children[4], NodeKeyword, ":region/scratch",
		":region/scratch (no space after ':') should parse as one KEYWORD token")
}

func TestParseNegativeAndDecimalNumbers(t *testing.T) {
	n, err := ParseProgram("-42 3.5 100")
	if err != nil {
		t.Fatalf("negative integers and decimals should parse as numbers: %v", err)
	}
	checkNodeAtom(t, n.Children[0], NodeNumber, "-42", "-42 should parse as a number")
	checkNodeAtom(t, n.Children[1], NodeNumber, "3.5", "3.5 should parse as a number")
}

func TestParseStringWithEscapeSequence(t *testing.T) {
	n, err := ParseProgram(`"hello\nworld"`)
	if err != nil {
		t.Fatalf("string with an escape sequence should parse: %v", err)
	}
	if n.Children[0].Type != NodeString || len(n.Children[0].Text) != 11 {
		t.Errorf("escaped \\n should count as one character in the decoded string, got %q (len %d)",
			n.Children[0].Text, len(n.Children[0].Text))
	}
}

func TestParseEmptyFile(t *testing.T) {
	n, err := ParseProgram("")
	if err != nil {
		t.Fatalf("an empty file should parse to an empty program, not an error: %v", err)
	}
	if len(n.Children) != 0 {
		t.Errorf("empty file should produce zero top-level forms, got %d", len(n.Children))
	}
}

// --- imbalanced / malformed S-expressions (DoD's own required negative case) ---

func expectParseError(t *testing.T, src, label string) {
	t.Helper()
	n, err := ParseProgram(src)
	if err == nil {
		t.Errorf("%s (got a real parse with no error)", label)
		return
	}
	if n != nil {
		t.Errorf("%s (node should be nil on a real parse error)", label)
	}
	t.Logf("%s -- real error: %v", label, err)
}

func TestParseUnterminatedList(t *testing.T) {
	expectParseError(t, "(defn foo [] 1", "unterminated list (missing final paren) should be a parse error")
}

func TestParseMismatchedBracketKind(t *testing.T) {
	expectParseError(t, "(foo (bar]", "mismatched bracket kind (opened with (, closed with ]) should be a parse error")
}

func TestParseStrayClosingParen(t *testing.T) {
	expectParseError(t, ")", "a stray closing paren with nothing open should be a parse error")
}

func TestParseUnterminatedStringLiteral(t *testing.T) {
	expectParseError(t, `(foo "unterminated string`, "unterminated string literal should be a parse error")
}

func TestParseMismatchedNestedBracket(t *testing.T) {
	expectParseError(t, "[1 2 (3 4]", "mismatched nested bracket should be a parse error")
}
