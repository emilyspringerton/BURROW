// emit_go_test.go — real v0 verification for the new Go emitter (emit_go.go), same real split
// emit_c_test.go already establishes: check the emitter's own success/failure behavior directly
// here. Real, new addition (kanban card 9988's own match/Result port):
// TestEmitGoMatchEndToEndBuildsAndRuns below DOES do the real `go build` + run verification as
// an actual `go test`, not manually alongside it -- real, valuable enough (match/Result's own
// payload-type resolution is genuinely easy to get subtly wrong) to be a permanent, repeatable
// regression check rather than a one-off manual probe.
package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func buildGo(t *testing.T, src string) (string, error) {
	t.Helper()
	program, err := ParseProgram(src)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if err := RegionAnalyze(program); err != nil {
		t.Fatalf("unexpected region-analyze error: %v", err)
	}
	return EmitGo(program)
}

func TestEmitGoZeroArgConstant(t *testing.T) {
	g, err := buildGo(t, "(defn xp-award [] : I32 60)")
	if err != nil {
		t.Fatalf("zero-arg I32 constant should emit successfully: %v", err)
	}
	if !strings.Contains(g, "func XpAward() int32 {") {
		t.Errorf("zero-arg defn name should be exported PascalCase, typed int32: got %s", g)
	}
	if !strings.Contains(g, "return 60") {
		t.Errorf("zero-arg defn body should return the real literal: got %s", g)
	}
	if !strings.Contains(g, "package burrowgen") {
		t.Errorf("emitted file should declare package burrowgen: got %s", g)
	}
}

func TestEmitGoScalarParamIfElseBinopNestedCall(t *testing.T) {
	g, err := buildGo(t, "(defn material-paper [] : I32 0)\n"+
		"(defn on-item-for-object-destroyed [(material : I32)] : I32\n"+
		"  (if (= material (material-paper)) 1 0))")
	if err != nil {
		t.Fatalf("scalar param + if/else + binop + nested call should emit successfully: %v", err)
	}
	if !strings.Contains(g, "func MaterialPaper() int32") {
		t.Errorf("first defn should be exported PascalCase: got %s", g)
	}
	if !strings.Contains(g, "func OnItemForObjectDestroyed(material int32) int32") {
		t.Errorf("second defn's own scalar param should be typed int32, real Go 'name Type' order: got %s", g)
	}
	// Real, empirical note: go/format.Source (run as EmitGo's own final pass, see its doc
	// comment) strips the redundant outer parens this emitter's own binop emission always adds
	// when the binop is the whole `if` condition -- expected, harmless, checked for the real
	// shape actually produced, not assumed from how emitGoExpr's own string-building looks in
	// isolation.
	if !strings.Contains(g, "material == MaterialPaper()") {
		t.Errorf("= binop should lower to Go's own ==, nested zero-arg call PascalCased: got %s", g)
	}
	if !strings.Contains(g, "if material == MaterialPaper() {") || !strings.Contains(g, "int32(1)") || !strings.Contains(g, "int32(0)") {
		t.Errorf("if/else should lower to a real Go if/else via an inline func literal, branches cast to the real declared return type: got %s", g)
	}
}

func TestEmitGoNotUnaryOp(t *testing.T) {
	// Same real gap as emit_c_test.go's own TestEmitCNotUnaryOp, same real trigger
	// (stdlib/k8s/operator.prn's own (if (not exists) ...)).
	g, err := buildGo(t, "(defn f [(x : Bool)] : Bool (not x))")
	if err != nil {
		t.Fatalf("not should emit successfully: %v", err)
	}
	if !strings.Contains(g, "(!(x))") {
		t.Errorf("not should lower to Go's own !, not a bogus call: got %s", g)
	}
}

func TestEmitGoBitwiseAndModOps(t *testing.T) {
	g, err := buildGo(t, "(defn base4-xor [(a : I32) (b : I32)] : I32 (bit-xor a b))\n"+
		"(defn base4-add [(a : I32) (b : I32)] : I32 (mod (+ a b) 4))")
	if err != nil {
		t.Fatalf("bitwise/mod ops should emit successfully: %v", err)
	}
	if !strings.Contains(g, "(a ^ b)") {
		t.Errorf("bit-xor should lower to ^: got %s", g)
	}
	if !strings.Contains(g, "((a + b) % 4)") {
		t.Errorf("mod should lower to %%: got %s", g)
	}
}

func TestEmitGoNoForwardDeclarationsNeeded(t *testing.T) {
	// Real, deliberate structural difference from emit_c.go: Go has no forward-declaration
	// concept at all, so a sibling call to a function defined LATER in the file just works
	// without any decl/def split -- this is the real, positive proof of that, not merely the
	// absence of a decl.
	g, err := buildGo(t, "(defn a [] : I32 (b))\n(defn b [] : I32 42)")
	if err != nil {
		t.Fatalf("forward-reference to a later sibling defn should emit successfully: %v", err)
	}
	if !strings.Contains(g, "func A() int32 {\n\treturn B()") {
		t.Errorf("a() should call B() directly with no forward declaration needed: got %s", g)
	}
	if !strings.Contains(g, "func B() int32 {\n\treturn 42") {
		t.Errorf("b() should be emitted with its own real definition: got %s", g)
	}
}

func TestEmitGoWrongArityBinopIsError(t *testing.T) {
	_, err := buildGo(t, "(defn f [] : I32 (+ 1 2 3))")
	if err == nil {
		t.Error("a binary operator called with the wrong arity should be a real, honest error, not silently accepted")
	}
}

func TestEmitGoArenaParamIsError(t *testing.T) {
	_, err := buildGo(t, "(defn f [(buf : Arena @ :region/scratch)] : I32 0)")
	if err == nil {
		t.Error("an Arena/region-annotated parameter should be a real, honest unsupported error, not silently guessed")
	}
}

func TestEmitGoUnknownNamespacedIdentifierIsError(t *testing.T) {
	// Same real class of bug emit_c.go's own test already guards against for the C target --
	// verifying the Go target rejects it exactly the same honest way, not silently mangling
	// `math/pi` through into broken Go syntax.
	_, err := buildGo(t, "(defn f [] : F64 math/pi)")
	if err == nil {
		t.Fatal("a real, unknown namespaced identifier should be a real, honest error, not silently mangled through")
	}
	if !strings.Contains(err.Error(), "unknown identifier 'math/pi'") {
		t.Errorf("error should name the real, exact unknown identifier: got %v", err)
	}
}

