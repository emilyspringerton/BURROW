// unparse.go — real *Node -> .prn source-text serializer. Real, direct trigger: LO's own
// compiler (lo_lexer.go/lo_parser.go) produces burrow's in-memory *Node AST directly, which is
// exactly right for burrow's own C/Go targets (EmitC/EmitGo consume *Node), but the real,
// original architectural correction in LO/NORTHSTAR.md is that LO reaches TypeScript/Java by
// emitting real .prn TEXT for the ALREADY-EXISTING `parena build foo.prn -o foo.ts/foo.java` to
// consume — burrow's own two emitters don't do TS/Java at all (see cmdBuild's own real
// notImplemented calls). This file is that missing link: `burrow build foo.llll -o foo.prn`
// unparses the real Node tree LO produced back into real, valid PARENA S-expression source,
// which `parena build` (a real, separate, already-working binary) then takes the rest of the way
// to any of its own real targets, C/TS/Java included — no new backend, exactly as designed.
//
// Real, general implementation, not narrowed to LO's own output shape specifically: every real
// NodeType this file's own sibling parser.go defines (List/Vec/Map/Symbol/Keyword/String/Number/
// Colon/At) round-trips, so this also works on a real, hand-written .prn file reparsed and
// reprinted, not just LO-synthesized ones.
package main

import "strings"

// UnparseProgram serializes a real *Node program (as ParseProgram/LoParseProgram both produce --
// a NodeList whose children are top-level forms) back into real .prn source text, one top-level
// form per line, each form printed as a single-line S-expression (real, minimal formatting --
// `burrow fmt`, once it exists per NORTHSTAR.md's own Phase 2, is the real place indentation
// belongs, not this function).
func UnparseProgram(program *Node) string {
	var b strings.Builder
	for _, form := range program.Children {
		unparseNode(&b, form)
		b.WriteString("\n")
	}
	return b.String()
}

func unparseNode(b *strings.Builder, n *Node) {
	switch n.Type {
	case NodeSymbol, NodeNumber:
		b.WriteString(n.Text)
	case NodeKeyword:
		// Real, direct round-trip: lexer.go's own real lexKeyword already includes the leading
		// ':' in Text (see emit_go.go's own get-field handling, which strips it back off) --
		// writing Text as-is here is therefore already correct, not missing a colon.
		b.WriteString(n.Text)
	case NodeString:
		b.WriteString("\"")
		b.WriteString(n.Text)
		b.WriteString("\"")
	case NodeColon:
		b.WriteString(":")
	case NodeAt:
		b.WriteString("@")
	case NodeList:
		b.WriteString("(")
		unparseChildren(b, n.Children)
		b.WriteString(")")
	case NodeVec:
		b.WriteString("[")
		unparseChildren(b, n.Children)
		b.WriteString("]")
	case NodeMap:
		b.WriteString("{")
		unparseChildren(b, n.Children)
		b.WriteString("}")
	}
}

func unparseChildren(b *strings.Builder, children []*Node) {
	for i, c := range children {
		if i > 0 {
			b.WriteString(" ")
		}
		unparseNode(b, c)
	}
}
