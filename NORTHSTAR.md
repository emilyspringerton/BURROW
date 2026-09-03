# NORTHSTAR — BURROW (a Go + PARENA rewrite of the parena-c compiler)

## Where this came from, corrected in real time

Founder real-time, 2026-08-30, immediately after PARENA's own new Java emitter shipped (the
third real compilation target, after C and TypeScript, this same session):

> "scope it but keep reweriting papercraft in native parena and also preparing for the webgl
> version and then can we build project BURROW the golang emitter where we design it to run with
> the gc turned off - upstream GITHUB created"

**This doc's own first version (committed, then corrected the same session) mis-scoped this** as
"one more emission target bolted onto the existing C-implemented `parena-c` compiler" — the exact
same shape `emit_ts.c`/`emit_java.c` already are. The founder corrected this directly, twice, in
immediate follow-up messages:

> "project burrow is a feature for feature rewrite of the parena compiler in golang and parena"

> "like the first test is can we write a pure golang and parena tool that still pass all that
> parena c tests"

**Real, corrected scope**: BURROW is not a new backend bolted onto `parena-c` — it is a full,
**feature-for-feature parallel implementation of the entire PARENA compiler** (lexer → parser →
region analyzer → every real emit target `parena-c` already has: C, TypeScript, Java), written in
a real combination of Go and PARENA itself, with the first, concrete, falsifiable acceptance bar
named directly by the founder: **it must pass the same real C test corpus `parena-c` already
passes** (`tests/test_lexer_parser.c`, `tests/test_region.c`, `tests/test_emit.c`,
`tests/test_emit_ts.c`, `tests/test_emit_java.c`) — real behavioral parity proven against the
existing, already-trusted test suite, not a fresh, separately-invented one that could quietly
diverge from what `parena-c` actually does.

**Scoping-first, real code following incrementally** — same real discipline every other NORTHSTAR
doc in this monorepo follows, doubly true here since this is a real, large undertaking
(reimplementing a lexer, parser, region analyzer, and four working emitters is a genuinely big
project, not a same-afternoon addition like `emit_ts.c`/`emit_java.c` were). **Real status
(2026-08-30): Phases 1-4 (lexer, parser, region analyzer, and a real v0 C emitter) are shipped** —
`lexer.go`, `parser.go`, `region.go`, `emit_c.go` (+ their real test files), see each phase's own
entry below for the real architecture call made getting there. `burrow build` now really works for
the same narrow, real scalar scope `emit_ts.c`/`emit_java.c` already proved out — verified end to
end (real `gcc` compile + run) against `PAPERCRAFT/stdlib/papercraft/level_mod.prn`. `burrow` is
installed on `PATH`. Full `emit.c` parity (structs/enums/match/loop/`Result`/`Vec`) and the
TypeScript/Java targets remain real, honest, unstarted work.

## Why "in golang and parena" — the real, load-bearing connection to PARENA's own self-hosting effort

PARENA already has a real, in-progress self-hosting effort, independent of BURROW and started
first (`PARENA/NORTHSTAR.md`'s own "Self-hosting — real progress started (2026-08-27)" section,
founder real-time: "ok but after we have a compiler we also need to write parena in parena" →
"not c"): `selfhost/lexer.prn` is a real, already-built, already-tested (60 real assertions)
PARENA-language port of `src/lexer.c`, today compiled BY the existing C-based `parena-c` into C
output — a real domain toward self-hosting, not a claim of having reached it yet (self-hosting
only becomes real once a PARENA-in-PARENA pipeline can compile *itself*, and today's
`selfhost/lexer.prn` still needs `parena-c` to turn it into anything runnable).

