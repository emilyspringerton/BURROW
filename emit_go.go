// emit_go.go — Phase 6, real, narrow v0 native Go emission target. Founder real-time: "continue
// DUNG's parena work, we want to emit go with burrow" — the real trigger BURROW/NORTHSTAR.md's
// own Phase 6 section named as its one open precondition: "No real host has asked for the Go
// emission target (Phase 6) yet." DUNG is that real host now — its own Go-hosted entrypoint
// (cmd/dung/main.go) can call a burrow-emitted Go package directly, natively, with zero cgo/FFI
// boundary, which the C target this same session already shipped structurally cannot offer a
// Go host.
//
// Real, honest, deliberate scope: the exact SAME narrow v0 template emit_c.go already proved out
// (scalar I32/F64/Bool/String params, no Arena/region annotations, a body that is exactly ONE
// real expression: number/symbol literals, the real binop set, `if` as a real Go conditional
// expression via an inline func-literal ternary substitute -- Go has no `?:`, see emitGoExpr's
// own "if" case -- and calls to another top-level `defn`) -- NOT the full `emit.c`. Growing this
// scope (defstruct/defenum/match/loop/Result/Vec) is real, separate, unstarted work, same as
// every other emit_*.go's own honest boundary.
//
// GC-off-safe by construction, matching NORTHSTAR.md's own "The real, honest GC-off design
// question" section directly: v0-scope emitted functions have scalar params, one-expression
// bodies, no let/blocks/structs/slices/maps/closures -- already GC-irrelevant, no heap allocation
// possible in the emitted code itself. A real host embedding this package is free to run
// debug.SetGCPercent(-1) around calls into it as an informed, deliberate choice; this emitter
// makes that choice SAFE, not the choice itself.
package main

import (
	"errors"
	"go/format"
	"strings"
)

// mangleGo — real, direct analog of mangleC's own kebab-case -> snake_case rule, EXCEPT Go's own
// real, standard exported-identifier convention is exported (capitalized) CamelCase, not
// snake_case (golint/go vet both flag snake_case identifiers) -- every top-level defn this v0
// emits is meant to be called from a real, separate Go host package (DUNG's own cmd/dung, the
// whole point of this target), so every emitted function name must be exported. kebab-case ->
// PascalCase: split on '-', capitalize each segment's first rune, join.
func mangleGo(name string) string {
	// Real gap found live (PARENA/stdlib/datetime.prn's own is-leap-year?): a trailing `?`/`!`
	// (the real, common Lisp predicate/mutation-sigil convention) produced an illegal Go
	// identifier character, confirmed via a real `gofmt` failure ("illegal character U+003F
	// '?'"). `src/emit.c`'s own C emitter already has the real, matching convention for this
	// exact case (`?`/`!` -> `_`, same as `-`/`/`) — mirrored here rather than invented fresh.
	name = strings.NewReplacer("?", "_", "!", "_").Replace(name)
	parts := strings.Split(name, "-")
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		b.WriteString(strings.ToUpper(p[:1]))
		b.WriteString(p[1:])
	}
	return b.String()
}

// resolveGoType — real, direct analog of resolveCType's own scalar-type table, Go's own real
// scalar types this v0 targets: Unit -> real Go has no true unit/void expression type, matching
// the same real "Unit means this function is called for effect" shape emit_ts.go/emit_java.c
// already resolve to void/Unit -- Go's own nearest real, idiomatic equivalent for a function with
// nothing meaningful to return is an explicit empty struct{} (the real, standard Go zero-alloc
// "no data" idiom, not a special case invented for this emitter). I32->int32 (real, fixed-width,
// not Go's own platform-sized `int` -- matching I32's own real, fixed-width contract PARENA's
// language itself already documents, not Go convenience), Bool->bool, F64->float64,
// String->string (Go's own real, built-in, already-GC-safe-for-string-literals type -- a v0-scope
// function only ever passes through or returns a String parameter/literal itself, per this file's
// own "one real expression" scope, so no string-building/concatenation allocation is reachable
// here regardless).
// resolveGoType takes `knownStructs` now — real, new capability (defstruct support, added the
// same day PARENA's own real k8s/helm stdlib packages needed it, see EmitGo's own doc comment):
// a registered defstruct name resolves to its own real, exported Go struct type name (PascalCase
// via mangleGo, matching every real generated Go struct field/function name this emitter already
// produces), passed by value -- the same real semantics parena-c's own C emitter already gives
// struct-typed parameters (confirmed by reading its own emitted output directly: `Deployment d`,
// not a pointer).
func resolveGoType(typeName string, knownStructs map[string]bool) (string, error) {
	switch typeName {
	case "Unit":
		return "struct{}", nil
	case "I32":
		return "int32", nil
	case "Bool":
		return "bool", nil
	case "F64":
		return "float64", nil
	case "String":
		return "string", nil
	default:
		if knownStructs[typeName] {
			return mangleGo(typeName), nil
		}
		return "", errors.New("emit_go: unsupported parameter/return type (v0 only understands I32/F64/Bool/String/Unit, or a registered defstruct type)")
	}
}

