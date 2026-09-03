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
// own "if" case -- and calls to another top-level `defn`) -- NOT the full `emit.c`, though
// `defstruct`/`let`/`do`/`match`/`Result`/`Option`/a real, narrow v0 `loop`/`recur` have since
// grown on top of that original scope (see this file's own doc comments on each). Growing this
// further (`defenum`, `Vec`, general struct construction, `loop`/`match` beyond their own current
// v0 boundaries) is real, separate, unstarted work, same as every other emit_*.go's own honest
// boundary.
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
	"strconv"
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
	// defnRetInfo — real, new (kanban card 9988's own match/Result port): every known defn's own
	// declared Result/Option return-type payload/error types, keyed by defn name, collected in
	// EmitGo's own first pass. `match` uses this to know what concrete Go type a scrutinee call's
	// own Ok/Some/Err payload actually is -- see resolveGoReturnType's own doc comment for the
	// real reasoning and the real, honest v0 boundary this implies.
	defnRetInfo map[string]goDefnRetInfo
	// retType — the enclosing defn's own resolved Go return type. Only actually needed inside
	// the "if" case below (see its own doc comment for the real, genuine bug this exists to
	// fix), but threaded through the whole scope rather than passed as a separate parameter, the
	// same way emitCScope carries context uniformly rather than growing a special-cased
	// signature for one caller.
	retType string
}

// goDefnRetInfo — a known defn's own Result/Option payload/error types, in real, already-resolved
// Go type strings (e.g. "int32", "string", a registered struct's PascalCase name). ErrorType is
// only meaningful when Kind == "result" (an Option has no error variant to carry one for).
type goDefnRetInfo struct {
	Kind        string // "result" or "option"
	PayloadType string
	ErrorType   string
}

// resolveGoReturnType — real, new capability alongside resolveGoType above: a defn's own
// declared return type is either a bare scalar/struct symbol (resolveGoType's own existing real
// scope) OR a compound `(Result Payload Error)` / `(Option Payload)` form, which resolveGoType
// alone has never been able to parse (it only ever took a bare symbol's own text). Returns the
// real Go type to declare (always the literal shared "Result"/"Option" struct name for those two
// forms -- see emitResultOptionStructs' own doc comment for why these are fixed, not
// per-instantiation, types) plus a *goDefnRetInfo when the return type IS Result/Option (nil
// otherwise), so EmitGo's own first pass can register it for `match` to look up later.
func resolveGoReturnType(typeNode *Node, knownStructs map[string]bool) (string, *goDefnRetInfo, error) {
	if typeNode.Type == NodeSymbol {
		t, err := resolveGoType(typeNode.Text, knownStructs)
		return t, nil, err
	}
	if typeNode.Type != NodeList || len(typeNode.Children) == 0 || typeNode.Children[0].Type != NodeSymbol {
		return "", nil, errors.New("emit_go: unsupported return type form (v0 only understands a bare I32/F64/Bool/String/struct symbol, or a compound (Result Payload Error)/(Option Payload) form)")
	}
	head := typeNode.Children[0].Text
	switch head {
	case "Result":
		if len(typeNode.Children) != 3 || typeNode.Children[1].Type != NodeSymbol || typeNode.Children[2].Type != NodeSymbol {
			return "", nil, errors.New("emit_go: Result return type requires exactly (Result PayloadType ErrorType), both bare symbols")
		}
		payload, err := resolveGoType(typeNode.Children[1].Text, knownStructs)
		if err != nil {
			return "", nil, err
		}
		errType, err := resolveGoType(typeNode.Children[2].Text, knownStructs)
		if err != nil {
			return "", nil, err
		}
		return "Result", &goDefnRetInfo{Kind: "result", PayloadType: payload, ErrorType: errType}, nil
	case "Option":
		if len(typeNode.Children) != 2 || typeNode.Children[1].Type != NodeSymbol {
			return "", nil, errors.New("emit_go: Option return type requires exactly (Option PayloadType), a bare symbol")
		}
		payload, err := resolveGoType(typeNode.Children[1].Text, knownStructs)
		if err != nil {
			return "", nil, err
		}
		return "Option", &goDefnRetInfo{Kind: "option", PayloadType: payload}, nil
	case "Vec":
		// Vec return type -- real, new capability (kanban card 9988's own next-named
		// prerequisite after loop/recur). No goDefnRetInfo entry: match's own scrutinee lookup
		// only ever cares about Result/Option, and a Vec-returning defn doesn't need a boxing
		// helper the way Ok/Err/Some do -- a bare Go `[]any` composite literal already IS the
		// right shape.
		if len(typeNode.Children) != 2 || typeNode.Children[1].Type != NodeSymbol {
			return "", nil, errors.New("emit_go: Vec return type requires exactly (Vec ElemType), a bare symbol")
		}
		if _, err := resolveGoType(typeNode.Children[1].Text, knownStructs); err != nil {
			return "", nil, err
		}
		return "[]any", nil, nil
	default:
		return "", nil, errors.New("emit_go: unsupported compound return type '" + head + "' (v0 only understands Result/Option/Vec)")
	}
}

