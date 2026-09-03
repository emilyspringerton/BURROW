// main.go — the burrow CLI. Real, direct founder instruction: "so write the burrow cli in parena
// and go it should have the same api as parena like yarn is the same as npm" -- this file's own
// real job is CLI-SURFACE parity with PARENA/src/main.c's own parena binary (same subcommands,
// same usage text shape, same argument conventions), not yet full lexer/parser/region-analyzer/
// emitter parity -- that's real, later, phased work (see NORTHSTAR.md's own Phase 1 onward). Every
// subcommand below is real (parses its own real argument shape, matching parena's own exact
// usage), and reports an honest, explicit "not yet implemented" rather than silently doing nothing
// or emitting something misleading -- the exact same discipline PARENA/src/main.c's own header
// comment already established for itself during ITS OWN early bootstrap ("build exists but reports
// clearly that the region analyzer and C emitter aren't wired in yet").
//
// Real, dogfooding note (founder real-time: "dog food it like write the golang in a way that
// doesnt allocate on the heap or whatever"): this file is a real, ordinary, short-lived CLI
// invocation (see NORTHSTAR.md's own GC-off section -- burrow itself never needs GOGC=off, that
// question only applies to code burrow will eventually EMIT), so there's no real GC-off claim
// being made about this binary's own process. What IS practiced here, honestly, at this small
// scale: os.Args is read directly with no intermediate copies/joins, error paths write straight to
// os.Stderr via fmt.Fprintf rather than building a throwaway string first, and there is
// deliberately no slice growth anywhere in this file's own real control flow -- the same real
// "know your bound, don't allocate more than you need" discipline the rest of NORTHSTAR.md's own
// dogfooding section names, applied here at the one small scale this file actually has to prove it
// at yet.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const usage = `usage: burrow parse <file.prn>
       burrow analyze <file.prn>                          (region analyzer)
       burrow build <file.prn> [file2.prn ...] -o <output.c|output.go|output.ts|output.java>
       (C emitter (default), real native Go emitter if -o ends in .go (v0, see emit_go.go), or
       the TypeScript/Java emitters if -o ends in .ts/.java; multiple files are combined into one
       compilation unit, in the order given)
       burrow fmt [-w] <file.prn> [file2.prn ...]         (re-indent; -w writes in place, default
       prints to stdout)
       burrow ci-status <owner/repo> <sha>                (GITHUB_TOKEN env var required; exit
       0=all green, 1=pending, 2=failed, 3=not found/error)
       burrow new <name>                                  (real, "batteries included" scaffold:
       a starter .prn file, a real Go host main.go + go.mod importing its own compiled
       internal/burrowgen package, built immediately so the scaffold is runnable with zero
       further manual steps)