// goBinopTable — same real operator SET emit_c.go's own cBinopTable uses (confirmed against the
// same real emit.c reference), Go's own real surface syntax differs only where C and Go actually
// diverge: Go has no `&&`-vs-C `&&` difference for and/or (both use the same real `&&`/`||`
// tokens C does), so this table is deliberately identical to cBinopTable's own values -- kept as
// its own real, separate table (not a shared alias) so a future real per-target divergence (there
// is none known today) doesn't require unpicking a shared one.
var goBinopTable = map[string]string{
	"+": "+", "-": "-", "*": "*", "/": "/",
	"<": "<", ">": ">", "<=": "<=", ">=": ">=",
	"=": "==", "and": "&&", "or": "||",
	"bit-and": "&", "bit-or": "|", "bit-xor": "^", "mod": "%",
}

// emitGoScope — real, direct analog of emitCScope; same real two-tier resolution (local params,
// then known top-level defns), same real reason (catch a real unknown/namespaced identifier like
// `math/pi` as a real, honest error instead of emitting broken Go).
type emitGoScope struct {
	knownDefns  map[string]bool
	localParams map[string]bool
	// retType — the enclosing defn's own resolved Go return type. Only actually needed inside
	// the "if" case below (see its own doc comment for the real, genuine bug this exists to
	// fix), but threaded through the whole scope rather than passed as a separate parameter, the
	// same way emitCScope carries context uniformly rather than growing a special-cased
	// signature for one caller.
	retType string
}