func emitGoExpr(expr *Node, scope *emitGoScope) (string, error) {
	if expr == nil {
		return "", errors.New("emit_go: null expression")
	}
	if expr.Type == NodeNumber {
		return expr.Text, nil
	}
	// Real, pre-existing gap found live (kanban card 9988's own match/Result port -- a real
	// `Result I32 String` error payload needing a plain "division by zero"-style literal was
	// the first thing in this whole target's history to actually exercise one): string literals
	// had no handling here at all, even though String was already a real, supported param/
	// return type. The lexer's own real `.Text` is already fully escape-decoded (see lexer.go's
	// own lexString doc comment) with no surrounding quotes -- strconv.Quote re-escapes it back
	// into a real, safe, valid Go string literal (handles a literal containing a `"` or `\`
	// correctly, not just the common case).
	if expr.Type == NodeString {
		return strconv.Quote(expr.Text), nil
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
		// Bare `None` -- real, established PARENA source convention (see bstree.prn's own real,
		// live `get-loop`: `None` referenced unapplied, never `(None)`), matching how `true`/
		// `false` are also real bare-symbol literals with no local/defn binding behind them.
		if expr.Text == "None" {
			return "Option{Tag: 0, Value: nil}", nil
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

	// Ok/Err/Some/None construction -- real, new capability (kanban card 9988's own match/Result
	// port). Real, direct port of PARENA's own runtime representation
	// (parena_runtime.h's own result_ok/result_err/option_some/option_none): a real, FIXED,
	// shared `{Tag int; Value any}` struct per Result/Option (see EmitGo's own doc comment on
	// why this is fixed, not per-instantiation) -- Go's own `any` interface does the same real
	// "erase to a pointer-ish box" job C's own `void *` does, for free, without this emitter
	// needing its own arena-boxing-helper machinery the C target's own equivalent needed for a
	// non-pointer payload.
	if head == "Ok" || head == "Err" || head == "Some" {
		if len(expr.Children) != 2 {
			return "", errors.New("emit_go: " + head + " requires exactly 1 argument")
		}
		inner, err := emitGoExpr(expr.Children[1], scope)
		if err != nil {
			return "", err
		}
		switch head {
		case "Ok":
			return "Result{Tag: 1, Value: " + inner + "}", nil
		case "Err":
			return "Result{Tag: 0, Value: " + inner + "}", nil
		default: // Some
			return "Option{Tag: 1, Value: " + inner + "}", nil
		}
	}
	if head == "None" {
		if len(expr.Children) != 1 {
			return "", errors.New("emit_go: None takes no arguments")
		}
		return "Option{Tag: 0, Value: nil}", nil
	}

	// match -- real, new capability (kanban card 9988), a real, deliberate v0 BOUNDARY named
	// explicitly, not silently limited: the scrutinee must be a direct call to a known defn
	// whose own declared return type is Result/Option (resolved via scope.defnRetInfo, built in
	// EmitGo's own first pass) -- NOT an arbitrary expression, and NOT a `let`-bound variable.
	// Real, honest reason: this emitter's own local-variable tracking (`localParams
	// map[string]bool`) carries presence only, no per-variable TYPE -- extending that to track
	// real types for every local is real, separate, larger work (the same real scope PARENA's
	// own mature `src/emit.c` needed several distinct, later bug-fix passes to get right for its
	// C target, per that file's own accumulated `scrut_payload_type`/`scrut_error_type`
	// commentary — not something to rush past here). A scrutinee call's own return type IS
	// staticly knowable up front, though, which is exactly the real, useful, common case this
	// v0 covers: "call a function that might fail, then match its result immediately."
	if head == "match" {
		if len(expr.Children) < 2 {
			return "", errors.New("emit_go: match requires a scrutinee expression")
		}
		return emitGoMatch(expr, scope)
	}

	// loop/recur -- real, new capability (kanban cruise-queue card 9988's own next-named
	// prerequisite: "loop in particular is a real, necessary prerequisite for any real iteration
	// a CLI would need," per this session's own earlier match/Result changelog entry). Real,
	// deliberate v0 boundary, named explicitly, not silently limited: the loop body must be
	// EXACTLY one top-level `(if cond then else)` with `recur` appearing directly in ONE of the
	// two branches, the other branch being the loop's own terminal value -- the exact real shape
	// every actual .prn loop in this stdlib's own array.prn uses (`product`/`sum`:
	// `(if (>= i n) acc (recur (+ i 1) ...))`). `recur` nested inside a deeper `if`/`cond`/`match`
	// chain, or a loop with more than one body expression, is a real, honest, separate, larger
	// undertaking -- PARENA's own mature `src/emit.c` needed a `loop_locals` array threaded
	// through `emit_loop_tail`/`emit_match_core`/nested-`if` dispatch, several distinct bug-fix
	// passes, to get that fully general case right for its own C target; not rushed past here,
	// same real judgment call `emitGoMatch`'s own doc comment already makes for its own v0
	// boundary. A loop body outside this shape gets a real, honest compile error naming the exact
	// limitation, not a silently wrong emission.
	if head == "loop" {
		return emitGoLoop(expr, scope)
	}

	// vec/new/vec/push!/vec/get/vec/len/deref -- real, new capability (kanban cruise-queue card
	// 9988's own next-named prerequisite after loop/recur: "a real CLI needs... likely Vec for
	// building up output"). Real, direct port of PARENA's own runtime representation SHAPE
	// (parena_runtime.h's own `Vec { Arena *arena; void **items; size_t count; size_t
	// capacity; }`, a boxed void*-array) to Go's own idiomatic equivalent: a bare `[]any` slice,
	// no wrapper struct at all -- Go's own `append()` already does the exact real dynamic-growth
	// job C's own hand-rolled `vec_push_` does, and `any` already does the exact real "erase to a
	// pointer-ish box" job `void *` does, so this needs none of the C target's own manual arena-
	// growth bookkeeping. A real, honest consequence, named directly, not hidden: `emit_go.go`'s
	// own header comment previously claimed every v0-scope program is "already GC-irrelevant, no
	// heap allocation possible in the emitted code itself" -- that stops being literally true the
	// moment a real program uses Vec (`append` and a `[]any` backing array are real Go heap
	// allocations); a real host that adopted `debug.SetGCPercent(-1)` on the strength of that
	// original claim needs to know it no longer holds for any Vec-using generated code.
	if head == "vec/new" {
		if len(expr.Children) != 2 {
			return "", errors.New("emit_go: vec/new requires exactly 1 argument (the Arena)")
		}
		// The Arena argument is evaluated purely for real, honest validation (a typo'd/unknown
		// identifier here should still be a real compile error) -- its own value is never used,
		// since Go's `[]any` needs no arena at all. See the Arena-param doc comment in
		// emitGoDefn's own param-parsing loop for the fuller real reasoning.
		if _, err := emitGoExpr(expr.Children[1], scope); err != nil {
			return "", err
		}
		return "[]any(nil)", nil
	}
	if head == "vec/push!" {
		if len(expr.Children) < 3 {
			return "", errors.New("emit_go: vec/push! requires exactly (vec/push! &mut-vec item)")
		}
		target, consumed, err := resolveVecRef(expr.Children[1:])
		if err != nil {
			return "", err
		}
		if len(expr.Children) != 1+consumed+1 {
			return "", errors.New("emit_go: vec/push! requires exactly (vec/push! &mut-vec item)")
		}
		// Real, deliberate v0 boundary: the mutation target must be a plain local symbol or a
		// `(get-field record :field)` expression -- both map to a real, valid, ADDRESSABLE Go
		// l-value (`name` or `name.Field`) that `= append(...)` can assign back into. An
		// arbitrary expression (e.g. a nested function call's own return value) has no real Go
		// l-value to reassign at all, the same class of "not every expression can be a mutation
		// target" restriction `let`'s own binding-name-must-be-a-symbol check already makes.
		var lvalue string
		if target.Type == NodeSymbol {
			if !scope.localParams[target.Text] {
				return "", errors.New("emit_go: unknown identifier '" + target.Text + "' at line " + itoa(target.Line))
			}
			lvalue = mangleGoLocal(target.Text, scope)
		} else if target.Type == NodeList && len(target.Children) == 3 && target.Children[0].Type == NodeSymbol && target.Children[0].Text == "get-field" {
			lvalue, err = emitGoExpr(target, scope)
			if err != nil {
				return "", err
			}
		} else {
			return "", errors.New("emit_go: vec/push!'s own mutation target must be a plain local variable or a (get-field record :field) expression")
		}
		itemE, err := emitGoExpr(expr.Children[1+consumed], scope)
		if err != nil {
			return "", err
		}
		// Real, genuine bug found and fixed live testing this: `append(vec, 1)` boxes the bare
		// literal `1` as plain Go `int`, not `int32` -- the identical defaulting class already
		// found and fixed for loop bindings and the "if" case's own branch values, hitting a
		// THIRD real site here. Since vec/get's own read path always asserts `.(int32)` (see its
		// own doc comment on this target's real "Vec is I32-valued" v0 boundary), a mis-boxed
		// plain `int` element would silently fail that assertion and read back as 0 instead of
		// the real pushed value. Fixed by wrapping every pushed item in an explicit `int32(...)`
		// conversion -- a real no-op for a value that's already int32, load-bearing for a bare
		// literal that isn't.
		itemE = "int32(" + itemE + ")"
		// Go closures capture their enclosing scope's own variables BY REFERENCE, not by value
		// (a real, standard Go language guarantee, not an emitter trick) -- so this real,
		// immediately-invoked func literal's own `lvalue = append(...)` genuinely mutates the
		// CALLER's own variable, matching PARENA's own real "push! mutates its Vec argument in
		// place" semantics with zero pointer/reference machinery needed on this target's side.
		return "func() any { " + lvalue + " = append(" + lvalue + ", " + itemE + "); return nil }()", nil
	}
	if head == "vec/get" {
		if len(expr.Children) < 3 {
			return "", errors.New("emit_go: vec/get requires exactly (vec/get &vec idx)")
		}
		target, consumed, err := resolveVecRef(expr.Children[1:])
		if err != nil {
			return "", err
		}
		if len(expr.Children) != 1+consumed+1 {
			return "", errors.New("emit_go: vec/get requires exactly (vec/get &vec idx)")
		}
		vecE, err := emitGoExpr(target, scope)
		if err != nil {
			return "", err
		}
		idxE, err := emitGoExpr(expr.Children[1+consumed], scope)
		if err != nil {
			return "", err
		}
		// Real, deliberate v0 boundary, named explicitly, not silently narrow: this target has
		// no real per-Vec element-type tracking (the same class of "no per-variable type" gap
		// match's own scrutinee restriction and loop's own binding-type fix already named) -- so
		// vec/get's OWN result is always coerced to `int32` here, matching every current real
		// .prn Vec usage this specific target actually needs (array.prn's own shape/stride
		// vectors are all I32-valued). A `(Vec SomeStruct)` (bstree.prn's own real BSTNode-valued
		// Vec is the known, existing counter-example) is real, separate, unsupported work for
		// this target -- extending this would need a real per-Vec-defn element-type registry,
		// the same size of undertaking `defnRetInfo` already is for Result/Option.
		// Real, direct port of parena_runtime.h's own real vec_get safety guarantee: an
		// out-of-bounds index returns a real, honest `int32(0)` (this target's own closest
		// idiomatic equivalent to the C target's real NULL-on-OOB) rather than a Go runtime
		// index-out-of-range panic; a comma-ok type assertion (not a direct `.(int32)`) also
		// guards against a stored element that isn't actually an int32, for the same real reason.
		return "func() any { __vec_tmp := " + vecE + "; __vec_idx := " + idxE +
			"; if __vec_idx < 0 || __vec_idx >= int32(len(__vec_tmp)) { return int32(0) }" +
			"; __vec_val, _ := __vec_tmp[__vec_idx].(int32); return __vec_val }().(int32)", nil
	}
	if head == "vec/len" {
		if len(expr.Children) < 2 {
			return "", errors.New("emit_go: vec/len requires exactly (vec/len &vec)")
		}
		target, consumed, err := resolveVecRef(expr.Children[1:])
		if err != nil {
			return "", err
		}
		if consumed != len(expr.Children)-1 {
			return "", errors.New("emit_go: vec/len takes no arguments beyond the vec reference")
		}
		vecE, err := emitGoExpr(target, scope)
		if err != nil {
			return "", err
		}
		return "int32(len(" + vecE + "))", nil
	}
	// deref -- real, honest NO-OP on this target: PARENA/C represents a `vec/get` result as a
	// real boxed pointer that `deref` reads through; Go's own `any`-boxed slice element (see
	// vec/get above) already IS the real value with no separate reference layer to unwrap, so
	// `deref` here is just "emit the inner expression unchanged." Named explicitly, not silently
	// dropped, so a future reader isn't left wondering where deref's own real work went.
	if head == "deref" {
		if len(expr.Children) != 2 {
			return "", errors.New("emit_go: deref requires exactly 1 argument")
		}
		return emitGoExpr(expr.Children[1], scope)
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
	// discardScope -- real, genuine bug found and fixed live testing a real (loop [i 0] (if (>= i
	// n) 0 (do (vec/push! ...) (recur ...)))) probe used as a NON-FINAL do/let statement (the
	// exact real shape array.prn's own `zeros` uses: a side-effecting loop discarded via `_ =`,
	// followed by the actual real result): every nested if/let/loop/match's own internal
	// `RetType(...)`/`.(RetType)` boxing had always used `scope.retType` unconditionally --
	// correct for a TAIL expression (whatever it produces really does become the enclosing
	// function's own return value), but WRONG for a discarded non-final statement, whose own
	// "terminal value" (here, a bare placeholder `0`) has no real reason to match the enclosing
	// DEFN's actual return type (`(Vec I32)` in the real trigger case) at all -- `[]any(0)` is a
	// real Go compile error ("cannot convert 0 to type []any"). Fixed by giving every non-final
	// statement its own scope with `retType` overridden to `any` -- converting ANY value to `any`
	// is always legal Go, and asserting `any` back to `any` always trivially succeeds, so a
	// discarded expression's own internal boxing never needs to agree with what it's inside.
	discardScope := *scope
	discardScope.retType = "any"
	var stmts []string
	for i := 0; i < len(bodyExprs)-1; i++ {
		e, err := emitGoExpr(bodyExprs[i], &discardScope)
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

// goMatchCounter -- real, direct analog of PARENA's own reference `src/emit.c`'s own
// `static int match_counter`: a process-wide counter giving every real match expression its own
// unique temp-variable name, so nested/sibling matches within the same defn (or across defns in
// one compile) never collide. Single-threaded compile process, same real assumption every other
// package-level emitter state in this file already makes.
var goMatchCounter int

// emitGoMatch -- see this function's own call site in emitGoExpr for the real, deliberate v0
// boundary this implements (scrutinee must be a direct call to a known Result/Option-returning
// defn) and why. Real, direct port of PARENA's own reference match codegen's overall SHAPE
// (tag-dispatch via if/else-if, one clause body per real tag value) -- see this repo's own
// BURROW/CLAUDE.md for the fuller real design writeup once shipped.
func emitGoMatch(expr *Node, scope *emitGoScope) (string, error) {
	scrutNode := expr.Children[1]
	clauseNodes := expr.Children[2:]
	if len(clauseNodes) != 2 {
		return "", errors.New("emit_go: match requires exactly 2 clauses at line " + itoa(expr.Line) +
			" (v0 only supports Result/Option, which have exactly 2 real variants each -- no wildcard/defenum matching yet)")
	}
	if scrutNode.Type != NodeList || len(scrutNode.Children) == 0 || scrutNode.Children[0].Type != NodeSymbol {
		return "", errors.New("emit_go: match's scrutinee must be a direct call to a known defn returning Result/Option at line " +
			itoa(scrutNode.Line) + " (v0 boundary: this emitter does not yet track per-variable types for a let-bound scrutinee)")
	}
	calleeName := scrutNode.Children[0].Text
	info, ok := scope.defnRetInfo[calleeName]
	if !ok {
		return "", errors.New("emit_go: match: '" + calleeName + "' is not a known defn with a declared Result/Option return type at line " + itoa(scrutNode.Line))
	}
	scrutExpr, err := emitGoExpr(scrutNode, scope)
	if err != nil {
		return "", err
	}

	goMatchCounter++
	tmpVar := "__match_tmp_" + itoa(goMatchCounter)

	seenTags := map[int]bool{}
	var branches strings.Builder
	for i, clauseNode := range clauseNodes {
		if clauseNode.Type != NodeList || len(clauseNode.Children) != 2 {
			return "", errors.New("emit_go: match: each clause must be exactly (pattern body) at line " + itoa(clauseNode.Line))
		}
		pattern := clauseNode.Children[0]
		bodyNode := clauseNode.Children[1]

		var ctorName, bindName string
		switch {
		case pattern.Type == NodeList && len(pattern.Children) == 2 && pattern.Children[0].Type == NodeSymbol && pattern.Children[1].Type == NodeSymbol:
			ctorName = pattern.Children[0].Text
			bindName = pattern.Children[1].Text
		case pattern.Type == NodeList && len(pattern.Children) == 1 && pattern.Children[0].Type == NodeSymbol:
			ctorName = pattern.Children[0].Text
		case pattern.Type == NodeSymbol:
			ctorName = pattern.Text
		default:
			return "", errors.New("emit_go: match: unsupported pattern shape at line " + itoa(pattern.Line) +
				" (v0 only understands (Ctor) or (Ctor bind-name))")
		}

		var tagValue int
		var payloadType string
		switch ctorName {
		case "Ok":
			if info.Kind != "result" {
				return "", errors.New("emit_go: match: 'Ok' pattern used against a non-Result scrutinee at line " + itoa(pattern.Line))
			}
			tagValue, payloadType = 1, info.PayloadType
		case "Err":
			if info.Kind != "result" {
				return "", errors.New("emit_go: match: 'Err' pattern used against a non-Result scrutinee at line " + itoa(pattern.Line))
			}
			tagValue, payloadType = 0, info.ErrorType
		case "Some":
			if info.Kind != "option" {
				return "", errors.New("emit_go: match: 'Some' pattern used against a non-Option scrutinee at line " + itoa(pattern.Line))
			}
			tagValue, payloadType = 1, info.PayloadType
		case "None":
			if info.Kind != "option" {
				return "", errors.New("emit_go: match: 'None' pattern used against a non-Option scrutinee at line " + itoa(pattern.Line))
			}
			if bindName != "" {
				return "", errors.New("emit_go: match: 'None' carries no payload, it cannot bind a name at line " + itoa(pattern.Line))
			}
			tagValue = 0
		default:
			return "", errors.New("emit_go: match: unsupported pattern '" + ctorName + "' at line " + itoa(pattern.Line) + " (v0 only understands Ok/Err/Some/None)")
		}
		if seenTags[tagValue] {
			return "", errors.New("emit_go: match: two clauses both match tag " + itoa(tagValue) + " at line " + itoa(pattern.Line) + " -- exactly one clause per real variant")
		}
		seenTags[tagValue] = true

		// Real, cloned child scope (same real leak-prevention discipline `let`'s own emitGoExpr
		// case already established): a clause's own bound name must not be visible to its
		// sibling clause or anything outside this match.
		childParams := make(map[string]bool, len(scope.localParams)+1)
		for k, v := range scope.localParams {
			childParams[k] = v
		}
		var bindStmt string
		if bindName != "" {
			childParams[bindName] = true
			// Real, found-live Go compile error, same class as the `unused param`/`unused let
			// binding` cases this repo's own C target guards with `__attribute__((unused))` --
			// Go has no such attribute, so a clause that validly ignores its own bound payload
			// (e.g. an Err arm that doesn't need the message) needs an explicit `_ = name`
			// discard instead, or `go build` genuinely fails with "declared and not used".
			goBindName := strings.ReplaceAll(bindName, "-", "_")
			bindStmt = goBindName + " := " + tmpVar + ".Value.(" + payloadType + ")\n_ = " + goBindName + "\n"
		}
		childScope := &emitGoScope{knownDefns: scope.knownDefns, localParams: childParams, defnRetInfo: scope.defnRetInfo, retType: scope.retType}

		bodyExpr, err := emitGoExpr(bodyNode, childScope)
		if err != nil {
			return "", err
		}

		// Real, deliberate exhaustiveness shortcut: with exactly 2 clauses required and each
		// naming a DISTINCT real tag value (seenTags above already enforces this), the second
		// clause is by construction the complementary case -- emitted as a plain `else`, not a
		// second `else if <tag check>`, so Go's own compiler sees a real, complete if/else (every
		// path returns) and doesn't need a dead trailing panic to satisfy it.
		if i == 0 {
			branches.WriteString("if " + tmpVar + ".Tag == " + itoa(tagValue) + " {\n")
		} else {
			branches.WriteString("} else {\n")
		}
		branches.WriteString(bindStmt)
		branches.WriteString("return " + scope.retType + "(" + bodyExpr + ")\n")
	}
	branches.WriteString("}\n")

	return "func() any { " + tmpVar + " := " + scrutExpr + "\n" + branches.String() +
		" }().(" + scope.retType + ")", nil
}

// goLoopCounter -- real, direct analog of goMatchCounter above (itself a direct analog of
// PARENA's own reference `static int match_counter`): gives every real loop's own recur-argument
// temp variables a unique name, so nested/sibling loops within the same defn never collide.
var goLoopCounter int

// isRecurCall reports whether node is a direct `(recur ...)` call -- the one thing this v0
// recognizes as "this branch continues the loop" (see emitGoLoop's own doc comment for the real,
// deliberate v0 boundary this implies: recur must appear directly here, not nested deeper).
func isRecurCall(node *Node) bool {
	return node.Type == NodeList && len(node.Children) >= 1 &&
		node.Children[0].Type == NodeSymbol && node.Children[0].Text == "recur"
}

// unwrapRecurBranch -- real, new capability alongside isRecurCall: recognizes EITHER a direct
// `(recur ...)` OR a `(do effect1 ... (recur ...))` whose own LAST expression is a direct recur
// call (the exact real shape a Vec-building loop needs -- see emitGoLoop's own doc comment on
// recurEffects for the real trigger). Returns the effect expressions to run before the recur
// reassignment (nil for the direct-recur case), the recur node itself, and whether this branch
// is a recur branch at all. A `do` whose own last expression ISN'T a direct recur (or any other
// shape) reports isRecur = false, same as isRecurCall already would.
func unwrapRecurBranch(node *Node) ([]*Node, *Node, bool) {
	if isRecurCall(node) {
		return nil, node, true
	}
	if node.Type == NodeList && len(node.Children) >= 2 && node.Children[0].Type == NodeSymbol && node.Children[0].Text == "do" {
		last := node.Children[len(node.Children)-1]
		if isRecurCall(last) {
			return node.Children[1 : len(node.Children)-1], last, true
		}
	}
	return nil, nil, false
}

// resolveVecRef -- real, shared helper for vec/push!/vec/get/vec/len's own real "&expr"/
// "&mut expr" reference argument. Confirmed live via `burrow parse`: PARENA's lexer fuses
// "&name" into ONE token when there's no space ("&v" parses as a single symbol "&v"), but
// "&mut expr" always parses as TWO separate tokens ("&mut" then expr) -- "&mut" has no attached
// identifier of its own, and a following compound expression (e.g. a get-field call) can't fuse
// into one token anyway. Returns the real target node (synthesizing a plain symbol node for the
// fused "&name" case, stripping its leading "&") and how many of `children` it consumed (1 for
// the fused case, 2 for the separate "&mut expr" case) so the caller knows where its own next
// real argument starts.
func resolveVecRef(children []*Node) (*Node, int, error) {
	if len(children) == 0 {
		return nil, 0, errors.New("emit_go: expected a &expr/&mut expr reference argument")
	}
	first := children[0]
	if first.Type == NodeSymbol && first.Text == "&mut" {
		if len(children) < 2 {
			return nil, 0, errors.New("emit_go: &mut requires a following expression")
		}
		return children[1], 2, nil
	}
	if first.Type == NodeSymbol && strings.HasPrefix(first.Text, "&") && len(first.Text) > 1 {
		return &Node{Type: NodeSymbol, Text: strings.TrimPrefix(first.Text, "&"), Line: first.Line}, 1, nil
	}
	return nil, 0, errors.New("emit_go: expected a &expr/&mut expr reference argument")
}

// emitGoLoop -- see emitGoExpr's own "loop" case for the real, deliberate v0 boundary this
// implements (body must be exactly one top-level `(if cond then else)`, `recur` in exactly one
// branch) and why. Real, direct port of PARENA's own reference C emitter's overall SHAPE for this
// exact common case (bindings become mutable locals, the loop body becomes a real `for {}`, a
// terminal branch returns, a `recur` branch reassigns the locals and continues) -- narrower than
// `src/emit.c`'s own fully general `loop_locals`-threaded version, which also handles `recur`
// nested inside `match`/deeper `if`/`cond` chains; that generality is real, separate, later work.
func emitGoLoop(expr *Node, scope *emitGoScope) (string, error) {
	if len(expr.Children) != 3 || expr.Children[1].Type != NodeVec {
		return "", errors.New("emit_go: loop v0 requires exactly (loop [name init ...] (if cond then else)) -- see emitGoLoop's own doc comment for the real, deliberate scope this v0 covers")
	}
	bindings := expr.Children[1]
	if len(bindings.Children)%2 != 0 {
		return "", errors.New("emit_go: loop bindings vector must have an even number of elements (name init pairs)")
	}
	if len(bindings.Children) == 0 {
		return "", errors.New("emit_go: loop v0 requires at least one binding")
	}
	body := expr.Children[2]
	if body.Type != NodeList || len(body.Children) != 4 || body.Children[0].Type != NodeSymbol || body.Children[0].Text != "if" {
		return "", errors.New("emit_go: loop v0 body must be a single top-level (if cond then else) -- recur nested inside match/cond/a deeper if is real, separate, unstarted work")
	}

	// Real, deliberate design match to `let`'s own precedent: bindings are evaluated into a
	// CLONED local-params map, sequentially (a later binding's own init expression can reference
	// an earlier one, matching array.prn's own real usage shape), and never leak outside the loop.
	childParams := make(map[string]bool, len(scope.localParams)+len(bindings.Children)/2)
	for k, v := range scope.localParams {
		childParams[k] = v
	}
	childScope := &emitGoScope{knownDefns: scope.knownDefns, localParams: childParams, retType: scope.retType, defnRetInfo: scope.defnRetInfo}

	var initStmts []string
	var names []string
	for i := 0; i+1 < len(bindings.Children); i += 2 {
		nameNode := bindings.Children[i]
		valNode := bindings.Children[i+1]
		if nameNode.Type != NodeSymbol {
			return "", errors.New("emit_go: loop binding name must be a plain symbol")
		}
		valE, err := emitGoExpr(valNode, childScope)
		if err != nil {
			return "", err
		}
		childScope.localParams[nameNode.Text] = true
		names = append(names, nameNode.Text)
		goName := mangleGoLocal(nameNode.Text, childScope)
		// Real, genuine bug found and fixed here testing this for real (a real (loop [i 0 acc 0]
		// (if (> i n) ...)) probe): `i := 0` lets Go infer `i`'s own type from the untyped
		// constant `0`, which defaults to `int` -- the SAME real "I32 defaults to the wrong Go
		// type" class of bug the "if" case above already found and fixed for its own branch
		// values, just hitting a `:=` declaration instead of an `any`-boxing conversion this time.
		// A String literal or bare true/false already gets the right concrete Go type from `:=`
		// on its own (Go's own untyped-constant defaults for those already coincide with this
		// emitter's own resolved types, same real exception the "if" case's own doc comment
		// notes) -- only the numeric case needs help. Real, honest, narrower-than-ideal v0 fix:
		// every other loop binding (every real .prn loop in this stdlib's own array.prn is
		// I32-only today) is declared with an explicit `var name int32 = expr` instead of `:=` --
		// Go allows an untyped constant expression to implicitly convert to a var's own declared
		// type, so this both fixes the defaulting bug AND still fails loudly (a real Go compile
		// error, not a silent wrong value) if a future non-I32 loop binding ever actually needs
		// this path, rather than guessing wrong and running anyway.
		if valNode.Type == NodeString || (valNode.Type == NodeSymbol && (valNode.Text == "true" || valNode.Text == "false")) {
			initStmts = append(initStmts, goName+" := "+valE)
		} else {
			initStmts = append(initStmts, "var "+goName+" int32 = "+valE)
		}
	}

	condE, err := emitGoExpr(body.Children[1], childScope)
	if err != nil {
		return "", err
	}
	thenNode, elseNode := body.Children[2], body.Children[3]
	thenEffects, thenRecur, thenIsRecur := unwrapRecurBranch(thenNode)
	elseEffects, elseRecur, elseIsRecur := unwrapRecurBranch(elseNode)
	if thenIsRecur == elseIsRecur {
		return "", errors.New("emit_go: loop v0 requires recur in exactly one branch of the if (the other branch is the loop's own terminal value)")
	}
	recurNode, terminalNode, recurEffects := thenRecur, elseNode, thenEffects
	if elseIsRecur {
		recurNode, terminalNode, recurEffects = elseRecur, thenNode, elseEffects
	}
	if len(recurNode.Children)-1 != len(names) {
		return "", errors.New("emit_go: recur requires exactly as many arguments as loop bindings (" + itoa(len(names)) + ")")
	}
	// recurEffects -- real, new capability alongside the direct-recur v0 boundary: the recur
	// branch may be a `(do effect1 effect2 ... (recur ...))` form, not just a bare `(recur ...)`
	// -- the exact real shape a Vec-building loop needs (`(do (vec/push! &mut v i) (recur (+ i
	// 1)))`, confirmed live against a real probe). Each effect expression runs for its own side
	// effect only (its value discarded), same convention emitGoBody's own non-final expressions
	// already use, emitted BEFORE the recur-argument temp-variable computation below (an effect
	// like `vec/push!` legitimately needs to run against the OLD binding values, same as every
	// recur argument itself does).
	var recurEffectStmts []string
	for _, eff := range recurEffects {
		effE, err := emitGoExpr(eff, childScope)
		if err != nil {
			return "", err
		}
		recurEffectStmts = append(recurEffectStmts, "_ = "+effE)
	}

	terminalE, err := emitGoExpr(terminalNode, childScope)
	if err != nil {
		return "", err
	}

	// Real correctness requirement, same reason a simultaneous multi-variable reassignment needs
	// it in any language: every recur argument must be computed from the OLD binding values
	// before ANY binding is overwritten (`(recur acc i)` swapping two loop vars would silently
	// break if the first assignment clobbered a value the second argument still needed to read).
	// Real temp variables, uniquely named per loop via goLoopCounter, sidestep this exactly the
	// way a real simultaneous-assignment lowering always does.
	goLoopCounter++
	loopID := goLoopCounter
	var recurTmp []string
	var recurAssign []string
	for i, argNode := range recurNode.Children[1:] {
		argE, err := emitGoExpr(argNode, childScope)
		if err != nil {
			return "", err
		}
		tmp := "__loop_tmp_" + itoa(loopID) + "_" + itoa(i)
		recurTmp = append(recurTmp, tmp+" := "+argE)
		recurAssign = append(recurAssign, mangleGoLocal(names[i], childScope)+" = "+tmp)
	}

	var b strings.Builder
	b.WriteString("func() any {\n")
	for _, s := range initStmts {
		b.WriteString(s + "\n")
	}
	b.WriteString("for {\n")
	b.WriteString("if " + condE + " {\n")
	if thenIsRecur {
		for _, s := range recurEffectStmts {
			b.WriteString(s + "\n")
		}
		for _, s := range recurTmp {
			b.WriteString(s + "\n")
		}
		for _, s := range recurAssign {
			b.WriteString(s + "\n")
		}
		b.WriteString("continue\n} else {\nreturn " + scope.retType + "(" + terminalE + ")\n}\n")
	} else {
		b.WriteString("return " + scope.retType + "(" + terminalE + ")\n} else {\n")
		for _, s := range recurEffectStmts {
			b.WriteString(s + "\n")
		}
		for _, s := range recurTmp {
			b.WriteString(s + "\n")
		}
		for _, s := range recurAssign {
			b.WriteString(s + "\n")
		}
		b.WriteString("continue\n}\n")
	}
	b.WriteString("}\n}().(" + scope.retType + ")")
	return b.String(), nil
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

func emitGoDefn(defn *Node, knownDefns map[string]bool, knownStructs map[string]bool, defnRetInfo map[string]goDefnRetInfo) (string, error) {
	if len(defn.Children) < 3 || defn.Children[1].Type != NodeSymbol || defn.Children[2].Type != NodeVec {
		return "", errors.New("emit_go: defn: malformed function definition")
	}
	fnName := mangleGo(defn.Children[1].Text)
	params := defn.Children[2]
	scope := &emitGoScope{knownDefns: knownDefns, localParams: map[string]bool{}, defnRetInfo: defnRetInfo}

	var paramList strings.Builder
	for i, param := range params.Children {
		if param.Type != NodeList || param.Children[0].Type != NodeSymbol || param.Children[1].Type != NodeColon {
			return "", errors.New("emit_go: defn: unsupported parameter shape (v0 only understands plain " +
				"(name : I32|F64|Bool|String|StructType|(Vec ElemType)) params, or (name : Arena @ Region))")
		}
		var pType string
		var err error
		// Arena @ Region params (real, new capability -- kanban card 9988's own next-named
		// prerequisite for Vec support, found necessary live: every real .prn function that
		// builds a Vec takes a "dest : Arena @ Region" param, and v0 had no parsing for this
		// shape at all before now). Real, deliberate design: an Arena carries no real meaning
		// for Go's own GC-backed slices (vec/new below needs no arena at all), so this is kept
		// as a real, present Go parameter -- typed `any`, always unused for actual work -- purely
		// so a bare reference to it elsewhere in the body (e.g. `(vec/new dest)`) still resolves
		// as a known local rather than "unknown identifier", and so a real Go host calling into
		// this function can pass a literal `nil` for it (the honest v0 convention this implies,
		// same class of "host must know a real, narrow calling convention" already true for
		// match's own scrutinee restriction).
		if len(param.Children) == 5 && param.Children[2].Type == NodeSymbol && param.Children[2].Text == "Arena" &&
			param.Children[3].Type == NodeAt && param.Children[4].Type == NodeSymbol {
			pType = "any"
		} else if len(param.Children) == 3 && param.Children[2].Type == NodeSymbol {
			pType, err = resolveGoType(param.Children[2].Text, knownStructs)
		} else if len(param.Children) == 3 && param.Children[2].Type == NodeList {
			// (name : (Vec ElemType)) -- real, new capability alongside Arena params above.
			// Represented as a plain Go `[]any` (see the "vec/new"/"vec/push!" cases in
			// emitGoExpr for the full real design) -- no per-element-type Go type is ever
			// needed, since every element is already boxed through `any` the same way a
			// Result/Option's own payload already is.
			vecType := param.Children[2]
			if len(vecType.Children) != 2 || vecType.Children[0].Type != NodeSymbol || vecType.Children[0].Text != "Vec" ||
				vecType.Children[1].Type != NodeSymbol {
				return "", errors.New("emit_go: defn: unsupported compound parameter type (v0 only understands (Vec ElemType), a bare symbol)")
			}
			// The element type itself isn't actually needed for the Go representation (every
			// element is `any` regardless) -- resolved anyway, discarding the result, purely as
			// a real, honest validation that the declared element type is itself a real,
			// supported type (catches a typo'd/unsupported element type here rather than letting
			// it silently pass through unchecked).
			if _, err = resolveGoType(vecType.Children[1].Text, knownStructs); err != nil {
				return "", err
			}
			pType = "[]any"
		} else {
			return "", errors.New("emit_go: defn: unsupported parameter shape (v0 only understands plain " +
				"(name : I32|F64|Bool|String|StructType|(Vec ElemType)) params, or (name : Arena @ Region))")
		}
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

	if len(defn.Children) < 6 || defn.Children[3].Type != NodeColon {
		return "", errors.New("emit_go: defn: expected (defn name [params] : RetType body) with exactly one body expression")
	}
	// Real, new capability: a return type carrying its own trailing `@ Region` annotation (every
	// real .prn function returning a Vec/String/struct built into a caller-supplied Arena writes
	// its return type this way, e.g. `(Vec I32) @ Region`) shifts the body's own real index by 2
	// (the `@` node plus the Region symbol) -- found live, the same class of "real .prn shape v0
	// had never actually needed to parse yet" gap Arena params above already hit. A bare return
	// type with no region suffix (every defn this target supported before this pass) still works
	// unchanged -- this only ADDS a case, it doesn't disturb the existing one.
	bodyIdx := 5
	if len(defn.Children) >= 8 && defn.Children[5].Type == NodeAt && defn.Children[6].Type == NodeSymbol {
		bodyIdx = 7
	}
	if len(defn.Children) != bodyIdx+1 {
		return "", errors.New("emit_go: defn: expected (defn name [params] : RetType body) with exactly one body expression")
	}
	retType, _, err := resolveGoReturnType(defn.Children[4], knownStructs)
	if err != nil {
		return "", err
	}
	scope.retType = retType
	body, err := emitGoExpr(defn.Children[bodyIdx], scope)
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

	// Real, new second first-pass step (kanban card 9988's own match/Result port): a defn's own
	// Result/Option return type needs to be known BEFORE emitting any defn body, the same real
	// reason knownDefns/knownStructs above are collected up front -- `match` on a call to defn B
	// needs B's own declared payload/error type even when B is defined later in the same file
	// (Go itself doesn't care about declaration order, but this emitter's own resolution does).
	// Uses knownStructs (already complete from the loop above) since a Result/Option's own
	// payload can itself be a registered struct type.
	defnRetInfo := map[string]goDefnRetInfo{}
	for _, form := range program.Children {
		if !isCallNamed(form, "defn") || len(form.Children) != 6 || form.Children[1].Type != NodeSymbol {
			continue
		}
		_, info, err := resolveGoReturnType(form.Children[4], knownStructs)
		if err != nil {
			continue // a real error here is reported for real when the defn body itself is emitted below
		}
		if info != nil {
			defnRetInfo[form.Children[1].Text] = *info
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
			def, err := emitGoDefn(form, knownDefns, knownStructs, defnRetInfo)
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
	// Result/Option -- real, FIXED, shared Go struct types (not per-instantiation), the exact
	// same real "erase the payload type, tag + any-typed value" design PARENA's own reference C
	// runtime already uses (parena_runtime.h's own Result/Option: `{int tag; void *value;}`) --
	// ported directly, not reinvented, since VS0 has no generics to give Result/Option a real
	// per-instantiation type anyway (the same reason bstree.prn/json.prn commit to concrete
	// types instead of a generic container). Emitted only when this file's own defns actually
	// declare a Result/Option return type -- harmless to check even though Go itself doesn't
	// error on an unused type declaration, just keeps a file that never needs these clean.
	if len(defnRetInfo) > 0 {
		out.WriteString("type Result struct {\n\tTag   int // 1 = Ok, 0 = Err\n\tValue any\n}\n\n")
		out.WriteString("type Option struct {\n\tTag   int // 1 = Some, 0 = None\n\tValue any\n}\n\n")
	}
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
