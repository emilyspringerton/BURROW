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

**Scoping only, this pass** — same real discipline every other NORTHSTAR doc in this monorepo
follows, doubly true here since this is a real, large undertaking (reimplementing a lexer, parser,
region analyzer, and four working emitters is a genuinely big project, not a same-afternoon
addition like `emit_ts.c`/`emit_java.c` were). No Go code exists yet.

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
new idea: the GC-off-safe discipline above was scoped only to code BURROW *emits* (Phase 5's own
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
GC-off-safe code-generation philosophy (Phase 5) is something its own authors actually believe in,
not just a claim made about code it hands to someone else. **Not yet a checked, enforced bar** (no
real Go code exists yet to check it against) — named here as a real, standing implementation
constraint for whoever writes Phase 1 onward, not deferred to a later cleanup pass.

## Real, phased plan (revised for the corrected, larger scope)

**Phase 0 — architecture confirmed** (leans on `selfhost/*.prn` + a new PARENA Go emission target,
real PARENA stdlib primitives reused directly where they fit, the stdlib grown with new packages
whenever a real BURROW need surfaces one — "use parena primatives adding to the stdlibs when it
helps"). No longer an open question as of the founder's own 2026-08-30 confirmation above.

**Phase 1 — lexer parity**: `selfhost/lexer.prn` already exists and is already tested against
`src/lexer.c`'s own real behavior (60 assertions, `tests/test_selfhost_lexer.c`) — the real,
smallest real slice of "does `burrow` produce the same tokens `parena-c` does" is already half
proven on the PARENA side; the real remaining work is getting it running as real Go (via the
hypothesis's own new Go emission target, or a hand-port if that hypothesis is rejected) and wired
into a real `burrow` CLI.

**Phase 2 — parser + region analyzer parity**: real, larger, currently-unstarted work — no
`selfhost/parser.prn`/`selfhost/region.prn` exist yet (only the lexer domain has been ported to
PARENA so far, per `PARENA/NORTHSTAR.md`'s own current, honest status).

**Phase 3 — emitter parity**: real C/TypeScript/Java output matching `emit.c`/`emit_ts.c`/
`emit_java.c`'s own real behavior, proven against the shared test corpus named above.

**Phase 4 — the real founder-named acceptance bar**: `burrow build`/`burrow parse`/`burrow
analyze` pass the exact same real `.prn` fixtures and assertions `parena-c`'s own
`tests/test_lexer_parser.c`/`test_region.c`/`test_emit.c`/`test_emit_ts.c`/`test_emit_java.c`
already check — this is the real, concrete "BURROW works" milestone, not a vaguer "feels
equivalent" one.

**Phase 5 (real, and the one piece of the ORIGINAL ask this doc's corrected scope still owns)** —
a real, new, native Go EMISSION TARGET (`parena build foo.prn -o foo.go`, or the BURROW-native
equivalent), designed GC-off-safe per the section above, proven against `GoblinFoxDragon`'s own
real Go backend as the named candidate host (not committed to).

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
- **No real host has asked for the Go emission target (Phase 5) yet** — same honest flag this
  doc's own first draft already carried for that specific piece; still true under the corrected,
  larger scope.

## Related

- `PARENA/NORTHSTAR.md`'s own "Self-hosting — real progress started" section — the real,
  independent, already-in-progress `selfhost/*.prn` effort this doc's own leading hypothesis leans
  on.
- `PARENA/src/emit_ts.c`, `PARENA/src/emit_java.c`, `PARENA/STDLIB.md` — the real, proven v0
  emitter template Phase 5's own Go emission target (if built) would still follow.
- `GoblinFoxDragon` — the real, named candidate host for a GC-off-safe PARENA-compiled decision
  layer (Phase 5, not committed to).
- `PAPERCRAFT/docs/NORTHSTAR_WEB_CLIENT.md` — the same-session precedent this doc's own "scoping
  pass before code" structure follows.
