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
nesting depth. **`defstruct`/`get-field` support added to the Go target the same day** (real
trigger: PARENA's own new `stdlib/k8s`/`stdlib/helm` packages needed it) — a registered
`defstruct` emits a real, exported Go struct type (nested struct fields included), `get-field`
lowers to plain Go dot access. Real, deliberate scope: construction (`{:field val}`) is NOT
emitted — every real function that would need it only ever *receives* a struct as a parameter; a
real Go host constructs one with an ordinary composite literal against the exported type, the
same split the C target's own real C test harnesses already use (`Deployment_new(...)`, not
in-PARENA construction). Real, honest boundary found the same day: `stdlib/k8s/k8s.prn`/
`stdlib/helm/helm.prn` themselves still only compile through the C target — their own
String+Arena-threaded string-building is fundamentally incompatible with this target's GC-off-safe
design (no allocation reachable); `stdlib/k8s/scaling.prn` (scalar-only) is the real proof the new
struct support actually works end to end for the Go target specifically, verified via a real
`go test` against `DUNG`'s own checked-in copy. **Unary `(not x)` added to BOTH emitters the same
day** (real trigger: PARENA's own new `stdlib/k8s/operator.prn`) — a real, genuine gap in both
`emit_c.go` and `emit_go.go` (fell through to a bogus call to a never-defined `not(...)` function;
the same real gap `parena-c`'s own `src/emit.c` already fixed on 2026-08-21, burrow just hadn't
hit a real file using it yet). `go test`: 52/52 (50 prior + 2 new). `burrow` is installed on
`PATH`. Full `emit.c` parity (`defenum`/`match`/`loop`/`Result`/`Vec`, struct *construction*,
`let`-bindings) and the TypeScript/Java targets remain unstarted.

**`defstruct`/`get-field` support added to the C target too, closing DUNG's own real found-live
gap** (2026-09-03, founder: "CONTINUE WORKING ON DUNG IDE" → investigated writing it in LO,
found LO itself not ready yet → "ok write it in parena and go? right? with burrow" — the real
next step was unblocking the exact gap `DUNG/parena/rect_probe.prn` found live on 2026-08-30:
`emit_c: unsupported top-level form`). Ported directly from `emit_go.go`'s own real, already-
shipped struct support (same real design: a registered `defstruct` emits a real C `typedef
struct {...} Name;`, `get-field` lowers to plain `(record).field` dot access, passed by value,
construction deliberately not emitted — same "receives, doesn't construct" split the Go target
already established). One real, C-specific difference from the Go port: struct typedefs are
collected into their own slice and emitted FIRST, before any function declaration/definition —
a real ordering constraint C has that Go doesn't. 3 new tests (2 mirrored from
`emit_go_test.go`'s own struct tests, adapted for C; a new ordering-specific test Go never
needed). Verified live, not just unit tests: a real `.prn` file with a `Rect`-shaped struct
compiled via `burrow build`, the emitted C compiled clean with `gcc -Wall -Wextra`, and run —
correct result (`(1440-20)/2 = 710`) against a real test harness. `go build`/`go vet`/`go test`
all clean (66/66 total).

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