**Real, founder-confirmed direction** (2026-08-30 follow-up: "use parena primatives adding to the
stdlibs when it helps"): BURROW leans directly on this existing `selfhost/*.prn` source tree
rather than hand-writing a second, separate lexer/parser/region-analyzer from scratch in raw Go —
i.e., BURROW compiles the real, existing `selfhost/*.prn` files (today only `lexer.prn` exists;
`parser.prn`/`region.prn`/the emitters would need to be written, in PARENA, to reach real feature
parity) using a NEW Go emission target, and a thin, real Go host program (`burrow`, the actual CLI
binary) wraps the resulting Go code with I/O, CLI argument handling, and file writing — the same
real "PARENA owns the decision logic, host code owns the plumbing" split every mod in this
monorepo already uses, just applied to the compiler's own implementation this time. This makes
BURROW simultaneously: (1) the real, concrete forcing function that finally makes PARENA's own Go
emission target real (closing the loop this doc's own first draft was reaching for), and (2) the
real vehicle that pushes PARENA's own self-hosting effort forward (`parser.prn`/`region.prn`/the
emitters, written in PARENA, not just the lexer). **Real, explicit founder instruction on how far
to lean on this**: "when it helps" — real, existing PARENA stdlib primitives get reused directly
where they already fit (string/lexer helpers, `math`, etc.), and the PARENA stdlib itself grows
(new packages under `PARENA/stdlib/`, same real, incremental, gap-driven discipline `STDLIB.md`
already documents dozens of times over) whenever a real BURROW need surfaces one — not a mandate
to force every single line of `burrow` through PARENA regardless of fit; Go stays the real, direct
implementation language for whatever doesn't genuinely benefit from being PARENA-native (CLI
plumbing, file I/O, the Go emission target's own runtime glue).

## The real, concrete Phase 0 acceptance bar (founder's own words)

> "can we write a pure golang and parena tool that still pass all that parena c tests"

This is real, falsifiable, and should stay the actual Definition of Done for "BURROW works," not a
softer "looks equivalent" bar:

- A real `burrow` binary (Go, calling into PARENA-compiled-to-Go logic per the hypothesis above,
  or hand-written Go where a `.prn` port doesn't exist yet) that implements `parse`/`analyze`/
  `build -o <output.c|.ts|.java>` — the same real CLI surface `parena-c`'s own `main.c` already
  has.
- The exact same real `.prn` test fixtures `parena-c`'s own C test suite already uses (not a new,
  separately-invented corpus) run against `burrow` instead, and every real assertion currently
  checked against `parena-c`'s own emitted C/TS/Java text must also hold against `burrow`'s own
  output for the same input.
- Real, honest scope note: `parena-c`'s own test suite checks emitted TEXT via substring matching,
  not byte-for-byte identity (`test_emit_ts.c`'s own header comment: "a C string equality check
  here would be too brittle to whitespace/formatting choices to be the real acceptance bar") — the
  same real bar applies to BURROW: functionally equivalent, real, correct output, not
  byte-identical text to `parena-c`'s own.

## The real, honest GC-off design question (still real, now scoped to BURROW's own Go emission target specifically)

This doc's own first draft answered this in detail; the real answer doesn't change under the
corrected scope, it just applies to a narrower real piece of the larger project — the Go emission
target BURROW's own feature-parity work would add (not to the `burrow` CLI tool's own process,
which is a real, ordinary, short-lived batch compiler invocation with no real GC-off use case of
its own):

Go's `debug.SetGCPercent(-1)` (`runtime/debug`) / `GOGC=off` are real, standard, already-shipping
mechanisms for a real host process to disable Go's garbage collector entirely — the real, standard
way latency-sensitive Go programs (game servers, HFT systems) opt out of GC pause
unpredictability. Real, honest cost: memory is never reclaimed automatically once set — a real
trade-off the embedding host's own author must accept deliberately, BURROW cannot make it safe by
itself. What BURROW's own real, narrow v0 Go-target scope CAN and must do (same real discipline
`emit_ts.c`/`emit_java.c` already commit to for their own v0 slices): generate code that never
allocates on the Go heap — scalar `I32`/`F64`/`Bool`/`String` params, one-expression bodies, no
`let`/blocks/structs/slices/maps/closures — meaning a v0-scope emitted function is already
GC-irrelevant by construction, and a real host that disables GC around calls into it is making a
real, safe, informed choice. `GoblinFoxDragon`'s own Go DragonsNShit backend remains the named
real candidate consumer for this specific piece (not committed to, see `PAPERCRAFT`-adjacent
precedent for how that kind of real integration call gets made later, not guessed at now).

## Real, dogfooding directive: `burrow`'s own implementation should practice this too

