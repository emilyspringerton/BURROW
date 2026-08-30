// parser.go — Phase 2 (parser parity), real, faithful Go port of PARENA/src/parser.c + ast.h (VS0's
// own C-implemented reference recursive-descent parser and its generic S-expression tree).
// Founder real-time: "continue project BURROW" (Phase 1's own real, still-open next step, per
// NORTHSTAR.md and BACKLOG.md's own S206-84 entry).
//
// Real, honest architecture call, same one Phase 1's own header comment already made and this
// file inherits directly, not re-litigated: `selfhost/parser.prn` does not exist yet (only
// `selfhost/lexer.prn` has been ported to PARENA so far, per PARENA/NORTHSTAR.md's own current,
// honest self-hosting status) — meaning there is no real PARENA-source parser to compile through
// a PARENA-Go emitter that also doesn't exist yet. This file takes the same real fallback Phase 1
// already established: a faithful hand-port of the actual C reference, verified against the same
// real test corpus src/parser.c's own test file already established.
//
// Real, deliberate idiomatic departure from src/parser.c's own C shape: the C reference uses
// setjmp/longjmp to unwind the whole recursive-descent parse on the first error, so
// parse_form/parse_compound's mutual recursion doesn't need to thread an error return through
// every call site. Go has no equivalent primitive worth reaching for here (panic/recover exists,
// but is idiomatically reserved for truly exceptional conditions, not routine "stop parsing on a
// syntax error" control flow) — this port uses Go's own real, idiomatic multi-return (*Node,
// error) propagation instead, the same real "idiomatic translation, not mechanical
// transliteration" discipline lexer.go's own header comment already established for the Next()
// method's (Token, error) shape over the C reference's own out-param.
package main

// NodeType — real, direct port of ast.h's own NodeType enum.
type NodeType int

const (
	NodeList NodeType = iota // (...)
	NodeVec                  // [...]
	NodeMap                  // {...}
	NodeSymbol
	NodeKeyword
	NodeString
	NodeNumber
	NodeColon // standalone ':' inside a form, e.g. (x : Type @ Region)
	NodeAt    // standalone '@'
)

func (t NodeType) String() string {
	switch t {
	case NodeList:
		return "list"
	case NodeVec:
		return "vec"
	case NodeMap:
		return "map"
	case NodeSymbol:
		return "symbol"
	case NodeKeyword:
		return "keyword"
	case NodeString:
		return "string"
	case NodeNumber:
		return "number"
	case NodeColon:
		return ":"
	case NodeAt:
		return "@"
	default:
		return "?"
	}
}

// Node — real, direct port of ast.h's own Node struct. Real, deliberate simplification over the
// C reference's own shape: no separate Arena-backed child_capacity/growable-array bookkeeping --
// Go's own native, real `append` on a slice already handles growth, the same real "Go's own GC
// and native mutation make several of the C reference's own workarounds moot" reasoning lexer.go's
// header comment already applies to the lexer domain.
type Node struct {
	Type NodeType
	Line int
	// Text — atom payload (Symbol/Keyword/String/Number). Real, deliberate departure: NodeColon/
	// NodeAt carry Text == "" here (the C reference stores a real NULL for these — an empty Go
	// string is the natural, idiomatic equivalent, same real choice lexer.go's own punctuation
	// Token.Text already makes).
	Text string
	// Children — compound payload (List/Vec/Map).
	Children []*Node
}

func newAtom(nodeType NodeType, text string, line int) *Node {
	return &Node{Type: nodeType, Line: line, Text: text}
}

func newCompound(nodeType NodeType, line int) *Node {
	return &Node{Type: nodeType, Line: line, Children: nil}
}

// ParseError — real, direct port of the one real error shape src/parser.c's own fail() produces:
// a message plus the real line it was reported against. Real, deliberate simplification: the C
// reference's own fail() takes a `fmt_prefix` parameter, but grepping every real call site in
// src/parser.c shows it is ALWAYS passed as "" — dead in practice, not ported here.
type ParseError struct {
	Msg  string
	Line int
}

func (e *ParseError) Error() string {
	return e.Msg + " at line " + itoa(e.Line)
}

// Parser — real, direct port of parser.c's own Parser struct (minus the Arena field and the
// setjmp/longjmp error-unwind machinery, see this file's own header comment for why).
type Parser struct {
	lx  *Lexer
	cur Token
}

