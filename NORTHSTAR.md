# NORTHSTAR — BURROW (PARENA's Go emitter, designed for a GC-off host)

## Where this came from

Founder real-time, 2026-08-30, immediately after PARENA's own new Java emitter shipped (the
third real compilation target, after C and TypeScript, this same session):

> "scope it but keep reweriting papercraft in native parena and also preparing for the webgl
> version and then can we build project BURROW the golang emitter where we design it to run with
> the gc turned off - upstream GITHUB created"

The founder had already created the real, empty upstream repo before asking
(`git@github.com:emilyspringerton/BURROW.git`, confirmed via `git ls-remote`, one commit —
GitHub's own stock LICENSE-only initial commit) — this doc is the scoping pass that follows it,
same real discipline `PAPERCRAFT/docs/NORTHSTAR_WEB_CLIENT.md` and every other NORTHSTAR doc in
this monorepo already follows: **the real architecture gets decided here, before the emitter gets
written**, not built first and explained after.

**Scoping only, this pass.** No `emit_go.c` exists yet. `NORTHSTAR.md`'s own "smallest real proof
point" convention applies here too — this doc names the real shape, the real, honest GC-off
design question gets answered directly (not glossed over), and the concrete next step is named at
the bottom, same as `NORTHSTAR_WEB_CLIENT.md`'s own Phase 0.

## What this is

**BURROW is PARENA's fourth real compilation target: Go.** Not a new language, not a new compiler
— a new emitter (`emit_go.c`, or possibly `emit_go.go` if self-hosted later, undecided, see below)
living inside the real, existing `PARENA` repo, following the exact same real, narrow v0 template
`emit_ts.c` and `emit_java.c` already proved out this same session — C first
(`emit.c`, VS0's own original target), then TypeScript (`emit_ts.c`), then Java (`emit_java.c`),
now Go, directly continuing `PARENA/CLAUDE.md`'s own pre-declared "C first, then JVM/TypeScript/
WebAssembly" roadmap (Go added to that list by this ask).

**Why a separate top-level repo (BURROW) rather than a fourth file inside PARENA itself** — a real,
deliberate call, not an oversight: `emit_ts.c`/`emit_java.c` both live inside `PARENA/src/`
because they're small, self-contained C files sharing PARENA's own existing Bazel/Makefile/CI
setup with near-zero marginal cost. A Go emitter's own real GC-off design work (see below) is a
genuinely bigger, more open-ended investigation — real Go runtime tuning, real benchmarking against
a real GC-on baseline, possibly a real, separate Go-side runtime support package (`burrow/runtime`)
consumed by the generated code — closer in shape to `SAND`/`JEWEL`/`EXODUS`'s own real "big enough
to want its own repo, its own README, its own issue tracker" precedent than to a same-file addition.
**Revisit this call if BURROW's own real scope turns out to be as narrow as `emit_ts.c`/
`emit_java.c`'s own ~360-400 lines each** — genuinely possible, not committed to either way yet.

## The real, honest GC-off design question, answered directly

Go is a garbage-collected language — there is no "Go without a GC" compiler flag, and BURROW isn't
attempting to build one (that would mean reimplementing Go's own memory model, a real, separate,
enormous undertaking, not a compilation-target scoping problem). What Go **does** offer, real and
already shipping in the standard library:

- **`debug.SetGCPercent(-1)`** (`runtime/debug`) — disables the garbage collector's automatic
  collection cycle entirely for the calling process. The real, standard way real, latency-sensitive
  Go programs (game servers, HFT systems) opt out of GC pause unpredictability. Real, honest cost:
  memory is never reclaimed automatically once this is set — a real, live process that disables GC
  and keeps allocating will eventually exhaust memory. Not a free lunch, a real trade-off the host
  program's own author must accept deliberately.
- **`GOGC=off`** (env var) — the same real effect, set at process-launch time instead of in code.

**Real, honest reframe of what "design it to run with the gc turned off" actually means for an
emitter, not a language**: BURROW's own generated code cannot make disabling the GC *safe* by
itself — that's a property of the whole process, decided by whatever real Go host program embeds
BURROW's output (a real candidate: `GoblinFoxDragon`'s own Go DragonsNShit backend, a real,
already-running, latency-sensitive game server in this monorepo — the natural real consumer for a
GC-off-safe PARENA-compiled decision layer, matching PAPERCRAFT's own C mods / MISHRI's own
TypeScript integration precedent). What BURROW's own emitter design **can** and must do: **generate
code that never allocates on the Go heap** — the same real, narrow v0 scope `emit_ts.c`/
`emit_java.c` already commit to (scalar `I32`/`F64`/`Bool`/`String` params, one-expression bodies,
no `let`/blocks/structs/slices/maps/closures) means every real function this v0 can emit already
operates purely on stack-resident scalar values with zero heap allocation — a real, happy structural
fact, not a coincidence this doc invents: **a v0-scope BURROW function is already GC-irrelevant by
construction**, meaning a host that disables GC around calls into BURROW-generated code is making a
real, informed, safe trade — the generated code itself never gives the collector anything to do.
This is the real, concrete, honest answer to "design it to run with the gc turned off": not a
special code-generation mode, but a scope discipline (no heap allocation in emitted code) that
makes GC-off a safe *choice for the host*, not a promise BURROW's own compiler enforces or verifies
across the host's other code.

**Named, not solved, by this pass**: verifying a real host's *own* code (not just BURROW's
generated slice of it) is also allocation-free enough to safely run `GOGC=off` indefinitely is a
real, separate, host-side concern — BURROW has no way to audit that, and this doc doesn't claim it
does.

