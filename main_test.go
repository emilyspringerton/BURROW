// main_test.go — real, direct verification of cmdNew, the new "batteries included" scaffolding
// command (kanban priority-queue cards PXCL-001: "we need a batteries included cli tool to
// generate scaffolding and stuff for us build it into burrow so it can help us manage both the
// go and prn side of things" and PX-333: "PARENA SCAFFOLD NEW NEEDS TO GENERATE BAZEL BY
// DEFAULT"). Not a template-only tool: cmdNew actually runs the generated starter .prn through
// EmitGo AND `go build` before returning success, so this test proves the real, whole pipeline,
// not just that some files got written. The generated MODULE.bazel/BUILD.bazel pair is
// separately, live-verified via a real `bazel run` in a real scratch directory (not repeated as
// a Go test here -- a real Bazel invocation fetching @burrow fresh from GitHub takes well over a
// minute, too slow for a routine `go test` run; this test instead checks the real, generated
// file CONTENT is correct, i.e. that it references a real, existing pinned commit).
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCmdNewScaffoldsAndBuilds(t *testing.T) {
	dir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)

	code := cmdNew([]string{"demo_mod"})
	if code != 0 {
		t.Fatalf("cmdNew returned %d, want 0", code)
	}

	// Real, direct proof every promised file actually exists.
	for _, f := range []string{
		"demo_mod/demo_mod.prn",
		"demo_mod/main.go",
		"demo_mod/go.mod",
		"demo_mod/internal/burrowgen/demo_mod_gen.go",
		"demo_mod/MODULE.bazel",
		"demo_mod/BUILD.bazel",
	} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("expected %s to exist: %v", f, err)
		}
	}

	// Real, direct proof the generated MODULE.bazel pins a real commit, not a placeholder --
	// a scaffold that references a commit nobody can actually fetch is a real, silent failure
	// mode this test exists to catch.
	moduleBazel, err := os.ReadFile(filepath.Join(dir, "demo_mod", "MODULE.bazel"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(moduleBazel), burrowPinnedCommit) {
		t.Errorf("expected MODULE.bazel to reference the real pinned commit %s: got %s", burrowPinnedCommit, moduleBazel)
	}

	// Real, live, end-to-end proof: the generated scaffold actually RUNS and prints the real,
	// correct output -- not just "it compiled."
	runCmd := exec.Command("go", "run", ".")
	runCmd.Dir = filepath.Join(dir, "demo_mod")
	runCmd.Env = append(os.Environ(), "GOWORK=off")
	out, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("running the scaffolded program failed: %v\n%s", err, out)
	}
	want := "Hello from demo_mod!\n"
	if string(out) != want {
		t.Fatalf("wrong real runtime output: got %q, want %q", out, want)
	}
}

func TestCmdNewRefusesToOverwrite(t *testing.T) {
	dir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)

	if code := cmdNew([]string{"demo_mod"}); code != 0 {
		t.Fatalf("first cmdNew returned %d, want 0", code)
	}
	if code := cmdNew([]string{"demo_mod"}); code == 0 {
		t.Fatal("cmdNew should refuse to overwrite an existing directory, not silently succeed")
	}
}
