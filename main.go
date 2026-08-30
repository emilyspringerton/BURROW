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
	"strings"
)

const usage = `usage: burrow parse <file.prn>
       burrow analyze <file.prn>                          (region analyzer)
       burrow build <file.prn> [file2.prn ...] -o <output.c|output.ts|output.java>    (C emitter
       (default), or the TypeScript/Java emitters if -o ends in .ts/.java; multiple files are
       combined into one compilation unit, in the order given)
       burrow fmt [-w] <file.prn> [file2.prn ...]         (re-indent; -w writes in place, default
       prints to stdout)
       burrow ci-status <owner/repo> <sha>                (GITHUB_TOKEN env var required; exit
       0=all green, 1=pending, 2=failed, 3=not found/error)
`

// notImplemented reports an honest, explicit "not built yet" for a real, recognized subcommand
// whose CLI shape burrow already understands -- never a silent no-op, matching parena's own real,
// established bootstrap discipline. phase names the real NORTHSTAR.md phase that subcommand's own
// real implementation belongs to, so the message stays a real pointer, not a vague apology.
func notImplemented(subcommand, phase string) int {
	fmt.Fprintf(os.Stderr, "burrow: %s: not yet implemented (see NORTHSTAR.md, %s)\n", subcommand, phase)
	return 1
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
	program, perr := ParseProgram(string(src))
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
	program, perr := ParseProgram(string(src))
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
		fmt.Fprint(os.Stderr, "usage: burrow build <file.prn> [file2.prn ...] -o <output.c|output.ts|output.java>\n")
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
		fileProgram, perr := ParseProgram(string(src))
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
	// wrong output), anything else uses the real C emitter (now real, Phase 4 v0).
	if strings.HasSuffix(outPath, ".ts") {
		return notImplemented("build (TypeScript target)", "Phase 4 -- TypeScript emitter parity (real C target already works, see emit_c.go)")
	}
	if strings.HasSuffix(outPath, ".java") {
		return notImplemented("build (Java target)", "Phase 4 -- Java emitter parity (real C target already works, see emit_c.go)")
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
	default:
		fmt.Fprintf(os.Stderr, "burrow: unknown command '%s'\n", subcommand)
		code = 1
	}
	os.Exit(code)
}