func emitGoExpr(expr *Node, scope *emitGoScope) (string, error) {
	if expr == nil {
		return "", errors.New("emit_go: null expression")
	}
	if expr.Type == NodeNumber {
		return expr.Text, nil
	}
	if expr.Type == NodeSymbol {
		// Real, genuine gap found live (PARENA/stdlib/datetime.prn's own is-leap-year?):
		// bare `true`/`false` Bool literals have no handling at all here, so any real .prn
		// function returning one hits "unknown identifier" rather than the correct Go literal.
		// Checked before the local-param/known-defn lookup since neither a param nor a defn is
		// ever actually named "true"/"false" in real PARENA source.
		if expr.Text == "true" || expr.Text == "false" {
			return expr.Text, nil
		}
		if !scope.localParams[expr.Text] && !scope.knownDefns[expr.Text] {
			return "", errors.New("emit_go: unknown identifier '" + expr.Text + "' at line " + itoa(expr.Line))
		}
		return mangleGoLocal(expr.Text, scope), nil
	}
	if expr.Type != NodeList || len(expr.Children) == 0 || expr.Children[0].Type != NodeSymbol {
		return "", errors.New("emit_go: unsupported expression form (v0 only understands numbers, symbols, binops, if, and calls)")
	}
	head := expr.Children[0].Text

	if head == "if" {
		if len(expr.Children) != 4 {
			return "", errors.New("emit_go: if requires exactly (if cond then else)")
		}
		cond, err := emitGoExpr(expr.Children[1], scope)
		if err != nil {
			return "", err
		}
		thenE, err := emitGoExpr(expr.Children[2], scope)
		if err != nil {
			return "", err
		}
		elseE, err := emitGoExpr(expr.Children[3], scope)
		if err != nil {
			return "", err
		}
		// Real, necessary structural departure from emit_c.go's own C ternary: Go has no `?:`
		// expression operator at all (a real, well-known, deliberate Go language design choice).
		// The real, standard Go idiom for an inline conditional expression is an immediately-
		// invoked func literal -- real, valid Go, and still exactly ONE expression from the
		// caller's own point of view (matching this file's own "one real expression" body scope;
		// the func literal's body is plumbing this emitter generates, not new emittable surface a
		// v0 .prn author can reach). No allocation: a zero-capture func literal that Go's own
		// compiler inlines/stack-allocates in the overwhelming common case (real, standard Go
		// compiler behavior, not a claim unique to this emitter).
		//
		// Real, genuine bug found and fixed here testing this for real (a nested `clamp01`
		// probe, two `if` levels deep): a bare branch value returned as-is into the func
		// literal's own `any`-typed return gets boxed with Go's own DEFAULT type for that
		// literal/expression, not this defn's actual declared return type -- for I32 specifically
		// (Go's own untyped-constant default for a bare integer literal is `int`, not `int32`),
		// that silently produced an `any` boxing the wrong concrete type, and the outer
		// `.(RetType)` assertion at the call site panicked at runtime: "interface conversion:
		// interface {} is int, not int32". F64/Bool/String don't hit this (Go's own literal
		// defaults for those already happen to coincide with this emitter's own resolved types),
		// but I32 does, every time, unconditionally -- not an edge case. Fixed by wrapping EVERY
		// branch value in an explicit `RetType(...)` conversion before it's boxed, so the `any`
		// always carries the exact right dynamic type regardless of whether the branch is a
		// literal, a param, a binop, or a nested call (the conversion is a real no-op for a value
		// that's already the right type, and load-bearing for a bare literal that isn't).
		//
		// Real, second genuine bug found and fixed here, same probe, one level deeper (a nested
		// `if` INSIDE another `if`'s own branch): the wrapping above -- `RetType(...)`, a Go type
		// CONVERSION -- is only valid when its argument already has a concrete static type. A
		// nested "if" case's own returned expression is `func() any {...}()`, whose static type
		// IS `any` -- converting an interface value with `T(x)` syntax is a compile error ("need
		// type assertion"), not the type ASSERTION `x.(T)` a concrete result out of an `any`
		// actually needs. Fixed at the source, not the call site: THIS case appends its own
		// `.(RetType)` assertion to ITS OWN result before returning it, so every value
		// emitGoExpr ever hands back to a caller already has a real, concrete, correctly-typed
		// Go expression -- literal, symbol, binop, call, AND nested-if all look identical to
		// whatever wraps them, and the uniform `RetType(...)` conversion above is always the
		// right one to apply, no special-casing needed for "was this branch itself an if."
		return "func() any { if " + cond + " { return " + scope.retType + "(" + thenE + ") }; return " +
			scope.retType + "(" + elseE + ") }().(" + scope.retType + ")", nil
	}

	// `let`/`do` -- real, new capability (kanban priority-queue card 1199/9988: "iterate on
	// project burrow... so that parena gets transformed into idiomatic go" / "emily for business
	// CLI written in GO with BURROW"). Real, decisive finding that motivated this: v0's own
	// "one-expression body" scope meant NO real function could declare a local variable at all
	// -- the single largest real gap blocking any real multi-statement logic (a CLI's own
	// argument parsing, string building, sequential setup) from ever reaching this target.
	//
	// Real design choice, matching the exact same real trick the "if" case above already uses:
	// emitted as an immediately-invoked func literal, boxing the final result through `any` and
	// asserting it back to the ENCLOSING defn's own declared return type (`scope.retType`) --
	// this emitter still does no real type inference of its own, so it leans on the same
	// already-proven mechanism rather than inventing a second one. This keeps `let`/`do` "one
	// real expression" from the caller's own point of view, same as `if`, and means a `let`
	// nested inside an `if` branch (or vice versa) composes correctly for free -- both cases
	// always hand back a concrete, correctly-typed Go expression to whatever wraps them.
	//
	// Real, deliberate scope, narrower than PARENA's own `let`: bindings are evaluated into a
	// CLONED local-params map (not scope's own), so a `let`'s bindings never leak into sibling
	// expressions outside it -- real, correct lexical scoping for this one level, not yet
	// verified against deeper real-world shadowing edge cases (e.g. a binding shadowing an outer
	// param of the same name is untested; PARENA's own real style avoids that anyway).
	if head == "let" {
		if len(expr.Children) < 3 || expr.Children[1].Type != NodeVec {
			return "", errors.New("emit_go: let requires (let [name expr ...] body...)")
		}
		bindings := expr.Children[1]
		if len(bindings.Children)%2 != 0 {
			return "", errors.New("emit_go: let bindings vector must have an even number of elements (name expr pairs)")
		}
		childParams := make(map[string]bool, len(scope.localParams)+len(bindings.Children)/2)
		for k, v := range scope.localParams {
			childParams[k] = v
		}
		childScope := &emitGoScope{knownDefns: scope.knownDefns, localParams: childParams, retType: scope.retType}

		var stmts []string
		for i := 0; i+1 < len(bindings.Children); i += 2 {
			nameNode := bindings.Children[i]
			valNode := bindings.Children[i+1]
			if nameNode.Type != NodeSymbol {
				return "", errors.New("emit_go: let binding name must be a plain symbol")
			}
			valE, err := emitGoExpr(valNode, childScope)
			if err != nil {
				return "", err
			}
			childScope.localParams[nameNode.Text] = true
			stmts = append(stmts, mangleGoLocal(nameNode.Text, childScope)+" := "+valE)
		}
		bodyE, err := emitGoBody(expr.Children[2:], childScope)
		if err != nil {
			return "", err
		}
		stmts = append(stmts, bodyE...)
		return wrapGoStmtsAsExpr(stmts, scope.retType), nil
	}

	if head == "do" {
		if len(expr.Children) < 2 {
			return "", errors.New("emit_go: do requires at least one body expression")
		}
		stmts, err := emitGoBody(expr.Children[1:], scope)
		if err != nil {
			return "", err
		}
		return wrapGoStmtsAsExpr(stmts, scope.retType), nil
	}

	// `(not x)` -- real, same unary-operator gap emit_c.go's own real fix just closed, same day,
	// same real trigger (stdlib/k8s/operator.prn's own `(if (not exists) ...)`); Go's own `!`
	// negation operator is the direct equivalent.
	if head == "not" {
		if len(expr.Children) != 2 {
			return "", errors.New("emit_go: not requires exactly 1 operand")
		}
		inner, err := emitGoExpr(expr.Children[1], scope)
		if err != nil {
			return "", err
		}
		return "(!(" + inner + "))", nil
	}

	if goOp, ok := goBinopTable[head]; ok {
		if len(expr.Children) != 3 {
			return "", errors.New("emit_go: binary operator requires exactly 2 operands (v0 has no variadic +/and/or)")
		}
		lhs, err := emitGoExpr(expr.Children[1], scope)
		if err != nil {
			return "", err
		}
		rhs, err := emitGoExpr(expr.Children[2], scope)
		if err != nil {
			return "", err
		}
		return "(" + lhs + " " + goOp + " " + rhs + ")", nil
	}

	// get-field — real, new capability, the read half of defstruct support (see
	// emitGoDefstruct's own doc comment for why construction isn't emitted here): PARENA's real
	// `(get-field record :field)` form lowers to Go's own real, plain `record.Field` dot access.
	// Real shape check matches the parser's own real AST for this form directly (confirmed via
	// `burrow parse` on a real probe before writing this, not guessed): exactly 3 children,
	// the record expression, then a `NodeKeyword` (`:field`, parsed with its leading colon
	// still part of `.Text` -- stripped here).
	if head == "get-field" {
		if len(expr.Children) != 3 || expr.Children[2].Type != NodeKeyword {
			return "", errors.New("emit_go: get-field requires exactly (get-field record :field)")
		}
		recordE, err := emitGoExpr(expr.Children[1], scope)
		if err != nil {
			return "", err
		}
		fieldName := strings.TrimPrefix(expr.Children[2].Text, ":")
		return "(" + recordE + ")." + mangleGo(fieldName), nil
	}

	// Otherwise: a real call to another top-level defn in the same generated package -- real,
	// honest validation first (see emitGoScope's own doc comment).
	if !scope.knownDefns[head] {
		return "", errors.New("emit_go: unknown identifier '" + head + "' at line " + itoa(expr.Children[0].Line))
	}
	var b strings.Builder
	b.WriteString(mangleGo(head))
	b.WriteString("(")
	for i := 1; i < len(expr.Children); i++ {
		if i > 1 {
			b.WriteString(", ")
		}
		arg, err := emitGoExpr(expr.Children[i], scope)
		if err != nil {
			return "", err
		}
		b.WriteString(arg)
	}
	b.WriteString(")")
	return b.String(), nil
}