func TestEmitGoUnknownCallIsError(t *testing.T) {
	_, err := buildGo(t, "(defn f [] : F64 (math/random))")
	if err == nil {
		t.Fatal("a real, unknown call should be a real, honest error, not silently mangled through")
	}
	if !strings.Contains(err.Error(), "unknown identifier 'math/random'") {
		t.Errorf("error should name the real, exact unknown call: got %v", err)
	}
}

func TestEmitGoUnitReturnType(t *testing.T) {
	// Real, Go-specific case emit_c.go's own test suite has no equivalent for: Unit's own real
	// scalar-type mapping is void in C but Go has no true void expression type, so this checks
	// the real, deliberate struct{} choice actually round-trips.
	g, err := buildGo(t, "(defn noop [] : Unit 0)")
	if err != nil {
		t.Fatalf("Unit-returning defn should emit successfully: %v", err)
	}
	if !strings.Contains(g, "func Noop() struct{}") {
		t.Errorf("Unit should map to Go's own real struct{} idiom: got %s", g)
	}
}

func TestEmitGoDefstructAndGetField(t *testing.T) {
	// Real, new capability -- the exact real shape stdlib/k8s/k8s.prn's own ServiceSpec + a
	// field-reading function need.
	g, err := buildGo(t, "(defstruct ServiceSpec\n  (name : String)\n  (port : I32))\n"+
		"(defn service-port [(s : ServiceSpec)] : I32\n"+
		"  (get-field s :port))")
	// Real defstruct field lists are plain lists, direct children of the defstruct form (not
	// wrapped in a vec) -- matching k8s.prn's own real shape, confirmed via a real `burrow
	// parse` probe before writing this test, not guessed.
	if err != nil {
		t.Fatalf("defstruct + get-field should emit successfully: %v", err)
	}
	if !strings.Contains(g, "type ServiceSpec struct") {
		t.Errorf("defstruct should emit an exported Go struct type: got %s", g)
	}
	if !strings.Contains(g, "Name string") || !strings.Contains(g, "Port int32") {
		t.Errorf("struct fields should be exported PascalCase, real Go types: got %s", g)
	}
	if !strings.Contains(g, "func ServicePort(s ServiceSpec) int32") {
		t.Errorf("a struct-typed param should resolve to the real exported struct type: got %s", g)
	}
	if !strings.Contains(g, "(s).Port") {
		t.Errorf("get-field should lower to real Go dot access: got %s", g)
	}
}

func TestEmitGoUnknownStructTypeIsError(t *testing.T) {
	_, err := buildGo(t, "(defn f [(x : NotRegistered)] : I32 0)")
	if err == nil {
		t.Error("an unregistered struct type name should be a real, honest error, not silently accepted")
	}
}

func TestEmitGoKebabCaseParamName(t *testing.T) {
	// Real Go-specific naming rule emit_c.go's own mangleC already handles identically for C
	// (dash -> underscore) but is worth its own explicit check here since Go param names are
	// lowercase (unexported) by real Go convention, unlike the exported PascalCase function name.
	g, err := buildGo(t, "(defn next-focus-index [(idx : I32) (len : I32) (delta : I32)] : I32\n"+
		"  (mod (+ (+ idx delta) len) len))")
	if err != nil {
		t.Fatalf("kebab-case param names should emit successfully: %v", err)
	}
	if !strings.Contains(g, "func NextFocusIndex(idx int32, len int32, delta int32) int32") {
		t.Errorf("multi-param real signature should be exported PascalCase fn, lowercase params: got %s", g)
	}
}

// TestEmitGoBoolLiterals -- real gap found live building PARENA/stdlib/datetime.prn's own
// is-leap-year?: bare true/false Bool literals had no handling at all in emitGoExpr's
// NodeSymbol case, so any real .prn function returning one failed with "unknown identifier
// 'false'" rather than emitting Go's own literal.
func TestEmitGoBoolLiterals(t *testing.T) {
	g, err := buildGo(t, "(defn always-true [] : Bool true)\n(defn always-false [] : Bool false)")
	if err != nil {
		t.Fatalf("bare true/false literals should emit successfully: %v", err)
	}
	if !strings.Contains(g, "return true") {
		t.Errorf("expected a real Go `true` literal: got %s", g)
	}
	if !strings.Contains(g, "return false") {
		t.Errorf("expected a real Go `false` literal: got %s", g)
	}
}

// TestEmitGoPredicateNameMangling -- real gap found live alongside the bool-literal one above:
// a trailing `?` (or `!`) in a defn name produced an illegal Go identifier character, confirmed
// via a real `gofmt` failure ("illegal character U+003F '?'"). `src/emit.c`'s own C emitter
// already mangles `?`/`!` to `_` (same as `-`/`/`) -- mirrored here, not invented fresh.
func TestEmitGoPredicateNameMangling(t *testing.T) {
	g, err := buildGo(t, "(defn is-leap-year? [(year : I32)] : Bool (= (mod year 4) 0))")
	if err != nil {
		t.Fatalf("a trailing '?' in a defn name should mangle to a legal Go identifier, not error: %v", err)
	}
	if !strings.Contains(g, "func IsLeapYear_(year int32) bool") {
		t.Errorf("expected '?' mangled to '_' matching emit_c.go's own convention: got %s", g)
	}
}

// TestEmitGoLetSingleBinding -- real, new capability (kanban card 1199/9988): the single largest
// real gap blocking any real multi-statement .prn function from reaching this target at all --
// v0 had no let-bindings whatsoever before this.
func TestEmitGoLetSingleBinding(t *testing.T) {
	g, err := buildGo(t, "(defn double-it [(x : I32)] : I32\n  (let [y (* x 2)] y))")
	if err != nil {
		t.Fatalf("a single-binding let should emit successfully: %v", err)
	}
	if !strings.Contains(g, "y := ") {
		t.Errorf("expected a real Go local variable declaration for the let binding: got %s", g)
	}
	if !strings.Contains(g, "x * 2") {
		t.Errorf("expected the binding's own expression to be emitted: got %s", g)
	}
}