## Real architecture (mirrors `emit_ts.c`/`emit_java.c`, one real Go-specific difference named)

Same real, narrow v0 scope as the TypeScript and Java emitters (not re-derived from scratch — see
`PARENA/STDLIB.md`'s own "Real, third proof: the new Java emitter" section for the full real
template this follows):

- `defn` with zero or more scalar (`I32`/`F64`/`Bool`/`String`) parameters, no `Arena`/region
  annotations (Go is garbage-collected — same real "no-op for this target" reasoning `emit_ts.h`
  already gives for TypeScript).
- A body that is exactly one real expression: number/symbol literals, the same real binop table
  (`+`/`-`/`*`/`/`/`=`→`==`/`<`/`>`/`<=`/`>=`/`and`→`&&`/`or`→`||` — Go's own `==` matches Java's
  real choice here, not TypeScript's `===`, for the identical real reason: no triple-equals token
  in the language), `if` — **real, Go-specific difference, the one genuinely new structural piece**:
  Go has no ternary expression operator at all (unlike TypeScript's `?:` and Java's `?:`, both
  already proven this session) — `if` must lower to a real Go **function literal**,
  `func() T { if cond { return then } ; return else }()`, called immediately, to keep the
  single-expression-body v0 contract intact without inventing statement-level codegen. A real,
  slightly uglier but completely valid piece of generated Go — flagged honestly here so it isn't
  mistaken for an oversight when `emit_go.c` actually gets written.