// emitGoBody — real, shared sequencing logic for both `let` and `do`: every expression but the
// last runs as a real Go statement for effect only (its value discarded via `_ =`, matching the
// same real PARENA semantics `do`/a `let`'s own multi-expression body already have -- side
// effects, not accumulation), and the final expression's own emitted Go becomes the sequence's
// real result statement (a bare `return`-less expression here; wrapGoStmtsAsExpr below is what
// actually turns it into a real `return`).
func emitGoBody(bodyExprs []*Node, scope *emitGoScope) ([]string, error) {
	if len(bodyExprs) == 0 {
		return nil, errors.New("emit_go: let/do requires at least one body expression")
	}
	var stmts []string
	for i := 0; i < len(bodyExprs)-1; i++ {
		e, err := emitGoExpr(bodyExprs[i], scope)
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, "_ = "+e)
	}
	resultE, err := emitGoExpr(bodyExprs[len(bodyExprs)-1], scope)
	if err != nil {
		return nil, err
	}
	stmts = append(stmts, "__RESULT__ "+resultE)
	return stmts, nil
}

// wrapGoStmtsAsExpr — turns a real statement list (the last one tagged "__RESULT__ <expr>" by
// emitGoBody above) into the exact same real immediately-invoked-func-literal-boxed-through-any
// shape the "if" case already established, so `let`/`do` compose with everything else (nested
// inside an `if` branch, or vice versa) without emitGoExpr's own caller ever needing to know the
// difference.
func wrapGoStmtsAsExpr(stmts []string, retType string) string {
	var b strings.Builder
	b.WriteString("func() any { ")
	for _, s := range stmts {
		if rest, ok := strings.CutPrefix(s, "__RESULT__ "); ok {
			b.WriteString("return " + retType + "(" + rest + ")")
		} else {
			b.WriteString(s)
			b.WriteString("; ")
		}
	}
	b.WriteString(" }().(" + retType + ")")
	return b.String()
}