// TestEmitGoLetMultipleBindingsSequentialScope -- a later binding can reference an earlier one
// (real, ordinary let semantics), and the let's own final body expression is the real result.
func TestEmitGoLetMultipleBindingsSequentialScope(t *testing.T) {
	g, err := buildGo(t, "(defn compute [(x : I32)] : I32\n  (let [a (+ x 1) b (* a 2)] b))")
	if err != nil {
		t.Fatalf("multi-binding let with sequential scope should emit successfully: %v", err)
	}
	if !strings.Contains(g, "a := ") || !strings.Contains(g, "b := ") {
		t.Errorf("expected both real Go local declarations: got %s", g)
	}
	if !strings.Contains(g, "a * 2") {
		t.Errorf("expected the second binding to reference the first by its real local name: got %s", g)
	}
}

// TestEmitGoLetBindingDoesNotLeakOutsideLet -- real, deliberate scope check: a name bound inside
// one let must not be visible to an expression outside it, even in the same function (e.g. two
// sibling lets, or a let followed by plain code) -- confirms the childParams CLONE (not mutating
// scope's own map) actually works, not just that it compiles once.
func TestEmitGoLetBindingDoesNotLeakOutsideLet(t *testing.T) {
	_, err := buildGo(t, "(defn broken [(x : I32)] : I32\n  (if (> x 0) (let [y (* x 2)] y) y))")
	if err == nil {
		t.Fatal("expected a real 'unknown identifier' error -- y is scoped to the let's own branch, not the sibling else branch")
	}
	if !strings.Contains(err.Error(), "unknown identifier 'y'") {
		t.Errorf("expected the real unknown-identifier error naming y specifically: got %v", err)
	}
}

// TestEmitGoNestedLetInsideIf -- real, deliberate composition probe, the same real discipline
// that caught two genuine bugs in the "if" case's own any-boxing (see that case's own doc
// comment): a let nested inside an if branch must produce a concrete, correctly-typed value the
// outer if's own any-boxing/RetType-cast can wrap without a second, different bug class.
func TestEmitGoNestedLetInsideIf(t *testing.T) {
	g, err := buildGo(t, "(defn clamp-and-double [(x : I32)] : I32\n"+
		"  (if (> x 0) (let [doubled (* x 2)] doubled) 0))")
	if err != nil {
		t.Fatalf("a let nested inside an if branch should emit and compile successfully: %v", err)
	}
	if !strings.Contains(g, "doubled := ") {
		t.Errorf("expected the nested let's own real local declaration: got %s", g)
	}
}

// TestEmitGoDoSequencesForEffect -- real, new capability alongside let: (do a b c) runs a and b
// for effect (their own values discarded), c is the real result.
func TestEmitGoDoSequencesForEffect(t *testing.T) {
	g, err := buildGo(t, "(defn material-paper [] : I32 0)\n"+
		"(defn seq-test [(x : I32)] : I32\n  (do (material-paper) x))")
	if err != nil {
		t.Fatalf("do should emit successfully: %v", err)
	}
	if !strings.Contains(g, "_ = MaterialPaper()") {
		t.Errorf("expected the non-final do expression to run for effect, value discarded: got %s", g)
	}
}

// TestEmitGoLetTwoBindingsChained -- real, second real multi-binding probe (a := then b := a+1,
// distinct expression shape from the multiply/add mix above) confirming sequential chaining
// isn't a one-shape coincidence.
func TestEmitGoLetTwoBindingsChained(t *testing.T) {
	g, err := buildGo(t, "(defn double-it [(x : I32)] : I32\n  (let [y (* x 2) z (+ y 1)] z))")
	if err != nil {
		t.Fatalf("unexpected emit error: %v", err)
	}
	if !strings.Contains(g, "y := ") || !strings.Contains(g, "z := ") {
		t.Errorf("expected both real Go local declarations: got %s", g)
	}
	if !strings.Contains(g, "y + 1") {
		t.Errorf("expected the second binding to reference the first: got %s", g)
	}
}

// TestEmitGoStringLiteral -- real, pre-existing gap found live while building match/Result
// support (kanban card 9988): string literals had no handling anywhere in this file, even
// though String was already a supported param/return type.
func TestEmitGoStringLiteral(t *testing.T) {
	g, err := buildGo(t, `(defn greeting [] : String "hello")`)
	if err != nil {
		t.Fatalf("a plain string literal should emit successfully: %v", err)
	}
	if !strings.Contains(g, `return "hello"`) {
		t.Errorf("expected a real Go string literal: got %s", g)
	}
}

// TestEmitGoStringLiteralWithQuoteAndBackslash -- real correctness check that strconv.Quote is
// actually doing the re-escaping work, not a naive passthrough that would emit invalid Go for a
// literal containing a real `"` or `\`.
func TestEmitGoStringLiteralWithQuoteAndBackslash(t *testing.T) {
	g, err := buildGo(t, `(defn tricky [] : String "a \"quoted\" \\ value")`)
	if err != nil {
		t.Fatalf("unexpected emit error: %v", err)
	}
	if !strings.Contains(g, `"a \"quoted\" \\ value"`) {
		t.Errorf("expected a real, correctly re-escaped Go string literal: got %s", g)
	}
}

// TestEmitGoResultConstruction -- real, new capability (kanban card 9988): Ok/Err construct the
// real, fixed, shared Result struct (Tag/Value), matching PARENA's own reference C runtime's
// {tag, value} representation.
func TestEmitGoResultConstruction(t *testing.T) {
	g, err := buildGo(t, "(defn safe-div [(a : I32) (b : I32)] : (Result I32 String)\n"+
		`  (if (= b 0) (Err "division by zero") (Ok (/ a b))))`)
	if err != nil {
		t.Fatalf("a Result-returning defn should emit successfully: %v", err)
	}
	if !strings.Contains(g, "type Result struct") {
		t.Errorf("expected the real, shared Result struct type declared: got %s", g)
	}
	if !strings.Contains(g, "func SafeDiv(a int32, b int32) Result") {
		t.Errorf("expected Result as the real declared Go return type: got %s", g)
	}
	if !strings.Contains(g, `Result{Tag: 0, Value: "division by zero"}`) {
		t.Errorf("expected a real Err construction: got %s", g)
	}
	if !strings.Contains(g, "Result{Tag: 1, Value: (a / b)}") {
		t.Errorf("expected a real Ok construction: got %s", g)
	}
}