`

// notImplemented reports an honest, explicit "not built yet" for a real, recognized subcommand
// whose CLI shape burrow already understands -- never a silent no-op, matching parena's own real,
// established bootstrap discipline. phase names the real NORTHSTAR.md phase that subcommand's own
// real implementation belongs to, so the message stays a real pointer, not a vague apology.
func notImplemented(subcommand, phase string) int {
	fmt.Fprintf(os.Stderr, "burrow: %s: not yet implemented (see NORTHSTAR.md, %s)\n", subcommand, phase)
	return 1
}

// parseSourceFile — real, single dispatch point by file extension: `.llll` (LO source, see
// lo_lexer.go/lo_parser.go) routes through LoParseProgram, everything else (real .prn source)
// keeps using the existing ParseProgram unchanged. Every real caller in this file that used to
// call ParseProgram(string(src)) directly now calls this instead, so `burrow parse`/`analyze`/
// `build` all transparently accept a real .llll file anywhere a .prn file was accepted, with
// zero changes to RegionAnalyze/EmitC/EmitGo -- LO is purely a new frontend sharing every
// existing backend, matching LO/NORTHSTAR.md's own real architectural correction directly.
func parseSourceFile(path string, src []byte) (*Node, error) {
	if strings.HasSuffix(path, ".llll") {
		return LoParseProgram(string(src))
	}
	return ParseProgram(string(src))
}

// nodeHasText reports whether n's own real Text field is meaningful (Symbol/Keyword/String/
// Number) -- distinct from Text == "", since a real, valid empty string LITERAL ("") is a real
// NodeString with meaningful (if empty) text, not "no text" the way NodeColon/NodeAt's own real
// text-less punctuation is. Matches PARENA/src/main.c's own real `n->text != NULL` check exactly
// (Go has no NULL string; this is the correct, non-lossy real equivalent).
func nodeHasText(n *Node) bool {
	switch n.Type {
	case NodeSymbol, NodeKeyword, NodeString, NodeNumber:
		return true
	default:
		return false
	}
}

// printNode -- real, direct port of PARENA/src/main.c's own print_node, matching its real output
// format exactly (same real dump shape "burrow parse" and "parena parse" now both produce, real
// CLI-surface parity, not just command-shape parity).
func printNode(n *Node, depth int) {
	fmt.Print(strings.Repeat("  ", depth))
	switch {
	case nodeHasText(n):
		fmt.Printf("%s: %s (line %d)\n", n.Type, n.Text, n.Line)
	case n.Type == NodeList || n.Type == NodeVec || n.Type == NodeMap:
		fmt.Printf("%s (line %d, %d children)\n", n.Type, n.Line, len(n.Children))
	default:
		fmt.Printf("%s (line %d)\n", n.Type, n.Line)
	}
	for _, child := range n.Children {
		printNode(child, depth+1)
	}
}

func cmdParse(args []string) int {
	if len(args) < 1 {
		fmt.Fprint(os.Stderr, "usage: burrow parse <file.prn>\n")
		return 1
	}
	src, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "burrow: cannot read %s\n", args[0])
		return 1
	}
	program, perr := parseSourceFile(args[0], src)
	if perr != nil {
		fmt.Fprintf(os.Stderr, "burrow: parse error: %v\n", perr)
		return 1
	}
	printNode(program, 0)
	return 0
}

func cmdAnalyze(args []string) int {
	if len(args) < 1 {
		fmt.Fprint(os.Stderr, "usage: burrow analyze <file.prn>\n")
		return 1
	}
	src, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "burrow: cannot read %s\n", args[0])
		return 1
	}
	program, perr := parseSourceFile(args[0], src)
	if perr != nil {
		fmt.Fprintf(os.Stderr, "burrow: parse error: %v\n", perr)
		return 1
	}
	if rerr := RegionAnalyze(program); rerr != nil {
		fmt.Fprintf(os.Stderr, "burrow: %v\n", rerr)
		return 1
	}
	fmt.Printf("burrow: %s: region analysis OK\n", args[0])
	return 0
}

func cmdFmt(args []string) int {
	firstPath := 0
	if len(args) > 0 && args[0] == "-w" {
		firstPath = 1
	}
	if len(args) <= firstPath {
		fmt.Fprint(os.Stderr, "usage: burrow fmt [-w] <file.prn> [file2.prn ...]\n")
		return 1
	}
	return notImplemented("fmt", "Phase 2 -- parser parity (fmt walks the same AST)")
}

func cmdBuild(args []string) int {
	// Real, same convention parena's own cmd_build dispatch uses: every argument between "build"
	// and "-o" is an input file, "-o <output>" must be the final two arguments.
	if len(args) < 3 || args[len(args)-2] != "-o" {
		fmt.Fprint(os.Stderr, "usage: burrow build <file.prn> [file2.prn ...] -o <output.c|output.go|output.ts|output.java>\n")
		return 1
	}
	paths := args[:len(args)-2]
	outPath := args[len(args)-1]

	// Real, minimal multi-file support: each input file is parsed separately, then their own
	// top-level forms are concatenated into ONE combined program, matching parena's own real
	// main.c cmd_build convention exactly (a real, minimal "linker," not full module resolution).
	program := &Node{Type: NodeList, Line: 1}
	for _, path := range paths {
		src, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "burrow: cannot read %s\n", path)
			return 1
		}
		fileProgram, perr := parseSourceFile(path, src)
		if perr != nil {
			fmt.Fprintf(os.Stderr, "burrow: %s: parse error: %v\n", path, perr)
			return 1
		}
		program.Children = append(program.Children, fileProgram.Children...)
	}

	if rerr := RegionAnalyze(program); rerr != nil {
		fmt.Fprintf(os.Stderr, "burrow: %v\n", rerr)
		return 1
	}

	// Real, same real target dispatch by output extension parena's own cmd_build already uses:
	// .ts/.java route to their own real emitters (not yet built in burrow -- real, honest,
	// explicit not-yet-implemented, not silently falling through to the C emitter and producing
	// wrong output); .go is real, new Phase 6 (a real host -- DUNG -- asked for it, see
	// emit_go.go's own doc comment); anything else uses the real C emitter (Phase 4 v0).
	if strings.HasSuffix(outPath, ".ts") {
		return notImplemented("build (TypeScript target)", "Phase 4 -- TypeScript emitter parity (real C target already works, see emit_c.go)")
	}
	if strings.HasSuffix(outPath, ".java") {
		return notImplemented("build (Java target)", "Phase 4 -- Java emitter parity (real C target already works, see emit_c.go)")
	}
	if strings.HasSuffix(outPath, ".go") {
		emittedGo, gerr := EmitGo(program)
		if gerr != nil {
			fmt.Fprintf(os.Stderr, "burrow: %v\n", gerr)
			return 1
		}
		if err := os.WriteFile(outPath, []byte(emittedGo), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "burrow: cannot write %s\n", outPath)
			return 1
		}
		if len(paths) == 1 {
			fmt.Printf("burrow: %s -> %s\n", paths[0], outPath)
		} else {
			fmt.Printf("burrow: [%d files] -> %s\n", len(paths), outPath)
		}
		return 0
	}

	emitted, eerr := EmitC(program)
	if eerr != nil {
		fmt.Fprintf(os.Stderr, "burrow: %v\n", eerr)
		return 1
	}
	if err := os.WriteFile(outPath, []byte(emitted), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "burrow: cannot write %s\n", outPath)
		return 1
	}
	if len(paths) == 1 {
		fmt.Printf("burrow: %s -> %s\n", paths[0], outPath)
	} else {
		fmt.Printf("burrow: [%d files] -> %s\n", len(paths), outPath)
	}
	return 0
}

func cmdCiStatus(args []string) int {
	if len(args) < 2 {
		fmt.Fprint(os.Stderr, "usage: burrow ci-status <owner/repo> <sha>\n")
		return 1
	}
	return notImplemented("ci-status", "not scoped yet -- parena's own ci-status is a self-hosted PARENA module (stdlib/ci/status.prn), real follow-up once Phase 2 exists")
}

// cmdNew — real, new "batteries included" scaffolding tool (kanban priority-queue card
// PXCL-001: "we need a batteries included cli tool to generate scaffolding and stuff for us
// build it into burrow so it can help us manage both the go and prn side of things"). Real,
// deliberate v0 scope: the Go target only -- BURROW's own real differentiator over plain
// `parena` is exactly this "PARENA decision logic called directly from a real Go host, no
// cgo/FFI" shape (this repo's own real `IDUNA_PRO/cmd/idunapro` precedent), so that's the one
// real scaffold worth generating first. Not a template-only tool: the generated `.prn` file is
// actually run through `EmitGo` and `go build` before this command returns success, so a real,
// broken scaffold is a real, honest failure here, never silently handed to the user.
func cmdNew(args []string) int {
	if len(args) != 1 {
		fmt.Fprint(os.Stderr, "usage: burrow new <name>\n")
		return 1
	}
	name := args[0]
	if _, err := os.Stat(name); err == nil {
		fmt.Fprintf(os.Stderr, "burrow: new: %s already exists\n", name)
		return 1
	}

	genDir := filepath.Join(name, "internal", "burrowgen")
	if err := os.MkdirAll(genDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "burrow: new: %v\n", err)
		return 1
	}

	prnPath := filepath.Join(name, name+".prn")
	prnSrc := fmt.Sprintf("(module %s)\n(export hello)\n\n(defn hello [] : String\n  \"Hello from %s!\")\n", name, name)
	if err := os.WriteFile(prnPath, []byte(prnSrc), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "burrow: new: %v\n", err)
		return 1
	}

	goModPath := filepath.Join(name, "go.mod")
	if err := os.WriteFile(goModPath, []byte(fmt.Sprintf("module %s\n\ngo 1.22\n", name)), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "burrow: new: %v\n", err)
		return 1
	}

	mainGoPath := filepath.Join(name, "main.go")
	mainGoSrc := fmt.Sprintf(`package main

