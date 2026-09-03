// lo_parser.go — real LO parser implementing LO/GRAMMAR.md's own canonical EBNF (the Phase-1
// subset lo_lexer.go's own header documents), producing burrow's existing *Node AST directly
// (the exact same shape parser.go's own ParseProgram produces for real PARENA S-expression
// source) — so RegionAnalyze/EmitC/EmitGo need ZERO changes; LO is purely a new frontend sharing
// every existing backend, matching LO/NORTHSTAR.md's own real architectural correction.
//
// Real, honest note on this v0's own real "program" shape: GRAMMAR.md's own Program/Door/Expr
// productions define a closed expression with no external inputs at all -- LO itself has no
// function/parameter concept anywhere in the canonical grammar (only its `qi` FRONTEND, a real,
// separate, later phase, introduces named functions/parameters). This parser's own real job is
// therefore to synthesize a real, zero-argument PARENA `defn` wrapping that one closed
// expression -- not to invent a parameter mechanism the grammar doesn't define (an earlier
// version of this file did exactly that, using 🧲 as a positional-parameter marker; removed here
// since it wasn't grammar-authorized -- reconciled with GRAMMAR.md, not layered on top of it).
package main

import "errors"

type loParser struct {
	toks []loToken
	pos  int
}

func (p *loParser) peek() (loToken, bool) {
	if p.pos >= len(p.toks) {
		return loToken{}, false
	}
	return p.toks[p.pos], true
}

func (p *loParser) next() (loToken, bool) {
	t, ok := p.peek()
	if ok {
		p.pos++
	}
	return t, ok
}

// LoParseProgram is the real, top-level entry point. Returns the exact same *Node program shape
// ParseProgram produces: a NodeList whose children are top-level forms -- here, always exactly
// one real, synthesized, zero-arg (defn Main [] : I32 body) form (see this file's own header for
// why LO itself has no named multi-function, parameterized programs).
func LoParseProgram(src string) (*Node, error) {
	toks, lerr := LoLex(src)
	if lerr != nil {
		return nil, lerr
	}
	p := &loParser{toks: toks}

	// Door ::= DOOR I32 -- GRAMMAR.md's own TypedExpr also allows a bare Expr with no Door at
	// all; this v0 requires it (a real, narrower, honest choice, not a violation -- Door-then-
	// Expr is a fully valid Program under the canonical grammar, it's just the only alternative
	// this compiler currently implements).
	door, ok := p.next()
	if !ok || door.Kind != loTokDoor {
		return nil, errors.New("lo_parser: v0 requires a program to start with DOOR (🚪); GRAMMAR.md's own bare-Expr alternative isn't implemented yet")
	}
	retTy, ok := p.next()
	if !ok || retTy.Kind != loTokI32Type {
		return nil, errors.New("lo_parser: v0 only supports I32 (🔢) as the Door's TypeAtom")
	}

	body, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if p.pos != len(p.toks) {
		return nil, errors.New("lo_parser: unexpected trailing tokens after the program's own single Expr at line " + itoa(p.toks[p.pos].Line))
	}

	defn := newCompound(NodeList, 1)
	defn.Children = []*Node{
		newAtom(NodeSymbol, "defn", 1),
		newAtom(NodeSymbol, "Main", 1),
		newCompound(NodeVec, 1), // zero params -- see this file's own header
		newAtom(NodeColon, "", 1),
		newAtom(NodeSymbol, "I32", 1),
		body,
	}

	program := newCompound(NodeList, 1)
	program.Children = []*Node{defn}
	return program, nil
}

// parseExpr ::= Ternary. Real, direct implementation of GRAMMAR.md §2/§3: Ternary is
// right-associative (a false-branch that is itself a Ternary parses correctly via this
// function's own recursive call for elseE, matching §3 rule 1 exactly), and Cond (the EQ
// production) binds tighter than QUERY/COLON per §3 rule 2 -- a bare Ternary is never used
// directly as a Cond, matching the grammar's own real "no bare-state/ternary-as-condition"
// boundary (GRAMMAR.md explicitly leaves general truthiness out of scope for v1).
func (p *loParser) parseExpr() (*Node, error) {
	value, err := p.parseValue()
	if err != nil {
		return nil, err
	}
	// Cond ::= Value EQ Value -- only recognized when the NEXT token is EQ; otherwise this
	// Value (a bare Arith/State) is the whole Expr, matching GRAMMAR.md's own
	// `Ternary ::= Cond QUERY Expr COLON Expr | Value` alternative directly.
	eqTok, ok := p.peek()
	if !ok || eqTok.Kind != loTokEq {
		return value, nil
	}
	p.next() // consume EQ
	rhs, err := p.parseValue()
	if err != nil {
		return nil, err
	}
	cond := newCompound(NodeList, eqTok.Line)
	cond.Children = []*Node{newAtom(NodeSymbol, "=", eqTok.Line), value, rhs}

	qTok, ok := p.next()
	if !ok || qTok.Kind != loTokQuestion {
		return nil, errors.New("lo_parser: a Cond (Value EQ Value) must be followed by QUERY (❓) at line " + itoa(eqTok.Line))
	}
	thenE, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	colonTok, ok := p.next()
	if !ok || colonTok.Kind != loTokColon {
		return nil, errors.New("lo_parser: expected COLON (:) at line " + itoa(qTok.Line))
	}
	elseE, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	ternary := newCompound(NodeList, qTok.Line)
	ternary.Children = []*Node{newAtom(NodeSymbol, "if", qTok.Line), cond, thenE, elseE}
	return ternary, nil
}

