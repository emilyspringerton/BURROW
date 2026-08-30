// region_test.go — real port of tests/test_region.c's own real test scenarios (VS0 DoD domain 2's
// own required "1 positive + 1 negative case," plus the same real edge cases that file's own
// header comment names) into idiomatic Go subtests. Every expectation here, including the exact
// literal error message text, is copied verbatim from that file's own real assertions.
package main

import "testing"

func analyzeSrc(t *testing.T, src string) error {
	t.Helper()
	program, err := ParseProgram(src)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	return RegionAnalyze(program)
}

func TestRegionValidScratchToBufferPromotion(t *testing.T) {
	// DoD's own required positive case: test.prn's load-config.
	err := analyzeSrc(t, "(defn load-config [(buf-arena : Arena @ :region/buffer)]\n"+
		"  (with-arena [scratch :region/scratch 1024]\n"+
		"    (let [temp-str (alloc scratch String \"config.json\")\n"+
		"          real-buf (alloc buf-arena String \"parsed_data\")]\n"+
		"      real-buf)))")
	if err != nil {
		t.Errorf("valid scratch-to-buffer promotion (test.prn's load-config) should analyze clean: %v", err)
	}
}

func TestRegionInvalidScratchIntoBufferEscape(t *testing.T) {
	// DoD's own required negative case: test.prn's break-safety, checked against the exact
	// literal error format NORTHSTAR.md's own DoD table specifies.
	err := analyzeSrc(t, "(defn break-safety [(buf-arena : Arena @ :region/buffer)]\n"+
		"  (with-arena [scratch :region/scratch 1024]\n"+
		"    (let [bad-str (alloc scratch String \"escaped_memory\")]\n"+
		"      (buffer/set-data buf-arena bad-str))))")
	if err == nil {
		t.Fatal("invalid scratch-into-buffer escape (test.prn's break-safety) should be rejected")
	}
	want := "Compile Error: Escaping region pointer from :region/scratch to :region/buffer at line 4"
	if err.Error() != want {
		t.Errorf("error message should match NORTHSTAR.md's own DoD table exactly, including line number: got %q want %q",
			err.Error(), want)
	}
}

func TestRegionValidFunctionDoesNotPoisonInvalidOne(t *testing.T) {
	// Both functions in one file, as the real test.prn ships them: the first (valid) function
	// must not somehow poison analysis of the second.
	err := analyzeSrc(t, "(defn load-config [(buf-arena : Arena @ :region/buffer)]\n"+
		"  (with-arena [scratch :region/scratch 1024]\n"+
		"    (let [real-buf (alloc buf-arena String \"parsed_data\")]\n"+
		"      real-buf)))\n"+
		"(defn break-safety [(buf-arena : Arena @ :region/buffer)]\n"+
		"  (with-arena [scratch :region/scratch 1024]\n"+
		"    (let [bad-str (alloc scratch String \"escaped_memory\")]\n"+
		"      (buffer/set-data buf-arena bad-str))))")
	if err == nil {
		t.Error("a valid function followed by an invalid one should still catch the real violation")
	}
}

func TestRegionSameRankAssignmentNotFalsePositive(t *testing.T) {
	// Edge case: same-rank assignment (buffer into buffer) is not a false positive --
	// Region(Source) >= Region(Destination) holds when they're equal, not just when Source is
	// strictly longer-lived.
	err := analyzeSrc(t, "(defn same-rank [(a : Arena @ :region/buffer) (b : Arena @ :region/buffer)]\n"+
		"  (buffer/set-data a b))")
	if err != nil {
		t.Errorf("same-rank assignment (buffer into buffer) should not be a false positive: %v", err)
	}
}

func TestRegionSafeDirectionNotFlagged(t *testing.T) {
	// Edge case: promoting a longer-lived value into a shorter-lived slot is always safe and
	// must not be flagged -- only the reverse direction is the real invariant violation.
	err := analyzeSrc(t, "(defn safe-direction [(buf-arena : Arena @ :region/buffer)]\n"+
		"  (with-arena [scratch :region/scratch 1024]\n"+
		"    (buffer/set-data scratch buf-arena)))")
	if err != nil {
		t.Errorf("assigning a longer-lived (buffer) value into a shorter-lived (scratch) slot should be safe: %v", err)
	}
}

func TestRegionUnconstrainedLetBindingNotFalsePositive(t *testing.T) {
	// Edge case: a non-alloc let binding (unconstrained rank) must not be treated as an
	// automatic violation just because it has no known region.
	err := analyzeSrc(t, "(defn plain-let [(buf-arena : Arena @ :region/buffer)]\n"+
		"  (let [n 42]\n"+
		"    (buffer/set-data buf-arena n)))")
	if err != nil {
		t.Errorf("an unconstrained (non-alloc) let binding should not be falsely flagged as an escape: %v", err)
	}
}

func TestRegionNestedWithArenaEscapeCaught(t *testing.T) {
	// Edge case: nested with-arena, escape from the innermost scope into the outermost still
	// gets caught through two scope levels.
	err := analyzeSrc(t, "(defn nested [(buf-arena : Arena @ :region/buffer)]\n"+
		"  (with-arena [scratch :region/scratch 1024]\n"+
		"    (with-arena [inner :region/scratch 512]\n"+
		"      (let [bad (alloc inner String \"x\")]\n"+
		"        (buffer/set-data buf-arena bad)))))")
	if err == nil {
		t.Error("an escape from a nested with-arena scope should still be caught")
	}
}