// mangleGoLocal — a bare symbol reference inside a body is either a LOCAL param or let-binding
// (Go convention: lowercase, unexported -- matches every real, idiomatic Go function's own
// parameter naming) or a real, zero-arg sibling call (which
// must be exported PascalCase + `()`, since it's a real function reference, not a value).
func mangleGoLocal(name string, scope *emitGoScope) string {
	if scope.localParams[name] {
		return strings.ReplaceAll(name, "-", "_")
	}
	// A bare (non-call-position) reference to a known zero-arg defn -- real, valid PARENA (see
	// entry.prn's own real examples elsewhere in this monorepo calling a zero-arg sibling
	// bare-symbol-style is NOT how PARENA calls a zero-arg fn -- it's always `(fn-name)`, so this
	// branch is dead in practice for well-formed input, kept only so an unrecognized bare symbol
	// that happens to match a known defn name fails with the SAME real "unknown identifier" class
	// of error a param-shadowing mistake would, rather than a confusing Go compile error instead.
	return mangleGo(name) + "()"
}

// goFunc holds one real, emitted Go function's own full definition. Real, deliberate structural
// difference from emit_c.go's own emitCDefn: Go has no forward-declaration concept at all (a
// real, standard Go source file's own top-level function order is irrelevant to the Go compiler
// -- sibling calls in any order already just work), so there is no decl/def split to carry here.
// emitGoDefstruct — real, new capability: a Go `type Name struct { Field1 T1; ... }` for each
// real `(defstruct Name (field1 : T1) ...)` form. Real, deliberate scope decision, checked
// against what the real trigger (PARENA's own new `stdlib/k8s`/`stdlib/helm` packages) actually
// needs before building anything more general: construction (`{:field val}`) and pattern-style
// access are NOT emitted here -- every real function in those two packages only ever RECEIVES a
// struct as a parameter (via `get-field`), never constructs one internally; the real host that
// constructs one (a Go program calling into this generated package) does it with an ordinary Go
// composite literal against the real, exported struct type this function emits -- the same real
// split `k8s.prn`'s own C target already uses (a real C test harness constructs via
// `Deployment_new(...)`, the PARENA-emitted code itself only ever calls `get-field`). Exported
// (PascalCase) field names, matching the exported function-name convention this emitter already
// uses, so a real external Go host (DUNG) can actually construct/read these fields.
func emitGoDefstruct(ds *Node, knownStructs map[string]bool) (string, error) {
	if len(ds.Children) < 2 || ds.Children[1].Type != NodeSymbol {
		return "", errors.New("emit_go: defstruct: malformed struct definition")
	}
	typeName := mangleGo(ds.Children[1].Text)
	var fields strings.Builder
	for _, field := range ds.Children[2:] {
		if field.Type != NodeList || len(field.Children) != 3 || field.Children[0].Type != NodeSymbol ||
			field.Children[1].Type != NodeColon || field.Children[2].Type != NodeSymbol {
			return "", errors.New("emit_go: defstruct: unsupported field shape (v0 only understands plain " +
				"(name : I32|F64|Bool|String|StructType) fields -- no Arena/region annotations, no Vec fields)")
		}
		fType, err := resolveGoType(field.Children[2].Text, knownStructs)
		if err != nil {
			return "", err
		}
		fields.WriteString(mangleGo(field.Children[0].Text))
		fields.WriteString(" ")
		fields.WriteString(fType)
		fields.WriteString("\n")
	}
	return "type " + typeName + " struct {\n" + fields.String() + "}", nil
}

