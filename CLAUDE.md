# BURROW

## What this is

A feature-for-feature parallel implementation of the PARENA compiler (`parena-c`'s own real
lexer → parser → region analyzer → every emit target: C, TypeScript, Java), written in a real
combination of Go and PARENA itself — corrected in real time from an initial, narrower mis-scope
("just a new Go emission target bolted onto the existing C compiler"). New repo (2026-08-30,
founder real-time: "can we build project BURROW the golang emitter where we design it to run with
the gc turned off - upstream GITHUB created" → "project burrow is a feature for feature rewrite of
the parena compiler in golang and parena" → "like the first test is can we write a pure golang and
parena tool that still pass all that parena c tests" → "dog food it like write the golang in a way
that doesnt allocate on the heap or whatever" → "use parena primatives adding to the stdlibs when
it helps"). **Read `NORTHSTAR.md` before writing any code** — it has the full real scoping pass:
the founder-named acceptance bar (pass `parena-c`'s own real test corpus), the real relationship
to PARENA's own already-in-progress self-hosting effort (`selfhost/*.prn`), the GC-off design for
the new native Go emission target this project would add, and the dogfooding directive extending
that same low-allocation discipline to `burrow`'s own Go implementation.

## Status

Real CLI shell shipped (`main.go`, full command parity with `parena`'s own surface). Phases 1-4
(lexer, parser, region analyzer, and a real v0 C emitter) shipped: `lexer.go`/`parser.go`/
`region.go`/`emit_c.go` + their real test files, faithful hand-ports of `PARENA/src/lexer.c`/
`parser.c`+`ast.h`/`region.c`, plus a new v0 C emitter matching `emit_ts.c`/`emit_java.c`'s own
proven narrow scope — verified against all 30 real assertions ported from PARENA's own C reference
test suite, a full real-corpus stress test (111 `.prn` files parse+region-analyze clean), and a
real end-to-end proof (`burrow build` → real `gcc` compile → real, correct behavior against
`level_mod_test.c`'s own assertions). **Phase 6 (real, native Go emission target) also shipped**
(2026-08-30, `emit_go.go`): the same narrow v0 scope as Phase 4's C target, GC-off-safe by
construction (scalar params, one-expression bodies, no heap allocation reachable), `-o *.go` on
`burrow build`. Real, Go-specific structural pieces the C target didn't need: kebab-case →
exported PascalCase names (the emitted package is meant to be imported by a real host, not just
linked), Go's own real scalar types (`I32`→`int32`, not C's platform `int`), `if` lowered to an
immediately-invoked func literal (Go has no `?:`) with every branch value explicitly converted to
the defn's own declared return type before being boxed through `any` — two real, genuine bugs
found and fixed live while testing a real nested-if probe (a bare `I32` literal defaults to Go's
`int` when boxed through `any`, not `int32`; a nested `if`'s own result is `any`-typed and needs a
type ASSERTION at its use site, not the same `T(...)` CONVERSION a concrete value needs), and a
final `go/format.Source` pass so emitted output is unconditionally gofmt-clean regardless of
nesting depth. `go test`: 48/48 (38 prior + 10 new). `burrow` is installed on `PATH`. Full
`emit.c` parity (`defstruct`/`defenum`/`match`/`loop`/`Result`/`Vec`) and the TypeScript/Java
targets remain unstarted.

**`DUNG` is its own separate repo** (`github.com/emilyspringerton/DUNG`, first scoped inside this
repo as `DUNG.md`, corrected by the founder into its own standalone, Bazel-built repo) — "the
BURROW editor," a unified terminal emulator + editor rewriting `PITVIPER` and PARENA's own
`stdlib/editor/*.prn`. Real, load-bearing relationship, now closing the full loop: DUNG's own
`cmd/dung/main.go` calls directly into `internal/burrowgen` (checked-in output of
`burrow build parena/entry.prn -o internal/burrowgen/entry_gen.go`, Phase 6's own real Go target)
for its split-layout and focus-cycling decision logic — no cgo/FFI boundary needed, the exact real
advantage a native Go emission target has over the C target for a Go-hosted consumer. DUNG is
this target's own real, live, first host, not `GoblinFoxDragon` (still just a named candidate, not
committed to) — Phase 6's own "no real host has asked for this yet" precondition is real, past
tense now.

## Related Repos

- `PARENA` — the language and compiler this is a fourth compilation target for; `src/emit_ts.c`
  and `src/emit_java.c` are the real, proven v0 emitter template this project follows.
- `DUNG` — the real, separate repo whose own build depends on this repo's own emit capability.
- `GoblinFoxDragon` — the named candidate real Go host for a GC-off-safe PARENA-compiled decision
  layer (Phase 6 in `NORTHSTAR.md`, not committed to yet).
- `EMILY` — RSI loop / backlog coordination for cross-repo work.

## Founder Real-Time Direction

Whenever the founder gives real-time direction — a new ask, a correction, a "can we also..." —
route it through `emily observe -s info "Founder real-time: <summary>"` first, even if it isn't
this repo's usual domain, then sprint-plan it into `EMILY/BACKLOG.md` (`emily backlog curate`,
scoped into a real SECTION/sub-item, not just a one-line log), and only then implement. See
`EMILY/docs/THE_EMILY_WAY.md` Principle 18 ("Pave the Cow Paths").

## Apple Filing Protocol

After any meaningful change, file an Apple:
```bash
emily apples post -t completion -repo BURROW "<title>" "<body with commit hash>"
```
Then mark the item done in `EMILY/BACKLOG.md` and commit.

## CHANGELOG Protocol

After any meaningful change, update CHANGELOG.md:
```bash
emily changelog add BURROW "<what changed>"
# or manually: append a dated bullet under ## YYYY-MM-DD in BURROW/CHANGELOG.md
```

## Frame-Break Reframing

Founder-sourced prompting technique (REDGARDEN/NORTHSTAR.md §28, full origin in
REDGARDEN/docs2/MULTI_AGENT_RD_RESEARCH_NOTES.md §5): given a request, name the underlying
structural/systemic pattern it's one instance of — one level of abstraction up — as an added
lens during planning/triage/judgment calls. Use it to spot the general case behind a specific
ask. It augments judgment, it does not replace doing the work: direct, concrete execution of
the literal task asked for still happens every time.

## Commit Protocol (standing instruction)

Always commit and push completed work immediately — don't wait to be asked. This is the default for every repo in this monorepo.

Every commit — human-written or produced by automated code paths (git-commit helpers in emily-agent, emily.cli, IDUNA handlers, etc.) — must carry the active `emily session` fingerprint as a `session: <tag>` trailer (blank line, then the trailer). This was silently missing from several independently-implemented automated commit helpers across the monorepo until an audit on 2026-08-10 (founder, real-time: "where in the fuck is my llm session id anywhere"). If you add a new automated git-commit code path anywhere, wire in the session tag the same way — don't assume an existing helper already does it.