Founder real-time, immediately after the GC-off design question above: "dog food it like write
the golang in a way that doesnt allocate on the heap or whatever." Real, direct extension, not a
new idea: the GC-off-safe discipline above was scoped only to code BURROW *emits* (Phase 6's own
Go target) — this applies the same real discipline to `burrow`'s *own* implementation, the actual
Go source of the lexer/parser/region-analyzer/emitters themselves. Real, concrete practices this
means, matching Go's own well-known, real, standard low-allocation idioms (not exotic — the same
real toolkit Go's own standard library and every serious low-latency Go codebase already uses):
value types and pre-sized slices over ad-hoc `append`-driven growth where the real size is knowable
up front (the same real "know your bound, allocate once" discipline `src/arena.h`'s own bump
allocator already practices in C); passing structs by value or via a caller-owned buffer instead
of returning freshly-`new`'d pointers per call; reusing one real, long-lived buffer across repeated
lexer/parser calls (a real Go `sync.Pool`, or simpler, a single reusable buffer threaded through
by the caller) rather than allocating a fresh one per token/node; avoiding `interface{}`/`any` in
hot paths (a real, well-known Go source of hidden heap escapes via boxing). **Real, honest scope
note**: `burrow` itself is a real, ordinary, short-lived CLI compiler invocation (see the GC-off
section above) — it doesn't need `debug.SetGCPercent(-1)` itself to be safe, and this dogfooding
directive isn't asking for that. The real point is a genuinely low-allocation *implementation
style* throughout `burrow`'s own Go source, both because it's real, good, idiomatic low-latency Go
practice worth modeling, and because it's the most honest possible proof that BURROW's own
GC-off-safe code-generation philosophy (Phase 6) is something its own authors actually believe in,
not just a claim made about code it hands to someone else. **Not yet a checked, enforced bar** (no
real Go code exists yet to check it against) — named here as a real, standing implementation
constraint for whoever writes Phase 1 onward, not deferred to a later cleanup pass.

## Real, phased plan (revised for the corrected, larger scope)