// closeName — real, direct port of parser.c's own close_name.
func closeName(t TokenType) string {
	switch t {
	case TRParen:
		return ")"
	case TRBracket:
		return "]"
	case TRBrace:
		return "}"
	default:
		return "?"
	}
}

func (p *Parser) advance() error {
	tok, err := p.lx.Next()
	if err != nil {
		// Real, direct use of the lex error's own real Line field (where the lexer actually
		// found the problem), not the stale p.cur from before this call -- matching
		// src/lexer.c's own lexer_next, whose error messages already carry the real line
		// internally rather than relying on the caller's own stale position.
		if lexErr, ok := err.(*LexError); ok {
			return &ParseError{Msg: lexErr.Error(), Line: lexErr.Line}
		}
		return &ParseError{Msg: err.Error(), Line: p.cur.Line}
	}
	p.cur = tok
	return nil
}

// parseCompound — real, direct port of parser.c's own parse_compound: consumes the opening
// bracket, then real forms until the matching close token, reporting a real, honest error on EOF
// (unterminated) or the wrong close token (mismatched bracket kind) rather than guessing.
func (p *Parser) parseCompound(nodeType NodeType, closeTok TokenType) (*Node, error) {
	openLine := p.cur.Line
	if err := p.advance(); err != nil { // consume the opening bracket
		return nil, err
	}
	node := newCompound(nodeType, openLine)
	for {
		if p.cur.Kind == closeTok {
			if err := p.advance(); err != nil {
				return nil, err
			}
			return node, nil
		}
		if p.cur.Kind == TEof {
			return nil, &ParseError{
				Msg:  "unterminated form: expected '" + closeName(closeTok) + "' to close the form opened",
				Line: openLine,
			}
		}
		if p.cur.Kind == TRParen || p.cur.Kind == TRBracket || p.cur.Kind == TRBrace {
			// A close token, but the WRONG one -- mismatched bracket kind, e.g. "(foo]" or "[bar)".
			return nil, &ParseError{
				Msg:  "mismatched bracket: expected '" + closeName(closeTok) + "' but found '" + closeName(p.cur.Kind) + "'",
				Line: p.cur.Line,
			}
		}
		child, err := p.parseForm()
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, child)
	}
}

// parseForm — real, direct port of parser.c's own parse_form.
func (p *Parser) parseForm() (*Node, error) {
	switch p.cur.Kind {
	case TLParen:
		return p.parseCompound(NodeList, TRParen)
	case TLBracket:
		return p.parseCompound(NodeVec, TRBracket)
	case TLBrace:
		return p.parseCompound(NodeMap, TRBrace)
	case TSymbol:
		n := newAtom(NodeSymbol, p.cur.Text, p.cur.Line)
		return n, p.advance()
	case TKeyword:
		n := newAtom(NodeKeyword, p.cur.Text, p.cur.Line)
		return n, p.advance()
	case TString:
		n := newAtom(NodeString, p.cur.Text, p.cur.Line)
		return n, p.advance()
	case TNumber:
		n := newAtom(NodeNumber, p.cur.Text, p.cur.Line)
		return n, p.advance()
	case TColon:
		n := newAtom(NodeColon, "", p.cur.Line)
		return n, p.advance()
	case TAt:
		n := newAtom(NodeAt, "", p.cur.Line)
		return n, p.advance()
	case TRParen, TRBracket, TRBrace:
		return nil, &ParseError{
			Msg:  "unexpected '" + closeName(p.cur.Kind) + "' with no matching open bracket",
			Line: p.cur.Line,
		}
	default: // TEof
		return nil, &ParseError{Msg: "unexpected end of file, expected a form", Line: p.cur.Line}
	}
}

// ParseProgram — real, direct port of parser.c's own parse_program: primes the first token, then
// reads real top-level forms until EOF, wrapping them all in one NodeList (matching the C
// reference's own real "the whole file is one implicit top-level list" convention).
func ParseProgram(src string) (*Node, error) {
	p := &Parser{lx: NewLexer(src)}
	if err := p.advance(); err != nil { // prime the first token
		return nil, err
	}

	program := newCompound(NodeList, 1)
	for p.cur.Kind != TEof {
		if p.cur.Kind == TRParen || p.cur.Kind == TRBracket || p.cur.Kind == TRBrace {
			return nil, &ParseError{
				Msg:  "unexpected '" + closeName(p.cur.Kind) + "' with no matching open bracket",
				Line: p.cur.Line,
			}
		}
		form, err := p.parseForm()
		if err != nil {
			return nil, err
		}
		program.Children = append(program.Children, form)
	}
	return program, nil
}