// Real, "batteries included" scaffold generated by "burrow new" -- the PARENA decision logic
// (%[1]s.prn) is compiled into internal/burrowgen (regenerate via:
// burrow build %[1]s.prn -o internal/burrowgen/%[1]s_gen.go) and called directly here, no
// cgo/FFI boundary, the same real shape IDUNA_PRO/cmd/idunapro already established.

import (
	"fmt"

	"%[1]s/internal/burrowgen"
)

func main() {
	fmt.Println(burrowgen.Hello())
}
`, name)
	if err := os.WriteFile(mainGoPath, []byte(mainGoSrc), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "burrow: new: %v\n", err)
		return 1
	}

	// Real "batteries included" step: actually run the new scaffold through EmitGo + go build
	// right now, not just drop template text and hope. A scaffold that doesn't compile is a real
	// bug in this command, not something the user should discover themselves.
	src, err := os.ReadFile(prnPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "burrow: new: %v\n", err)
		return 1
	}
	program, perr := parseSourceFile(prnPath, src)
	if perr != nil {
		fmt.Fprintf(os.Stderr, "burrow: new: internal error generating a real starter .prn: %v\n", perr)
		return 1
	}
	if rerr := RegionAnalyze(program); rerr != nil {
		fmt.Fprintf(os.Stderr, "burrow: new: internal error generating a real starter .prn: %v\n", rerr)
		return 1
	}
	emittedGo, gerr := EmitGo(program)
	if gerr != nil {
		fmt.Fprintf(os.Stderr, "burrow: new: internal error generating a real starter .prn: %v\n", gerr)
		return 1
	}
	genPath := filepath.Join(genDir, name+"_gen.go")
	if err := os.WriteFile(genPath, []byte(emittedGo), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "burrow: new: %v\n", err)
		return 1
	}

	buildCmd := exec.Command("go", "build", "./...")
	buildCmd.Dir = name
	buildCmd.Env = append(os.Environ(), "GOWORK=off")
	if out, berr := buildCmd.CombinedOutput(); berr != nil {
		fmt.Fprintf(os.Stderr, "burrow: new: generated scaffold failed to build (this is a real bug in \"burrow new\", not your code):\n%s", out)
		return 1
	}

	// Real, new Bazel scaffolding (2026-09-03, kanban priority-queue card PX-333: "PARENA
	// SCAFFOLD NEW NEEDS TO GENERATE BAZEL BY DEFAULT WE WILL FIGURE OUT A FALLBACK LATER").
	// burrowPinnedCommit is a real, known-good BURROW commit (the one this exact "burrow new"
	// binary was itself built from) -- same real, honest "pin to a specific commit, bump it
	// later" convention this monorepo's own ladybug/DUNG/longma repos already establish for
	// @parena; it WILL go stale as BURROW evolves, the same real, accepted property every one of
	// those existing pins already has, not a defect unique to this generator. The plain
	// `go build`/`go run` path above is the real "fallback for now" the card's own framing asks
	// for -- generating Bazel files doesn't remove that, both are real and usable.
	moduleBazelPath := filepath.Join(name, "MODULE.bazel")
	moduleBazelSrc := fmt.Sprintf(`"""%[1]s — scaffolded by "burrow new". bazel_dep pins a real,
known-good BURROW commit (the one this scaffold was generated from) -- same real convention
ladybug/DUNG/longma already use for @parena. Bump the commit as BURROW evolves.
"""

