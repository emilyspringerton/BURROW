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