// parseValue ::= Arith | State. Real, direct implementation of GRAMMAR.md §2's own `Value`
// nonterminal, narrowed to this v0's own real subset (Labeled/LinAlg/VectorLit/MagnetExpr/nested
// Ternary-as-Value all need String/Vec/Matrix support this v0 doesn't attempt yet).
func (p *loParser) parseValue() (*Node, error) {
	tok, ok := p.next()
	if !ok {
		return nil, errors.New("lo_parser: unexpected end of program")
	}
	switch tok.Kind {
	case loTokState:
		return newAtom(NodeNumber, itoa(tok.Val), tok.Line), nil
	case loTokArith:
		return p.parseArith(tok)
	default:
		return nil, errors.New("lo_parser: expected a State or Arith Value at line " + itoa(tok.Line))
	}
}

// parseArith ::= ArithOp Value Value (prefix -- see NORTHSTAR.md's own real note that this
// monorepo's entire PARENA/burrow toolchain is prefix-notation throughout; GRAMMAR.md §2 writes
// `Value ArithOp Value` infix for readability, but never actually resolves relative precedence
// between chained/nested Arith uses (§2's own note: "the source material never establishes
// relative precedence... every real example nests exactly one operator deep"). Real, deliberate
// v0 decision, consistent with §3 rule 3's own "Phase 0 decision made because the source is
// silent, not a discovered fact": this parser reads ArithOp PREFIX (`op lhs rhs`), which is
// fully unambiguous with no precedence table needed at all, and both operands are themselves
// full Values (State or a nested Arith) -- real, recursive, self-delimiting per operator arity.
func (p *loParser) parseArith(op loToken) (*Node, error) {
	lhs, err := p.parseValue()
	if err != nil {
		return nil, err
	}
	rhs, err := p.parseValue()
	if err != nil {
		return nil, err
	}
	switch op.Text {
	case "plus4":
		// Real base4 mod-4 addition (GRAMMAR.md §5.1): (mod (+ lhs rhs) 4). Always
		// non-negative since neither a State value nor + can go negative here.
		return loWrapMod4(op.Line, "+", lhs, rhs, false), nil
	case "minus4":
		// Real base4 mod-4 subtraction: (mod (+ (- lhs rhs) 4) 4) -- the real "add the modulus
		// before reducing" guard needed since lhs - rhs can go negative (e.g. 1 - 3 = -2), and
		// C/Go's own % on a negative operand does not return the real, correct non-negative
		// base4 state the way this guard does.
		return loWrapMod4(op.Line, "-", lhs, rhs, true), nil
	default:
		// and4/or4/xor4: real, direct bitwise ops, no wrapping needed (2-bit values already
		// stay in [0,3] under AND/OR/XOR).
		parenaOp := map[string]string{"and4": "bit-and", "or4": "bit-or", "xor4": "bit-xor"}[op.Text]
		n := newCompound(NodeList, op.Line)
		n.Children = []*Node{newAtom(NodeSymbol, parenaOp, op.Line), lhs, rhs}
		return n, nil
	}
}

// loWrapMod4 builds (mod (OP lhs rhs) 4), optionally guarding against a negative intermediate
// result first by adding 4 before the outer mod (needed for subtraction, not addition -- see
// parseArith's own two call sites).
func loWrapMod4(line int, op string, lhs, rhs *Node, guardNegative bool) *Node {
	inner := newCompound(NodeList, line)
	inner.Children = []*Node{newAtom(NodeSymbol, op, line), lhs, rhs}
	target := inner
	if guardNegative {
		plus4 := newCompound(NodeList, line)
		plus4.Children = []*Node{newAtom(NodeSymbol, "+", line), inner, newAtom(NodeNumber, "4", line)}
		target = plus4
	}
	wrapped := newCompound(NodeList, line)
	wrapped.Children = []*Node{newAtom(NodeSymbol, "mod", line), target, newAtom(NodeNumber, "4", line)}
	return wrapped
}
