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

func cmdParse(args []string) int {
	if len(args) < 1 {
		fmt.Fprint(os.Stderr, "usage: burrow parse <file.prn>\n")
		return 1
	}
	return notImplemented("parse", "Phase 1 -- lexer parity")
}

func cmdAnalyze(args []string) int {
	if len(args) < 1 {
		fmt.Fprint(os.Stderr, "usage: burrow analyze <file.prn>\n")
		return 1
	}
	return notImplemented("analyze", "Phase 2 -- parser + region analyzer parity")
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
	return notImplemented("build", "Phase 3 -- emitter parity (C/TypeScript/Java), Phase 5 for a new native Go target")
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