// TestEmitGoOptionConstruction -- real, new capability: Some/bare None construct the real,
// shared Option struct. Bare `None` (no parens) is the real, established PARENA source
// convention (see PARENA/stdlib/bstree.prn's own real, live get-loop).
func TestEmitGoOptionConstruction(t *testing.T) {
	g, err := buildGo(t, "(defn half-of-even [(x : I32)] : (Option I32)\n"+
		"  (if (= (mod x 2) 0) (Some (/ x 2)) None))")
	if err != nil {
		t.Fatalf("an Option-returning defn should emit successfully: %v", err)
	}
	if !strings.Contains(g, "type Option struct") {
		t.Errorf("expected the real, shared Option struct type declared: got %s", g)
	}
	if !strings.Contains(g, "Option{Tag: 1, Value: (x / 2)}") {
		t.Errorf("expected a real Some construction: got %s", g)
	}
	if !strings.Contains(g, "Option{Tag: 0, Value: nil}") {
		t.Errorf("expected bare None to construct a real, empty Option: got %s", g)
	}
}

// TestEmitGoMatchOnResult -- real, new capability: matching a direct call to a known
// Result-returning defn, both Ok and Err clauses, real payload/error types resolved from that
// defn's own declared return type (not left generically `any`).
func TestEmitGoMatchOnResult(t *testing.T) {
	g, err := buildGo(t, "(defn safe-div [(a : I32) (b : I32)] : (Result I32 String)\n"+
		`  (if (= b 0) (Err "division by zero") (Ok (/ a b))))`+"\n"+
		"(defn describe-div [(a : I32) (b : I32)] : I32\n"+
		"  (match (safe-div a b)\n"+
		"    ((Ok result) result)\n"+
		"    ((Err msg) -1)))")
	if err != nil {
		t.Fatalf("match on a real Result-returning defn should emit successfully: %v", err)
	}
	if !strings.Contains(g, "SafeDiv(a, b)") {
		t.Errorf("expected the real scrutinee call: got %s", g)
	}
	if !strings.Contains(g, ".Value.(int32)") {
		t.Errorf("expected the Ok clause's bound value cast to the real, resolved payload type (int32), not left as 'any': got %s", g)
	}
	if !strings.Contains(g, ".Value.(string)") {
		t.Errorf("expected the Err clause's bound value cast to the real, resolved error type (string): got %s", g)
	}
	// The Err clause's own bound "msg" is never used in its body -- must not trip Go's real
	// "declared and not used" compile error (a real, found-live bug this test guards against).
	if !strings.Contains(g, "_ = msg") {
		t.Errorf("expected an explicit discard for the unused Err-clause binding: got %s", g)
	}
}

// TestEmitGoMatchOnOption -- same real capability, Option/Some/None this time, including that
// `None` (no payload) correctly takes no binding at all.
func TestEmitGoMatchOnOption(t *testing.T) {
	g, err := buildGo(t, "(defn half-of-even [(x : I32)] : (Option I32)\n"+
		"  (if (= (mod x 2) 0) (Some (/ x 2)) None))\n"+
		"(defn describe-half [(x : I32)] : I32\n"+
		"  (match (half-of-even x)\n"+
		"    ((Some result) result)\n"+
		"    ((None) -99)))")
	if err != nil {
		t.Fatalf("match on a real Option-returning defn should emit successfully: %v", err)
	}
	if !strings.Contains(g, "HalfOfEven(x)") {
		t.Errorf("expected the real scrutinee call: got %s", g)
	}
	if !strings.Contains(g, ".Value.(int32)") {
		t.Errorf("expected the Some clause's bound value cast to the real, resolved payload type: got %s", g)
	}
}

// TestEmitGoMatchUnknownScrutineeIsError -- real, deliberate v0 boundary enforced: matching
// anything other than a direct call to a known Result/Option-returning defn is a real, honest
// compile error, not silently wrong output.
func TestEmitGoMatchUnknownScrutineeIsError(t *testing.T) {
	_, err := buildGo(t, "(defn describe [(x : I32)] : I32\n"+
		"  (match x\n"+
		"    ((Ok result) result)\n"+
		"    ((Err msg) -1)))")
	if err == nil {
		t.Fatal("expected a real error: v0 only supports matching a direct call to a known Result/Option-returning defn")
	}
}

// TestEmitGoMatchDuplicateTagIsError -- real, deliberate exhaustiveness check: two clauses
// naming the same real tag (e.g. two Ok clauses) is a real, honest error, not silently letting
// the second one be dead code.
func TestEmitGoMatchDuplicateTagIsError(t *testing.T) {
	_, err := buildGo(t, "(defn safe-div [(a : I32) (b : I32)] : (Result I32 String)\n"+
		`  (if (= b 0) (Err "division by zero") (Ok (/ a b))))`+"\n"+
		"(defn describe-div [(a : I32) (b : I32)] : I32\n"+
		"  (match (safe-div a b)\n"+
		"    ((Ok result) result)\n"+
		"    ((Ok result) result)))")
	if err == nil {
		t.Fatal("expected a real error: two clauses both matching tag 1 (Ok)")
	}
}