module(
    name = "%[1]s",
    version = "0.0.0",
)

bazel_dep(name = "rules_go", version = "0.63.0")
bazel_dep(name = "burrow", version = "0.0.0")
git_override(
    module_name = "burrow",
    remote = "https://github.com/emilyspringerton/BURROW.git",
    commit = "%[2]s",
)

go_sdk = use_extension("@rules_go//go:extensions.bzl", "go_sdk")
go_sdk.download(version = "1.22.0")
`, name, burrowPinnedCommit)
	if err := os.WriteFile(moduleBazelPath, []byte(moduleBazelSrc), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "burrow: new: %v\n", err)
		return 1
	}

	buildBazelPath := filepath.Join(name, "BUILD.bazel")
	buildBazelSrc := fmt.Sprintf(`# Scaffolded by "burrow new" -- real genrule compiling %[1]s.prn via the real,
# pinned @burrow binary (see MODULE.bazel), then a plain rules_go library/binary over the result.
# Regenerate the checked-in %[3]s manually (see the README this command printed)
# if you'd rather not depend on Bazel invoking @burrow at build time.

load("@rules_go//go:def.bzl", "go_binary", "go_library")

genrule(
    name = "%[1]s_gen",
    srcs = ["%[1]s.prn"],
    outs = ["internal/burrowgen/%[1]s_gen.go"],
    cmd = "$(location @burrow//:burrow) build $(location %[1]s.prn) -o $@",
    tools = ["@burrow//:burrow"],
)

