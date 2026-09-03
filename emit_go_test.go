// emit_go_test.go — real v0 verification for the new Go emitter (emit_go.go), same real split
// emit_c_test.go already establishes: check the emitter's own success/failure behavior directly
// here, verify actual `go build` acceptance of the emitted output separately (not as a go test),
// by actually invoking `burrow build -o *.go` + `go build` on real .prn input and running it.
package main

import (
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