func emitGoDefn(defn *Node, knownDefns map[string]bool, knownStructs map[string]bool) (string, error) {
	if len(defn.Children) < 3 || defn.Children[1].Type != NodeSymbol || defn.Children[2].Type != NodeVec {
		return "", errors.New("emit_go: defn: malformed function definition")
	}
	fnName := mangleGo(defn.Children[1].Text)
	params := defn.Children[2]
	scope := &emitGoScope{knownDefns: knownDefns, localParams: map[string]bool{}}

	var paramList strings.Builder
	for i, param := range params.Children {
		if param.Type != NodeList || len(param.Children) != 3 || param.Children[0].Type != NodeSymbol ||
			param.Children[1].Type != NodeColon || param.Children[2].Type != NodeSymbol {
			return "", errors.New("emit_go: defn: unsupported parameter shape (v0 only understands plain " +
				"(name : I32|F64|Bool|String|StructType) params -- no Arena/region annotations)")
		}
		pType, err := resolveGoType(param.Children[2].Text, knownStructs)
		if err != nil {
			return "", err
		}
		if i > 0 {
			paramList.WriteString(", ")
		}
		// Real Go declaration order: "name Type" (matching TS's own "name: Type" shape more than
		// C/Java's "Type name" -- a real, plain, unavoidable Go syntax fact, not a stylistic
		// choice this emitter is making).
		paramList.WriteString(strings.ReplaceAll(param.Children[0].Text, "-", "_"))
		paramList.WriteString(" ")
		paramList.WriteString(pType)
		scope.localParams[param.Children[0].Text] = true
	}

	if len(defn.Children) != 6 || defn.Children[3].Type != NodeColon {
		return "", errors.New("emit_go: defn: expected (defn name [params] : RetType body) with exactly one body expression")
	}
	retType, err := resolveGoType(defn.Children[4].Text, knownStructs)
	if err != nil {
		return "", err
	}
	scope.retType = retType
	body, err := emitGoExpr(defn.Children[5], scope)
	if err != nil {
		return "", err
	}

	// No wrapping needed here: emitGoExpr's own "if" case now appends its own `.(RetType)`
	// assertion to its own result before returning (see that case's own second doc comment,
	// added after a real, genuine nested-if compile error) -- every value this function gets
	// back, "if"-bodied or not, is already a real, concrete, correctly-typed Go expression.

	// Real, necessary detail: a tab, not spaces, for the body indent -- gofmt's own real,
	// non-negotiable convention for Go source (confirmed the hard way: an early version of this
	// emitter used a 4-space indent and `gofmt -l` correctly flagged the output as not
	// gofmt-clean). Generated code that isn't gofmt-clean is a real, honest quality bug in this
	// emitter, not a cosmetic nit -- every other real generated source in this monorepo holds
	// itself to its own target language's own formatter/linter bar (gcc -Wall clean for C).
	return "func " + fnName + "(" + paramList.String() + ") " + retType + " {\n\treturn " + body + "\n}", nil
}