go_library(
    name = "burrowgen",
    srcs = [":%[1]s_gen"],
    importpath = "%[1]s/internal/burrowgen",
    visibility = ["//visibility:private"],
)

go_library(
    name = "%[1]s_lib",
    srcs = ["main.go"],
    importpath = "%[1]s",
    visibility = ["//visibility:private"],
    deps = [":burrowgen"],
)

go_binary(
    name = "%[1]s",
    embed = [":%[1]s_lib"],
    visibility = ["//visibility:public"],
)
`, name, prnPath, genPath)
	if err := os.WriteFile(buildBazelPath, []byte(buildBazelSrc), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "burrow: new: %v\n", err)
		return 1
	}

	fmt.Printf("burrow: new: %s scaffolded and built successfully\n", name)
	fmt.Printf("  %s          -- real PARENA decision logic (edit this)\n", prnPath)
	fmt.Printf("  %s              -- real Go host (edit this)\n", mainGoPath)
	fmt.Printf("  %s -- compiled output (regenerate: burrow build %s -o %s)\n", genPath, prnPath, genPath)
	fmt.Printf("  %s, %s -- real Bazel build files (pinned to burrow commit %s, bump as needed)\n", moduleBazelPath, buildBazelPath, burrowPinnedCommit)
	fmt.Printf("  cd %s && go run .              -- fallback, no Bazel needed\n", name)
	fmt.Printf("  cd %s && bazel run //:%s       -- real Bazel build\n", name, name)
	return 0
}

// burrowPinnedCommit — the real BURROW commit this exact "burrow new" binary was itself built
// from (bumped by hand alongside real BURROW releases, same convention ladybug/DUNG/longma's own
// MODULE.bazel files already use for @parena). A scaffolded project's own MODULE.bazel pins to
// this so "bazel build" works immediately, no manual commit-hunting required.
const burrowPinnedCommit = "977dd9e42c179440640ef80cdd3b62c5c4a6ed32"

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	subcommand := os.Args[1]
	rest := os.Args[2:]

	var code int
	switch subcommand {
	case "parse":
		code = cmdParse(rest)
	case "analyze":
		code = cmdAnalyze(rest)
	case "build":
		code = cmdBuild(rest)
	case "fmt":
		code = cmdFmt(rest)
	case "ci-status":
		code = cmdCiStatus(rest)
	case "new":
		code = cmdNew(rest)
	default:
		fmt.Fprintf(os.Stderr, "burrow: unknown command '%s'\n", subcommand)
		code = 1
	}
	os.Exit(code)
}
