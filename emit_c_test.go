// emit_c_test.go — real, v0 verification for the new C emitter (emit_c.go), same real
// "check the emitter's own success/failure behavior directly, verify actual gcc acceptance
// separately" split PARENA's own test_emit_ts.c/test_emit_java.c already establish -- the real
// "does a real gcc actually accept and correctly run this" half is verified separately (not as a
// go test, matching that same real precedent), by actually invoking `burrow build` + `gcc` on a
// real .prn file and running the result.
package main

import (
	"strings"
	"testing"
)

func buildC(t *testing.T, src string) (string, error) {
	t.Helper()
	program, err := ParseProgram(src)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if err := RegionAnalyze(program); err != nil {
		t.Fatalf("unexpected region-analyze error: %v", err)
	}
	return EmitC(program)
}

func TestEmitCZeroArgConstant(t *testing.T) {
	// Same real shape xp_award_mod.prn's own real, first-ever PARENA mod uses.
	c, err := buildC(t, "(defn xp-award [] : I32 60)")
	if err != nil {
		t.Fatalf("zero-arg I32 constant should emit successfully: %v", err)
	}
	if !strings.Contains(c, "int xp_award(void) {") {
		t.Errorf("zero-arg defn name should be snake_cased, typed int: got %s", c)
	}
	if !strings.Contains(c, "return 60;") {
		t.Errorf("zero-arg defn body should return the real literal: got %s", c)
	}
}

func TestEmitCScalarParamIfElseBinopNestedCall(t *testing.T) {
	// Same real shape item_drop_mod.prn's own real PAPERCRAFT mod uses.
	c, err := buildC(t, "(defn material-paper [] : I32 0)\n"+
		"(defn on-item-for-object-destroyed [(material : I32)] : I32\n"+
		"  (if (= material (material-paper)) 1 0))")
	if err != nil {
		t.Fatalf("scalar param + if/else + binop + nested call should emit successfully: %v", err)
	}
	if !strings.Contains(c, "int material_paper(void)") {
		t.Errorf("first defn should be snake_cased correctly: got %s", c)
	}
	if !strings.Contains(c, "int on_item_for_object_destroyed(int material)") {
		t.Errorf("second defn's own scalar param should be typed int, real C 'Type name' order: got %s", c)
	}
	if !strings.Contains(c, "(material == material_paper())") {
		t.Errorf("= binop should lower to C's own ==, nested zero-arg call snake_cased: got %s", c)
	}
	if !strings.Contains(c, "? 1 : 0") {
		t.Errorf("if/else should lower to a real ternary: got %s", c)
	}
}

func TestEmitCNotUnaryOp(t *testing.T) {
	// Real, genuine gap found and fixed the same day writing stdlib/k8s/operator.prn ((if (not
	// exists) ...)): `not` is a real, distinct 1-argument form, not a binop -- it used to fall
	// through cBinopTable's own 2-argument-only dispatch into a bogus call to a never-defined
	// `not(...)` C function, the exact same real gap src/emit.c's own header comment already
	// documents fixing for the same reason.
	c, err := buildC(t, "(defn f [(x : Bool)] : Bool (not x))")
	if err != nil {
		t.Fatalf("not should emit successfully: %v", err)
	}
	if !strings.Contains(c, "(!(x))") {
		t.Errorf("not should lower to C's own !, not a bogus call: got %s", c)
	}
}

func TestEmitCBitwiseAndModOps(t *testing.T) {
	// Same real binop set base4/algebra.prn already proves compiles correctly through the real
	// C-based parena-c.
	c, err := buildC(t, "(defn base4-xor [(a : I32) (b : I32)] : I32 (bit-xor a b))\n"+
		"(defn base4-add [(a : I32) (b : I32)] : I32 (mod (+ a b) 4))")
	if err != nil {
		t.Fatalf("bitwise/mod ops should emit successfully: %v", err)
	}
	if !strings.Contains(c, "(a ^ b)") {
		t.Errorf("bit-xor should lower to ^: got %s", c)
	}
	if !strings.Contains(c, "((a + b) % 4)") {
		t.Errorf("mod should lower to %%: got %s", c)
	}
}

func TestEmitCForwardDeclarations(t *testing.T) {
	// A defn calling a SIBLING defn defined LATER in the file needs a real forward declaration
	// -- the one genuinely new structural piece over emit_ts.go/emit_java.c's own single-pass
	// shape.
	c, err := buildC(t, "(defn a [] : I32 (b))\n(defn b [] : I32 42)")
	if err != nil {
		t.Fatalf("forward-reference to a later sibling defn should emit successfully: %v", err)
	}
	declIdx := strings.Index(c, "int b(void);")
	defIdx := strings.Index(c, "int a(void) {")
	if declIdx < 0 {
		t.Fatalf("expected a real forward declaration for b(): got %s", c)
	}
	if declIdx > defIdx {
		t.Errorf("forward declaration should come before a()'s own definition: decl at %d, a() def at %d", declIdx, defIdx)
	}
}

func TestEmitCWrongArityBinopIsError(t *testing.T) {
	_, err := buildC(t, "(defn f [] : I32 (+ 1 2 3))")
	if err == nil {
		t.Error("a binary operator called with the wrong arity should be a real, honest error, not silently accepted")
	}
}

func TestEmitCArenaParamIsError(t *testing.T) {
	_, err := buildC(t, "(defn f [(buf : Arena @ :region/scratch)] : I32 0)")
	if err == nil {
		t.Error("an Arena/region-annotated parameter should be a real, honest unsupported error, not silently guessed")
	}
}

func TestEmitCUnknownNamespacedIdentifierIsError(t *testing.T) {
	// Real, genuine bug found and fixed while testing this emitter for real against
	// stdlib/mishri/humanness.prn: a bare `math/pi` reference used to silently mangle straight
	// through into broken C syntax (a literal `/` in an identifier). Cross-checked directly
	// against the real reference: `parena build` on the same real input reports
	// "unknown identifier 'math/pi' at line N" -- this v0's own C target genuinely has no
	// math/* primitive support (matching STDLIB.md's own documented gap), so the real, honest
	// behavior is to reject it the same way, not silently emit something gcc would choke on.
	_, err := buildC(t, "(defn f [] : F64 math/pi)")
	if err == nil {
		t.Fatal("a real, unknown namespaced identifier should be a real, honest error, not silently mangled through")
	}
	if !strings.Contains(err.Error(), "unknown identifier 'math/pi'") {
		t.Errorf("error should name the real, exact unknown identifier: got %v", err)
	}
}

func TestEmitCUnknownCallIsError(t *testing.T) {
	_, err := buildC(t, "(defn f [] : F64 (math/random))")
	if err == nil {
		t.Fatal("a call to a real, unknown function should be a real, honest error, not silently mangled through")
	}
	if !strings.Contains(err.Error(), "unknown identifier 'math/random'") {
		t.Errorf("error should name the real, exact unknown call: got %v", err)
	}
}