**Phase 0 — architecture confirmed** (leans on `selfhost/*.prn` + a new PARENA Go emission target,
real PARENA stdlib primitives reused directly where they fit, the stdlib grown with new packages
whenever a real BURROW need surfaces one — "use parena primatives adding to the stdlibs when it
helps"). No longer an open question as of the founder's own 2026-08-30 confirmation above.

**Phase 1 — lexer parity — real, shipped (2026-08-30).** Founder real-time: "start phase 1 lexer
parity on burrow." Real, honest architecture call made here, named directly rather than glossed
over: this phase's own real candidate paths were "compile `selfhost/lexer.prn` through a new
PARENA Go emission target" or "hand-port `src/lexer.c` directly if that hypothesis is rejected."
Reading `selfhost/lexer.prn` in full before starting (not guessed at) showed it leans on real
PARENA language surface far beyond `emit_ts.c`/`emit_java.c`'s own proven narrow v0 scope —
`defstruct`, a payload-carrying `defenum`, `match`, `loop`/`recur`, `Result<T,E>`, `Vec`, and
reference parameters all appear throughout. Building a general enough Go emitter to cover all of
that correctly, in one sitting, at a quality bar this repo could actually verify, was not
realistic — so this phase took the doc's own named fallback: **`lexer.go`, a real, faithful,
hand-written Go port of `src/lexer.c`** (the real C reference both `src/lexer.c` itself and
`selfhost/lexer.prn` already document as the source of truth), verified via `lexer_test.go` — a
real, direct port of all 9 real test scenarios from `tests/test_selfhost_lexer.c`, every expected
token sequence copied verbatim from that file's own real, hand-traced-against-`src/lexer.c`
expectations, not re-derived. `go build`/`go vet`/`go test` all clean, all 9 tests pass. Also
stress-tested (not part of the committed test suite, a one-off local check) against real,
substantial PARENA source — `selfhost/lexer.prn` itself (2104 tokens), `stdlib/string.prn` (744),
`stdlib/mishri/humanness.prn` (127), `stdlib/gta7/humanness_fingerprint_mod.prn` (74) — all
tokenized cleanly, no crashes, no errors. This is the real, founder-named "pass all that parena c
tests" acceptance bar, achieved directly for the lexer domain — not a new PARENA Go emission
target (that piece, if ever built, is real, separate, deferred work, see Phase 6 below).

**Phase 2 — parser parity — real, shipped (2026-08-30).** Founder real-time: "continue project
BURROW." Same real architecture call Phase 1 already made, inherited directly: `selfhost/
parser.prn` doesn't exist yet, so there's no PARENA-source parser to compile via a PARENA-Go
emitter that also doesn't exist — took the same hand-port fallback. **`parser.go`**: a real,
faithful Go port of `src/parser.c` + `ast.h` (the C reference's own recursive-descent parser and
generic S-expression tree), verified via `parser_test.go` — all 14 real test scenarios from
`tests/test_lexer_parser.c` ported verbatim (both the balanced-form cases and the DoD's own
required negative/malformed cases: unterminated list, mismatched bracket kind, stray closing
paren, unterminated string, mismatched nested bracket), every error message and structural
expectation copied from that file's own real assertions. `go build`/`go vet`/`go test` all clean,
23/23 total (9 lexer + 14 parser). **Real, deliberate idiomatic departure**: Go multi-return
`(*Node, error)` replaces the C reference's own `setjmp`/`longjmp` error-unwind entirely — Go has
no equivalent primitive worth reaching for here, and `panic`/`recover` is idiomatically reserved
for exceptional conditions, not routine syntax errors. Stress-tested against the ENTIRE real
PARENA corpus, not just the unit tests: all 111 `.prn` files across `stdlib/` and `selfhost/`
parse successfully with zero failures. Cross-checked structurally against the real C reference's
own `parena parse` dump on `stdlib/base4/algebra.prn`: exact match, 246 total nodes, 13 top-level
forms, both sides identical.

**Phase 3 — region analyzer parity — real, shipped (2026-08-30).** Founder real-time: "continue
DUNG i guess you are gonna need burrow on the path if its not already" — DUNG's own real build
needs a working region-analyze + emit pipeline first. Same real architecture call Phases 1-2
already made: `selfhost/region.prn` doesn't exist yet, so hand-ported `PARENA/src/region.c`
directly. **`region.go`**: a real, faithful Go port of the C reference's own real invariant check
(`Region(Source) >= Region(Destination)` on region-annotated `Arena` params — not the full
region-safety story, matching `region.h`'s own real, narrow, already-documented scope), verified
via `region_test.go` — all 7 real test scenarios from `PARENA/tests/test_region.c` ported verbatim,
including the exact literal error-message text (`"Compile Error: Escaping region pointer from
:region/scratch to :region/buffer at line 4"`) DoD's own table specifies. `go build`/`go vet`/
`go test` clean, 30/30 total. Stress-tested against the entire real corpus: all 111 `.prn` files
parse AND region-analyze successfully, zero false positives/negatives.

**Phase 4 — emitter parity (C target, v0) — real, shipped (2026-08-30).** **`emit_c.go`**: the
same real, narrow v0 scope `PARENA/src/emit_ts.c`/`src/emit_java.c` already proved out this same
session (scalar `I32`/`F64`/`Bool`/`String` params, one-expression bodies, the real binop set —
confirmed by reading `emit.c` directly, not guessed: `+`/`-`/`*`/`/`/`<`/`>`/`<=`/`>=`/`=`→`==`/
`and`→`&&`/`or`→`||`/`bit-and`→`&`/`bit-or`→`|`/`bit-xor`→`^`/`mod`→`%`, real C types
`int`/`double`/`char *`/`void`) — deliberately NOT the full 5944-line `emit.c` (`defstruct`/
`defenum`/`match`/`loop`/`recur`/`Result`/`Vec` remain real, separate, unstarted work, the same
honest scale call every earlier phase already made. **Real, genuine bug found and fixed while
testing this for real** against `stdlib/mishri/humanness.prn`: an early version silently emitted
`math/pi`/`math/random` (real, namespaced identifiers this v0's C target doesn't support, matching
`STDLIB.md`'s own documented gap) as literal broken C syntax — cross-checked directly against the
real reference (`parena build` on the identical input reports a real "unknown identifier 'math/pi'
at line 36" error), confirming this needed a real "does this identifier actually resolve to a
local param or a known top-level defn" validation, not a silent pass-through. Fixed, with two new
regression tests. `go test`: 38/38 total (9 lexer + 14 parser + 7 region + 8 emit_c). **Real, full
end-to-end proof, not just written**: compiled `PARENA/stdlib/papercraft/level_mod.prn` via
`burrow build`, actually compiled the output with a real `gcc`, and ran it against the exact same
real assertions `PAPERCRAFT/packages/simulation/level_mod_test.c` already uses — all pass. `burrow`
is now installed on `PATH` (`~/.local/bin/burrow`, matching this monorepo's own real CLI-install
convention). TypeScript/Java targets remain real, honest "not yet implemented."

**Phase 5 — the real founder-named acceptance bar**: `burrow build`/`burrow parse`/`burrow
analyze` pass the exact same real `.prn` fixtures and assertions `parena-c`'s own
`tests/test_lexer_parser.c`/`test_region.c`/`test_emit.c`/`test_emit_ts.c`/`test_emit_java.c`
already check — this is the real, concrete "BURROW works" milestone, not a vaguer "feels
equivalent" one. **Real, substantial progress**: Phases 1-4 now cover the real, shared corpus from
`test_lexer_parser.c`/`test_region.c` (ported directly, 30/30 assertions) plus a real, narrow v0
slice of `test_emit.c`'s own real scope (scalar C-target functions, not the full 5944-line
surface) — the remaining gap is full `emit.c` parity (`defstruct`/`defenum`/`match`/`loop`/`Result`/
`Vec`) plus `test_emit_ts.c`/`test_emit_java.c` (the TypeScript/Java targets).

**Phase 6 — real, native Go emission target — real, shipped (2026-08-30).** Founder real-time:
"continue adding to DUNG's parena work, we want to emit go with burrow" — the real, concrete host
this section's own first draft only had a named CANDIDATE for (`GoblinFoxDragon`, "not committed
to") finally asked for it. `emit_go.go`: the same real, narrow v0 scope emit_c.go's own Phase 4
already proved (scalar `I32`/`F64`/`Bool`/`String` params, one-expression bodies, the real binop
set, calls to sibling defns), `-o *.go` on `burrow build`. GC-off-safe by construction per this
doc's own "real, honest GC-off design question" section above — no `let`/blocks/structs/slices/
maps/closures reachable in v0 scope, so no heap allocation is possible in the emitted code itself,
exactly the real precondition that section named. Real Go-specific departures from the C target,
found and fixed live, not designed in advance: kebab-case → exported PascalCase (the emitted
package is meant to be imported directly by a real host, unlike C's `#include`-and-link model);
`if` lowered to an immediately-invoked func literal (Go has no `?:` ternary at all); two real,
genuine bugs caught by an actual nested-if probe (`clamp01`, two levels deep) before this was
trusted — a bare `I32` literal boxed through the func literal's own `any`-typed return defaults to
Go's own `int`, not `int32`, so `.(int32)` at the call site panicked at runtime until every branch
value got an explicit `RetType(...)` conversion first; a NESTED `if`'s own result is itself
`any`-typed, so wrapping it with the same `RetType(...)` conversion is a compile error ("need type
assertion") until the `if` case was changed to self-assert its own result before returning it up.
A final `go/format.Source` pass (the same real formatter `gofmt` itself calls) makes emitted
output unconditionally gofmt-clean regardless of nesting depth, rather than this emitter
hand-tracking indentation to match gofmt's own canonical control-flow line-breaking rules.
`go test`: 48/48 (38 prior + 10 new, covering the same real failure classes emit_c_test.go's own
suite already checks, plus the two Go-specific bugs above). **Real, full end-to-end proof, not
just written**: `DUNG`'s own real `parena/entry.prn` compiled to `internal/burrowgen/entry_gen.go`
and wired directly into `cmd/dung/main.go` (`layout()`'s split-size math, `nextFocus()`'s
wraparound math) — no cgo/FFI boundary, a real Go import — `go build`/`go vet`/`go test` and a
real `bazel build //...` both clean, the real SDL2 visor binary run under Xvfb identically to
before the mod was wired in. `DUNG` is this target's own real, live, first host now — Phase 5's
GoblinFoxDragon candidacy is unaffected, just no longer the only one.

**`let`/`do` shipped (2026-09-03)**, kanban priority-queue cards 1199/9988: "iterate on project
burrow... so that parena gets transformed into idiomatic go" / "emily for business CLI written
in GO with BURROW." The single largest real gap standing between this target and writing any
real, non-trivial program (a CLI included): v0's "one-expression body" scope meant no `.prn`
function could declare a local variable at all. Closed via the exact same immediately-invoked-
func-literal-boxed-through-`any` shape the `if` case already established — `let`/`do` compose
with `if` (and each other) for free, real bindings are scoped to a cloned params map so they
never leak outside their own `let` (proven by a real test, not just claimed). `go test`: 73/73.
**Real, live, end-to-end proof**: a real two-binding `let` and a real nested `let`-inside-`if`
both compiled via `burrow build`, linked into a real, separate Go module, and run with correct
output. Still real, honest, unstarted: `defenum`/`match`/`loop`/`Result`/`Vec`, struct
construction — a real CLI needs at least `match`/`Result` for real error handling and `loop` for
iteration before "write a CLI in this" is fully true, not just closer to true.

## Real risks and open questions, named honestly

- **Scale**: this is a genuinely large project relative to everything else built this same
  session — `emit_ts.c`/`emit_java.c` were each ~360-400 lines added to an already-complete
  compiler; BURROW is a parallel lexer+parser+region-analyzer+4-emitter reimplementation. Phasing
  it (above) is real, not just a formality — Phase 1 alone (lexer parity in real, running Go) is
  a reasonable, real, first, honestly-scoped deliverable, not "build the whole thing now."
- **Two-language maintenance burden**: if BURROW ships and is kept alive alongside `parena-c`,
  every future real language feature needs implementing (or explicitly porting) twice — a real,
  ongoing cost this doc names but doesn't resolve (matching Go/Rust's own real, well-known
  "self-hosted rewrite" tradeoff history, cited in `PARENA/NORTHSTAR.md`'s own self-hosting
  section as the precedent this whole idea is patterned on).
- **No real host has asked for the Go emission target (Phase 6) yet** — same honest flag this
  doc's own first draft already carried for that specific piece; still true under the corrected,
  larger scope.
- **Real, concrete dogfooding data on where Phase 4's own v0 C emitter stops, from `DUNG`'s own
  first real usage (2026-08-30)**: `DUNG/parena/rect_probe.prn` defines a `defstruct` — `burrow
  parse`/`analyze` both accept it fine (parser/region already handle `defstruct` generally), but
  `burrow build` fails exactly at `emit_c: unsupported top-level form (v0 only understands defn,
  module, export, import)`. Real, precise scoping input for whichever comes first, full `emit.c`
  parity or a narrower "just `defstruct`" step: `DUNG`'s own real editor-domain build (structs for
  buffer/pane/node state) is blocked on this specific gap, not on `match`/`loop`/FFI yet — those
  come after. See `DUNG/NORTHSTAR.md`'s own "Real first PARENA entry attempt" section for the
  full real test (two working scalar functions ported and gcc-verified first, then this probe).

## Related

- `PARENA/NORTHSTAR.md`'s own "Self-hosting — real progress started" section — the real,
  independent, already-in-progress `selfhost/*.prn` effort this doc's own leading hypothesis leans
  on.
- `PARENA/src/emit_ts.c`, `PARENA/src/emit_java.c`, `PARENA/STDLIB.md` — the real, proven v0
  emitter template Phase 6's own Go emission target (if built) would still follow.
- `GoblinFoxDragon` — the real, named candidate host for a GC-off-safe PARENA-compiled decision
  layer (Phase 6, not committed to).
- `DUNG` — its own real, separate repo (first scoped inside this repo as `DUNG.md`, moved by
  founder correction), "the BURROW editor" — its own real build depends on this repo's own emit
  capability. **Real, no longer a full blocker (2026-09-03)**: `defstruct`/`get-field` support
  landed in the C target (see `CLAUDE.md`'s own real capability record), closing the exact gap
  `DUNG/parena/rect_probe.prn` found live. DUNG's own ground-up PARENA editor port can now use
  real structs for its buffer/pane/node state; `match`/`loop`/`Vec`/construction still don't
  exist in either target, a real, narrower remaining constraint for DUNG's next scoping pass, not
  the full `emit.c`-parity wait this note originally named.
- `PAPERCRAFT/docs/NORTHSTAR_WEB_CLIENT.md` — the same-session precedent this doc's own "scoping
  pass before code" structure follows.