- The same real `math/*` primitive table, ported to Go's own real standard library names — one
  genuine naming/import wrinkle, unlike TS/Java where `Math.*` covered every entry from one
  namespace: Go's `math` package has `math.Floor`/`math.Sqrt`/`math.Log`/`math.Cos`/`math.Pi`, but
  **no random-number function** — `math/random` must lower to `math/rand`'s own `rand.Float64()`,
  a **second, separate real import** the emitter must conditionally emit only when a generated file
  actually uses `math/random` (`emit_ts.c`/`emit_java.c` never needed conditional imports at all,
  since `Math.*` was always available with no import statement in either target — a real, new,
  slightly bigger emitter responsibility BURROW's own v0 must take on).
- Go **does** have real top-level free functions (unlike Java) — no class-wrapper needed, closer
  to TypeScript's own shape. Every top-level `defn` becomes one real, exported (capitalized, Go's
  own real public/private-by-case convention — a `defn` name is always exported, matching every
  other target's "everything is public" v0 default) top-level `func`, inside a real
  `package <name>` header the caller supplies (mirroring `class_name` in `emit_java`'s own real
  design, but naming a package instead of wrapping a class).

## Real, phased plan

**Phase 0 — the emitter itself**, matching `emit_ts.c`/`emit_java.c`'s own real build order:
1. `emit_go.h`/`emit_go.c` inside `PARENA/src/` (see "why a separate repo" above for why the
   *repo* is BURROW but the emitter source may still land in PARENA — real, still-open placement
   question, not contradicted by BURROW's own existence: BURROW could end up being the real Go
   *runtime support package* + benchmarking home, while `PARENA/src/emit_go.c` stays where every
   other emitter already lives. **Not decided by this pass** — flagged, not resolved.)
2. `tests/test_emit_go.c`, same real substring-check discipline, same real failure-case coverage
   (wrong-arity math primitive, `Arena`-typed parameter, `let`-block body).
3. Wire into `main.c`'s existing extension dispatch (`.go` routes to `emit_go`).
4. Real proof: compile the exact same, unmodified `stdlib/mishri/bezier_interp.prn` and
   `humanness.prn` sources already proven for C/TypeScript/Java to real `.go`, and actually run
   `go build`/`go vet` against the output — matching the "verified with a real toolchain, not just
   written" bar every prior target met.

**Phase 1 — the real GC-off proof** (the part that makes BURROW more than "one more emitter"):
a small, real benchmark program that calls BURROW-generated functions in a tight loop with
`debug.SetGCPercent(-1)` set, confirming real, stable memory (no growth) over a real, sustained
run — the concrete, measured evidence for this doc's own "v0-scope BURROW functions are already
GC-irrelevant by construction" claim above, not left as an assertion.

**Phase 2 — a real host integration** (design only, not detailed here): `GoblinFoxDragon`'s own Go
backend is the named real candidate consumer, not committed to yet — genuinely a separate, later
decision for whoever owns that repo's own roadmap at the time.

## Real risks and open questions, named honestly

- **Emitter source location** (inside `PARENA/src/` vs. inside `BURROW` itself) — not decided,
  flagged above.
- **`if`-as-immediately-invoked-closure** codegen — valid Go, but real, unaudited question of
  whether Go's own compiler reliably inlines/optimizes this away in practice, or whether it's a
  real, measurable overhead in a GC-off, latency-sensitive host — a real Phase 1 benchmark
  question, not assumed either way here.
- **Conditional import emission** (`math/rand` only when actually used) — the one genuinely new
  emitter-complexity piece none of the prior three targets needed; real, not yet designed at the
  line-of-code level.
- **No real host has asked for this yet** — unlike MISHRI (the real, concrete TypeScript proving
  ground) this is speculative infrastructure investment ahead of a named consumer, a real, honest
  difference from how the TS/Java targets got built this session. Not a reason not to build it
  (the founder's own explicit ask), just named so the real "why now" isn't misremembered later.

## Related

- `PARENA/src/emit_ts.c`, `PARENA/src/emit_java.c`, `PARENA/STDLIB.md` — the real, proven v0
  emitter template this doc follows target-for-target.
- `PARENA/CLAUDE.md` — the original "C first, then JVM/TypeScript/WebAssembly" roadmap statement;
  Go is a real, new addition to that list, made by this doc.
- `GoblinFoxDragon` — the real, named candidate host for a GC-off-safe PARENA-compiled decision
  layer (Phase 2, not committed to).
- `PAPERCRAFT/docs/NORTHSTAR_WEB_CLIENT.md` — the same-session precedent this doc's own "scoping
  pass before code" structure follows.