// EmitGo — real, top-level entry point: walks every top-level (defn ...) form in `program` and
// produces one real, complete, `gofmt`-clean Go source file, package `burrowgen` (a real,
// intentionally neutral name -- the real host importing this package chooses its own local import
// alias, matching how a real generated .c file doesn't get to pick its own #include guard name
// either). No forward-declaration pass needed (see emitGoDefn's own doc comment) -- real,
// structurally simpler single pass, unlike EmitC's own real two-pass decl/def split.
func EmitGo(program *Node) (string, error) {
	knownDefns := map[string]bool{}
	knownStructs := map[string]bool{}
	for _, form := range program.Children {
		if isCallNamed(form, "defn") && len(form.Children) >= 2 && form.Children[1].Type == NodeSymbol {
			knownDefns[form.Children[1].Text] = true
		}
		if isCallNamed(form, "defstruct") && len(form.Children) >= 2 && form.Children[1].Type == NodeSymbol {
			knownStructs[form.Children[1].Text] = true
		}
	}

	var defs []string
	for _, form := range program.Children {
		if isCallNamed(form, "module") || isCallNamed(form, "export") || isCallNamed(form, "import") {
			continue
		}
		if isCallNamed(form, "defstruct") {
			def, err := emitGoDefstruct(form, knownStructs)
			if err != nil {
				return "", err
			}
			defs = append(defs, def)
			continue
		}
		if isCallNamed(form, "defn") {
			def, err := emitGoDefn(form, knownDefns, knownStructs)
			if err != nil {
				return "", err
			}
			defs = append(defs, def)
			continue
		}
		return "", errors.New("emit_go: unsupported top-level form (v0 only understands defn, defstruct, module, export, import)")
	}

	var out strings.Builder
	out.WriteString("// Code generated by burrow build (Go target) -- VS0-for-Go v0, DO NOT EDIT.\n\n")
	out.WriteString("package burrowgen\n\n")
	out.WriteString(strings.Join(defs, "\n\n"))
	out.WriteString("\n")

	// Real, necessary final pass, not optional polish: hand-tracking indentation through nested
	// if-as-func-literal expressions (emitGoExpr's own "if" case) to stay gofmt-clean by
	// construction would mean re-implementing gofmt's own canonical control-flow line-breaking
	// rules inside this emitter -- found the hard way, empirically, testing a real nested-if
	// probe (`clamp01`, two levels deep): the single-line `if cond { return X }` shape this
	// emitter's own string-building naturally produces is syntactically valid Go but NOT
	// gofmt-clean (gofmt always expands a braced if-with-a-statement onto multiple lines).
	// go/format.Source is the exact same real formatter `gofmt` itself calls -- running the
	// fully-assembled source through it here guarantees gofmt-clean output unconditionally, for
	// any real nesting depth, without this emitter tracking indentation by hand at all. A failure
	// here means this emitter produced syntactically INVALID Go -- a real, genuine bug in the
	// emitter itself, surfaced as a real, honest error rather than handed to the caller broken.
	formatted, ferr := format.Source([]byte(out.String()))
	if ferr != nil {
		return "", errors.New("emit_go: internal error, emitted source failed to gofmt (this is a real bug in emit_go.go, not the input .prn): " + ferr.Error())
	}
	return string(formatted), nil
}