// TestEmitGoMatchEndToEndBuildsAndRuns -- real, live, end-to-end proof matching this session's
// own established discipline: emitted code isn't just accepted by this emitter, it's real,
// valid, RUNNABLE Go with correct semantics for all four real branches (Ok/Err/Some/None).
func TestEmitGoMatchEndToEndBuildsAndRuns(t *testing.T) {
	src := "(module match-e2e)\n(export describe-div describe-half)\n" +
		"(defn safe-div [(a : I32) (b : I32)] : (Result I32 String)\n" +
		`  (if (= b 0) (Err "division by zero") (Ok (/ a b))))` + "\n" +
		"(defn half-of-even [(x : I32)] : (Option I32)\n" +
		"  (if (= (mod x 2) 0) (Some (/ x 2)) None))\n" +
		"(defn describe-div [(a : I32) (b : I32)] : I32\n" +
		"  (match (safe-div a b)\n" +
		"    ((Ok result) result)\n" +
		"    ((Err msg) -1)))\n" +
		"(defn describe-half [(x : I32)] : I32\n" +
		"  (match (half-of-even x)\n" +
		"    ((Some result) result)\n" +
		"    ((None) -99)))\n"

	program, err := ParseProgram(src)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if err := RegionAnalyze(program); err != nil {
		t.Fatalf("unexpected region-analyze error: %v", err)
	}
	g, err := EmitGo(program)
	if err != nil {
		t.Fatalf("unexpected emit error: %v", err)
	}

	dir := t.TempDir()
	pkgDir := dir + "/burrowgen"
	if err := os.Mkdir(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pkgDir+"/gen.go", []byte(g), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dir+"/go.mod", []byte("module matche2e\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mainSrc := `package main

import (
	"fmt"

	"matche2e/burrowgen"
)

func main() {
	fmt.Println(burrowgen.DescribeDiv(10, 2))
	fmt.Println(burrowgen.DescribeDiv(10, 0))
	fmt.Println(burrowgen.DescribeHalf(8))
	fmt.Println(burrowgen.DescribeHalf(7))
}
`
	if err := os.WriteFile(dir+"/main.go", []byte(mainSrc), 0o644); err != nil {
		t.Fatal(err)
	}

	binPath := dir + "/bin"
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	buildCmd.Dir = dir
	buildCmd.Env = append(os.Environ(), "GOWORK=off")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("emitted Go failed to build (this is a real bug in emit_go.go, not the test): %v\n%s", err, out)
	}

	runCmd := exec.Command(binPath)
	out, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("running the built binary failed: %v\n%s", err, out)
	}
	got := string(out)
	want := "5\n-1\n4\n-99\n"
	if got != want {
		t.Fatalf("wrong real runtime output: got %q, want %q", got, want)
	}
}

// TestEmitGoLoopTerminalInElseBranch -- the real, common shape array.prn's own product/sum
// functions use: (if cond terminal (recur ...)).
func TestEmitGoLoopTerminalInElseBranch(t *testing.T) {
	g, err := buildGo(t, "(defn sum-to [(n : I32)] : I32\n"+
		"  (loop [i 0 acc 0]\n"+
		"    (if (> i n) acc (recur (+ i 1) (+ acc i)))))")
	if err != nil {
		t.Fatalf("a loop with recur in the else branch should emit successfully: %v", err)
	}
	if !strings.Contains(g, "for {") {
		t.Errorf("expected a real Go for-loop: got %s", g)
	}
	if !strings.Contains(g, "continue") {
		t.Errorf("expected the recur branch to continue the loop: got %s", g)
	}
}

// TestEmitGoLoopTerminalInThenBranch -- the other real ordering: (if cond (recur ...) terminal).
func TestEmitGoLoopTerminalInThenBranch(t *testing.T) {
	g, err := buildGo(t, "(defn sum-to [(n : I32)] : I32\n"+
		"  (loop [i 0 acc 0]\n"+
		"    (if (<= i n) (recur (+ i 1) (+ acc i)) acc)))")
	if err != nil {
		t.Fatalf("a loop with recur in the then branch should emit successfully: %v", err)
	}
	if !strings.Contains(g, "for {") {
		t.Errorf("expected a real Go for-loop: got %s", g)
	}
}

// TestEmitGoLoopBindingDoesNotLeakOutside -- same real scoping discipline as let's own leak test.
func TestEmitGoLoopBindingDoesNotLeakOutside(t *testing.T) {
	_, err := buildGo(t, "(defn broken [(n : I32)] : I32\n"+
		"  (if (> n 0) (loop [i 0] (if (> i n) i (recur (+ i 1)))) i))")
	if err == nil {
		t.Fatal("expected a real 'unknown identifier' error -- i is scoped to the loop, not the sibling else branch")
	}
	if !strings.Contains(err.Error(), "unknown identifier 'i'") {
		t.Errorf("expected the real unknown-identifier error naming i specifically: got %v", err)
	}
}

// TestEmitGoLoopRecurInBothBranchesIsError -- real, honest boundary: a loop needs exactly one
// terminal branch and one recur branch, not two of either.
func TestEmitGoLoopRecurInBothBranchesIsError(t *testing.T) {
	_, err := buildGo(t, "(defn broken [(n : I32)] : I32\n"+
		"  (loop [i 0] (if (> i n) (recur i) (recur (+ i 1)))))")
	if err == nil {
		t.Fatal("expected a real error -- recur in both branches leaves no real terminal value")
	}
}

// TestEmitGoLoopArityMismatchIsError -- recur must supply exactly as many arguments as bindings.
func TestEmitGoLoopArityMismatchIsError(t *testing.T) {
	_, err := buildGo(t, "(defn broken [(n : I32)] : I32\n"+
		"  (loop [i 0 acc 0] (if (> i n) acc (recur (+ i 1)))))")
	if err == nil {
		t.Fatal("expected a real arity-mismatch error -- recur supplies 1 argument for 2 loop bindings")
	}
}

// TestEmitGoLoopNestedIfBodyIsError -- real, deliberate v0 boundary named explicitly in
// emitGoLoop's own doc comment: recur nested deeper than a single top-level if is not supported.
func TestEmitGoLoopNestedIfBodyIsError(t *testing.T) {
	_, err := buildGo(t, "(defn broken [(n : I32)] : I32\n"+
		"  (loop [i 0] (+ 1 (if (> i n) i (recur (+ i 1))))))")
	if err == nil {
		t.Fatal("expected a real error -- the loop body here is a binop wrapping an if, not a bare top-level if")
	}
}

// TestEmitGoLoopEndToEndBuildsAndRuns -- real, live proof: a real .prn loop (sum-to, matching
// array.prn's own real product/sum shape) compiled via burrow, linked into a real, separate Go
// module, run, and checked against the hand-computed correct answer (triangular number formula).
func TestEmitGoLoopEndToEndBuildsAndRuns(t *testing.T) {
	src := "(module loop-e2e)\n(export sum-to)\n" +
		"(defn sum-to [(n : I32)] : I32\n" +
		"  (loop [i 0 acc 0]\n" +
		"    (if (> i n) acc (recur (+ i 1) (+ acc i)))))\n"

	program, err := ParseProgram(src)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if err := RegionAnalyze(program); err != nil {
		t.Fatalf("unexpected region-analyze error: %v", err)
	}
	g, err := EmitGo(program)
	if err != nil {
		t.Fatalf("unexpected emit error: %v", err)
	}

	dir := t.TempDir()
	pkgDir := dir + "/burrowgen"
	if err := os.Mkdir(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pkgDir+"/gen.go", []byte(g), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dir+"/go.mod", []byte("module loope2e\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mainSrc := `package main

import (
	"fmt"

	"loope2e/burrowgen"
)

func main() {
	fmt.Println(burrowgen.SumTo(0))
	fmt.Println(burrowgen.SumTo(1))
	fmt.Println(burrowgen.SumTo(10))
	fmt.Println(burrowgen.SumTo(100))
}
`
	if err := os.WriteFile(dir+"/main.go", []byte(mainSrc), 0o644); err != nil {
		t.Fatal(err)
	}

	binPath := dir + "/bin"
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	buildCmd.Dir = dir
	buildCmd.Env = append(os.Environ(), "GOWORK=off")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("emitted Go failed to build (this is a real bug in emit_go.go, not the test): %v\n%s", err, out)
	}

	runCmd := exec.Command(binPath)
	out, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("running the built binary failed: %v\n%s", err, out)
	}
	got := string(out)
	// 0, 0+1=1, 0+..+10=55, 0+..+100=5050 -- real, hand-computed triangular numbers.
	want := "0\n1\n55\n5050\n"
	if got != want {
		t.Fatalf("wrong real runtime output: got %q, want %q", got, want)
	}
}

// TestEmitGoVecArenaParam -- real, new capability: an Arena @ Region param is accepted and typed
// `any`, and a return type carrying its own trailing `@ Region` still resolves correctly.
func TestEmitGoVecArenaParam(t *testing.T) {
	g, err := buildGo(t, "(defn make [(dest : Arena @ Region)] : (Vec I32) @ Region\n"+
		"  (vec/new dest))")
	if err != nil {
		t.Fatalf("an Arena @ Region param + a region-annotated Vec return type should emit successfully: %v", err)
	}
	if !strings.Contains(g, "func Make(dest any) []any {") {
		t.Errorf("expected dest typed any and a []any return type: got %s", g)
	}
	if !strings.Contains(g, "[]any(nil)") {
		t.Errorf("expected vec/new to emit a nil []any: got %s", g)
	}
}

// TestEmitGoVecPushAndLen -- real construction + real length read.
func TestEmitGoVecPushAndLen(t *testing.T) {
	g, err := buildGo(t, "(defn count-em [(dest : Arena @ Region)] : I32\n"+
		"  (let [v (vec/new dest)]\n"+
		"    (do\n"+
		"      (vec/push! &mut v 1)\n"+
		"      (vec/push! &mut v 2)\n"+
		"      (vec/len &v))))")
	if err != nil {
		t.Fatalf("vec/push!/vec/len should emit successfully: %v", err)
	}
	if !strings.Contains(g, "v = append(v, int32(1))") || !strings.Contains(g, "v = append(v, int32(2))") {
		t.Errorf("expected real append calls with explicit int32 boxing: got %s", g)
	}
}

// TestEmitGoVecPushFusedRef -- the OTHER real reference-token shape ("&v", no space, one fused
// token) alongside "&mut v" (two tokens) already covered above.
func TestEmitGoVecPushFusedRef(t *testing.T) {
	g, err := buildGo(t, "(defn total [(v : (Vec I32))] : I32\n"+
		"  (vec/len &v))")
	if err != nil {
		t.Fatalf("a fused &v reference should emit successfully: %v", err)
	}
	if !strings.Contains(g, "int32(len(v))") {
		t.Errorf("expected a real len() call against v: got %s", g)
	}
}

// TestEmitGoVecPushInvalidTargetIsError -- real, deliberate v0 boundary: the mutation target
// must be a plain local or a get-field expression, not an arbitrary expression.
func TestEmitGoVecPushInvalidTargetIsError(t *testing.T) {
	_, err := buildGo(t, "(defn f [(dest : Arena @ Region)] : I32\n"+
		"  (do (vec/push! &mut (vec/new dest) 1) 0))")
	if err == nil {
		t.Fatal("expected a real error -- (vec/new dest) is not a real Go l-value vec/push! can reassign")
	}
}

// TestEmitGoVecEndToEndBuildsAndRuns -- real, live proof: a real .prn program builds a Vec via a
// loop + vec/push! (the exact real "do a side effect, then recur" shape array.prn's own zeros
// uses), then reads it back via vec/len + vec/get + deref, compiled via burrow, linked into a
// real, separate Go module, and run.
func TestEmitGoVecEndToEndBuildsAndRuns(t *testing.T) {
	src := "(module vec-e2e)\n(export sum-of)\n" +
		"(defn build [(n : I32) (dest : Arena @ Region)] : (Vec I32) @ Region\n" +
		"  (let [v (vec/new dest)]\n" +
		"    (do\n" +
		"      (loop [i 0]\n" +
		"        (if (>= i n) 0 (do (vec/push! &mut v i) (recur (+ i 1)))))\n" +
		"      v)))\n" +
		"(defn total [(v : (Vec I32))] : I32\n" +
		"  (loop [i 0 acc 0]\n" +
		"    (if (>= i (vec/len &v)) acc (recur (+ i 1) (+ acc (deref (vec/get &v i)))))))\n" +
		"(defn sum-of [(n : I32) (dest : Arena @ Region)] : I32\n" +
		"  (total (build n dest)))\n"

	program, err := ParseProgram(src)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if err := RegionAnalyze(program); err != nil {
		t.Fatalf("unexpected region-analyze error: %v", err)
	}
	g, err := EmitGo(program)
	if err != nil {
		t.Fatalf("unexpected emit error: %v", err)
	}

	dir := t.TempDir()
	pkgDir := dir + "/burrowgen"
	if err := os.Mkdir(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pkgDir+"/gen.go", []byte(g), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dir+"/go.mod", []byte("module vece2e\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mainSrc := `package main

import (
	"fmt"

	"vece2e/burrowgen"
)

func main() {
	fmt.Println(burrowgen.SumOf(5, nil))
	fmt.Println(burrowgen.SumOf(0, nil))
	fmt.Println(burrowgen.SumOf(10, nil))
}
`
	if err := os.WriteFile(dir+"/main.go", []byte(mainSrc), 0o644); err != nil {
		t.Fatal(err)
	}

	binPath := dir + "/bin"
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	buildCmd.Dir = dir
	buildCmd.Env = append(os.Environ(), "GOWORK=off")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("emitted Go failed to build (this is a real bug in emit_go.go, not the test): %v\n%s", err, out)
	}

	runCmd := exec.Command(binPath)
	out, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("running the built binary failed: %v\n%s", err, out)
	}
	got := string(out)
	// sum(0..4)=10, sum of nothing=0, sum(0..9)=45 -- real, hand-computed.
	want := "10\n0\n45\n"
	if got != want {
		t.Fatalf("wrong real runtime output: got %q, want %q", got, want)
	}
}

// TestEmitGoDefenumZeroPayloadVariant and its siblings below verify real, new `defenum` support
// (kanban card 9988's own next-named prerequisite after match/Result/loop/Vec).
func TestEmitGoDefenumZeroPayloadVariant(t *testing.T) {
	g, err := buildGo(t, "(defenum ParseError (EmptyInput) (Invalid (msg : String)))\n"+
		"(defn empty-error [] : ParseError EmptyInput)")
	if err != nil {
		t.Fatalf("zero-payload defenum variant reference should emit successfully: %v", err)
	}
	if !strings.Contains(g, "type ParseError struct {") {
		t.Errorf("defenum should emit a real, exported Go struct type: got %s", g)
	}
	if !strings.Contains(g, "func ParseError_EmptyInput() ParseError {") {
		t.Errorf("zero-payload variant should get a real, real, no-arg constructor: got %s", g)
	}
	if !strings.Contains(g, "return ParseError_EmptyInput()") {
		t.Errorf("bare zero-payload variant reference should call its own constructor: got %s", g)
	}
}

func TestEmitGoDefenumSinglePayloadVariant(t *testing.T) {
	g, err := buildGo(t, "(defenum ParseError (EmptyInput) (Invalid (msg : String)))\n"+
		"(defn too-big [] : ParseError (Invalid \"too big\"))")
	if err != nil {
		t.Fatalf("single-payload defenum variant call should emit successfully: %v", err)
	}
	if !strings.Contains(g, "func ParseError_Invalid(v0 any) ParseError {") {
		t.Errorf("single-payload variant should get a real, one-arg constructor: got %s", g)
	}
	if !strings.Contains(g, `return ParseError_Invalid("too big")`) {
		t.Errorf("payload-carrying variant call should pass its own argument through: got %s", g)
	}
}

func TestEmitGoDefenumZeroPayloadCalledWithArgIsError(t *testing.T) {
	_, err := buildGo(t, "(defenum ParseError (EmptyInput) (Invalid (msg : String)))\n"+
		"(defn bad [] : ParseError (EmptyInput \"oops\"))")
	if err == nil {
		t.Fatal("calling a zero-payload variant with an argument should be a real, honest compile error, not silently mis-emitted")
	}
}

func TestEmitGoDefenumMultiFieldVariantIsError(t *testing.T) {
	_, err := buildGo(t, "(defenum Bad (TwoFields (a : I32) (b : I32)))")
	if err == nil {
		t.Fatal("a 2+-field defenum variant is a real, deliberate v0 boundary and should error, not silently mis-emit")
	}
}

func TestEmitGoDefenumInResultErrorPosition(t *testing.T) {
	g, err := buildGo(t, "(defenum ParseError (EmptyInput) (Invalid (msg : String)))\n"+
		"(defn parse-positive [(n : I32)] : (Result I32 ParseError)\n"+
		"  (if (<= n 0) (Err EmptyInput)\n"+
		"    (if (> n 100) (Err (Invalid \"too big\")) (Ok n))))")
	if err != nil {
		t.Fatalf("a registered defenum should be usable as a Result's own ErrorType: %v", err)
	}
	if !strings.Contains(g, "func ParsePositive(n int32) Result {") {
		t.Errorf("Result return type should still resolve to the real, fixed Result Go type: got %s", g)
	}
}

// TestEmitGoDefenumEndToEndBuildsAndRuns — real, live proof, not just unit tests: a real defenum
// used as a Result's own ErrorType, constructed via both a bare zero-payload reference and a
// payload-carrying call, consumed via the same real `match` this target already supports
// (Ok-vs-Err only -- matching WHICH error variant came back is real, separate, unstarted work,
// named explicitly in emitGoDefenum's own doc comment).
func TestEmitGoDefenumEndToEndBuildsAndRuns(t *testing.T) {
	src := "(module enum-e2e)\n(export describe)\n" +
		"(defenum ParseError\n  (EmptyInput)\n  (Invalid (msg : String)))\n" +
		"(defn parse-positive [(n : I32)] : (Result I32 ParseError)\n" +
		"  (if (<= n 0) (Err EmptyInput)\n" +
		"    (if (> n 100) (Err (Invalid \"too big\")) (Ok n))))\n" +
		"(defn describe [(n : I32)] : I32\n" +
		"  (match (parse-positive n)\n" +
		"    ((Ok v) v)\n" +
		"    ((Err e) -1)))\n"

	program, err := ParseProgram(src)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if err := RegionAnalyze(program); err != nil {
		t.Fatalf("unexpected region-analyze error: %v", err)
	}
	g, err := EmitGo(program)
	if err != nil {
		t.Fatalf("unexpected emit error: %v", err)
	}

	dir := t.TempDir()
	pkgDir := dir + "/burrowgen"
	if err := os.Mkdir(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pkgDir+"/gen.go", []byte(g), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dir+"/go.mod", []byte("module enume2e\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mainSrc := `package main

import (
	"fmt"

	"enume2e/burrowgen"
)

func main() {
	fmt.Println(burrowgen.Describe(5))
	fmt.Println(burrowgen.Describe(0))
	fmt.Println(burrowgen.Describe(500))
}
`
	if err := os.WriteFile(dir+"/main.go", []byte(mainSrc), 0o644); err != nil {
		t.Fatal(err)
	}

	binPath := dir + "/bin"
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	buildCmd.Dir = dir
	buildCmd.Env = append(os.Environ(), "GOWORK=off")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("emitted Go failed to build (this is a real bug in emit_go.go, not the test): %v\n%s", err, out)
	}

	runCmd := exec.Command(binPath)
	out, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("running the built binary failed: %v\n%s", err, out)
	}
	got := string(out)
	// describe(5)=5 (real Ok), describe(0)=-1 (EmptyInput -> Err), describe(500)=-1 (Invalid -> Err).
	want := "5\n-1\n-1\n"
	if got != want {
		t.Fatalf("wrong real runtime output: got %q, want %q", got, want)
	}
}

// TestEmitGoMatchOnUserDefenum -- real, new capability (kanban card 9988's own next-named
// prerequisite after match/Result/loop/Vec/defenum: "match on a user defenum" was named as
// unstarted in every status update on this card since defenum itself shipped). A 3-variant enum
// (more than Result/Option's own fixed 2) confirms this isn't just a coincidental reuse of the
// old 2-clause path.
func TestEmitGoMatchOnUserDefenum(t *testing.T) {
	g, err := buildGo(t, "(defenum Light (Red) (Yellow) (Green))\n"+
		"(defn next-light [(x : I32)] : Light\n"+
		"  (if (= x 0) Red (if (= x 1) Yellow Green)))\n"+
		"(defn light-code [(x : I32)] : I32\n"+
		"  (match (next-light x)\n"+
		"    ((Red) 1)\n"+
		"    ((Yellow) 2)\n"+
		"    ((Green) 3)))")
	if err != nil {
		t.Fatalf("match on a real user-defenum-returning defn should emit successfully: %v", err)
	}
	if !strings.Contains(g, "NextLight(x)") {
		t.Errorf("expected the real scrutinee call: got %s", g)
	}
	if !strings.Contains(g, ".Tag == 0") || !strings.Contains(g, ".Tag == 1") {
		t.Errorf("expected real tag checks for at least 2 of the 3 real variants (the last is a plain else): got %s", g)
	}
}

// TestEmitGoMatchOnUserDefenumWithPayload -- confirms a payload-carrying variant binds correctly
// (boxed through `any`, since v0 has no per-variant payload type tracked -- see
// emitGoMatchEnum's own doc comment).
func TestEmitGoMatchOnUserDefenumWithPayload(t *testing.T) {
	g, err := buildGo(t, "(defenum ParseError (EmptyInput) (Invalid (msg : String)))\n"+
		"(defn classify [(n : I32)] : ParseError\n"+
		"  (if (= n 0) EmptyInput (Invalid \"bad\")))\n"+
		"(defn error-code [(n : I32)] : I32\n"+
		"  (match (classify n)\n"+
		"    ((EmptyInput) 1)\n"+
		"    ((Invalid msg) 2)))")
	if err != nil {
		t.Fatalf("match on a user defenum with a payload-carrying variant should emit successfully: %v", err)
	}
	if !strings.Contains(g, ".Value\n") {
		t.Errorf("expected the payload-carrying clause to bind .Value (boxed through any): got %s", g)
	}
	if !strings.Contains(g, "_ = msg") {
		t.Errorf("expected an explicit discard for the unused Invalid-clause binding: got %s", g)
	}
}

// TestEmitGoMatchOnUserDefenumNotExhaustiveIsError -- real exhaustiveness check: a 3-variant
// enum matched with only 2 clauses must be a real, honest error, unlike Result/Option's own
// fixed "exactly 2 clauses" shortcut which can't even express this gap.
func TestEmitGoMatchOnUserDefenumNotExhaustiveIsError(t *testing.T) {
	_, err := buildGo(t, "(defenum Light (Red) (Yellow) (Green))\n"+
		"(defn next-light [(x : I32)] : Light Red)\n"+
		"(defn light-code [(x : I32)] : I32\n"+
		"  (match (next-light x)\n"+
		"    ((Red) 1)\n"+
		"    ((Yellow) 2)))")
	if err == nil {
		t.Fatal("expected a real error: matching only 2 of a 3-variant enum's real variants is not exhaustive")
	}
}

// TestEmitGoMatchOnUserDefenumWrongEnumVariantIsError -- a clause naming a real variant that
// belongs to a DIFFERENT registered defenum than the scrutinee's own must be a real, honest
// error, not silently mis-emitted (the flat, global knownEnumVariants namespace makes this a
// real, checkable mistake to guard against, not just a hypothetical one).
func TestEmitGoMatchOnUserDefenumWrongEnumVariantIsError(t *testing.T) {
	_, err := buildGo(t, "(defenum Light (Red) (Green))\n"+
		"(defenum Mood (Happy) (Sad))\n"+
		"(defn next-light [(x : I32)] : Light Red)\n"+
		"(defn light-code [(x : I32)] : I32\n"+
		"  (match (next-light x)\n"+
		"    ((Red) 1)\n"+
		"    ((Happy) 2)))")
	if err == nil {
		t.Fatal("expected a real error: 'Happy' belongs to Mood, not Light -- matching it against a Light scrutinee should fail")
	}
}

// TestEmitGoMatchOnUserDefenumEndToEndBuildsAndRuns — real, live proof, not just unit tests: a
// real 3-variant defenum, matched exhaustively, compiled/linked/run for real.
func TestEmitGoMatchOnUserDefenumEndToEndBuildsAndRuns(t *testing.T) {
	src := "(module enum-match-e2e)\n(export light-code)\n" +
		"(defenum Light\n  (Red)\n  (Yellow)\n  (Green))\n" +
		"(defn next-light [(x : I32)] : Light\n" +
		"  (if (= x 0) Red (if (= x 1) Yellow Green)))\n" +
		"(defn light-code [(x : I32)] : I32\n" +
		"  (match (next-light x)\n" +
		"    ((Red) 1)\n" +
		"    ((Yellow) 2)\n" +
		"    ((Green) 3)))\n"

	program, err := ParseProgram(src)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if err := RegionAnalyze(program); err != nil {
		t.Fatalf("unexpected region-analyze error: %v", err)
	}
	g, err := EmitGo(program)
	if err != nil {
		t.Fatalf("unexpected emit error: %v", err)
	}

	dir := t.TempDir()
	pkgDir := dir + "/burrowgen"
	if err := os.Mkdir(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pkgDir+"/gen.go", []byte(g), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dir+"/go.mod", []byte("module matche2e\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mainSrc := `package main

import (
	"fmt"

	"matche2e/burrowgen"
)

func main() {
	fmt.Println(burrowgen.LightCode(0))
	fmt.Println(burrowgen.LightCode(1))
	fmt.Println(burrowgen.LightCode(2))
}
`
	if err := os.WriteFile(dir+"/main.go", []byte(mainSrc), 0o644); err != nil {
		t.Fatal(err)
	}

	binPath := dir + "/bin"
	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	buildCmd.Dir = dir
	buildCmd.Env = append(os.Environ(), "GOWORK=off")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("emitted Go failed to build (this is a real bug in emit_go.go, not the test): %v\n%s", err, out)
	}

	runCmd := exec.Command(binPath)
	out, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("running the built binary failed: %v\n%s", err, out)
	}
	got := string(out)
	// light-code(0)=1 (Red), light-code(1)=2 (Yellow), light-code(2)=3 (Green).
	want := "1\n2\n3\n"
	if got != want {
		t.Fatalf("wrong real runtime output: got %q, want %q", got, want)
	}
}
