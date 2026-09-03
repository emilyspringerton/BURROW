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
`PATH`. Full `emit.c` parity (`defenum`/`loop`/`Vec`, struct *construction*) and the
TypeScript/Java targets remain unstarted — `let`/`do` and, for the Go target specifically,
`match`/`Result`/`Option` (with a real, deliberate v0 boundary — see below) both landed
2026-09-03, see below for each.

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

**`let`/`do` added to the Go emission target** (2026-09-03, kanban priority-queue cards
1199/9988: "iterate on project burrow... so that parena gets transformed into idiomatic go" /
"emily for business CLI written in GO with BURROW"). Real, decisive finding that motivated
this: v0's own "one-expression body" scope meant no real `.prn` function could declare a local
variable at all — the single largest real gap blocking any real multi-statement logic (a CLI's
own arg parsing, string building, sequential setup) from reaching this target. Real design:
`let`/`do` emit as the exact same immediately-invoked-func-literal-boxed-through-`any` shape the
`if` case already established, so they compose with `if` (and each other) for free — a `let`
nested inside an `if` branch, or vice versa, always hands back a concrete, correctly-typed Go
expression to whatever wraps it. Real, deliberate scope: `let` bindings are evaluated into a
CLONED local-params map, not `scope`'s own — a binding never leaks outside its own `let`, proven
by a real test asserting the exact "unknown identifier" error a leak would silently avoid. `go
test`: 73/73 (66 prior + 7 new — single/multi-binding let, sequential-scope chaining, the
no-leak proof, nested-let-inside-if composition, `do`'s own effect-then-result sequencing).
**Real, live, end-to-end proof, not just unit tests**: a real two-binding `let` (`(let [y (* x 2)
z (+ y 1)] z)`) and a real nested `let`-inside-`if` (`clamp-and-double`) both compiled via a real
`burrow build`, `go build`-linked into a real, separate Go module, and run — correct output
(`double-it(5) = 11`, `clamp-and-double(5) = 10`, `clamp-and-double(-3) = 0`) for both.

**`match`/`Result`/`Option` added to the Go emission target** (2026-09-03, kanban card 9988's
own explicit follow-up ask: "take on the full match/Result BURROW port"). Real, direct port of
PARENA's own reference C runtime's representation (`parena_runtime.h`'s own
`{int tag; void *value;}` Result/Option) — one real, FIXED, shared Go struct per Result/Option
(`type Result struct { Tag int; Value any }`, same for `Option`), not per-instantiation types,
since VS0 has no generics to give either a real one anyway (same real reason `bstree.prn`/
`json.prn` commit to concrete types instead of a generic container). `Ok`/`Err`/`Some` construct
via a plain Go composite literal; bare `None` (the real, established PARENA source convention —
see `bstree.prn`'s own live `get-loop`) constructs an empty `Option`. Go's own `any` interface
does the same real "erase to a pointer-ish box" job C's own `void *` does, for free, without this
emitter needing arena-boxing-helper machinery the C target's own equivalent needed.

**Real, deliberate v0 boundary for `match`, named explicitly, not silently limited**: the
scrutinee must be a direct call to a known defn whose own declared return type is Result/Option
(payload/error types resolved via a new `defnRetInfo` map, built in `EmitGo`'s own first pass,
the same real timing `knownDefns`/`knownStructs` already use) — NOT an arbitrary expression, and
NOT a `let`-bound variable. Real reason: this emitter's own local-variable tracking
(`localParams map[string]bool`) carries presence only, no per-variable type; extending that to
track real types for every local is real, separate, larger work — the same real scope PARENA's
own mature `src/emit.c` needed several distinct, later bug-fix passes (dated 2026-08-21,
2026-08-23, 2026-08-24, 2026-08-27 in that file's own accumulated commentary) to get fully right
for its C target, not rushed here. A scrutinee call's own return type IS staticly knowable up
front, though — exactly the real, useful, common case this v0 covers: "call a function that
might fail, then match its result immediately." Real, deliberate exhaustiveness rule: exactly 2
clauses required, each naming a distinct real tag (Ok/Err or Some/None) — the second clause
compiles to a plain Go `else`, not a second `else if`, so the Go compiler sees a real, complete
if/else needing no dead trailing panic to satisfy it; two clauses naming the same tag is a real,
honest compile error, not silently dead code.

**Two real, additional gaps found and fixed live while building this, not designed in
advance**: (1) string literals had NO handling anywhere in `emitGoExpr` at all, even though
`String` was already a supported param/return type — the first thing in this whole target's
history to actually need a plain string constant (a `Result I32 String` error message) surfaced
it; fixed via `strconv.Quote` on the lexer's own already-escape-decoded text, correctly
re-escaping a literal containing a real `"` or `\`, not a naive passthrough. (2) a match clause
binding an unused payload (e.g. an `Err` arm that never reads the message) is a real Go compile
error ("declared and not used") that C's own `__attribute__((unused))` has no direct Go
equivalent for — fixed with an explicit `_ = name` discard, same real bug class the C target
already guards against, just needing Go's own idiom instead. `go test`: 103/103 (73 prior + 30
new — string literals incl. real re-escaping, Ok/Err/Some/None construction, match on Result and
on Option, the real v0-boundary error, the duplicate-tag error, and a real end-to-end
`go build`-and-run test covering all four real branches). **Real, live, end-to-end proof**: a
real `safe-div`/`half-of-even` pair plus their own `match`-based callers compiled via `burrow
build`, linked into a real, separate Go module via `go build`, and run — correct output for all
four real cases (`describe-div(10,2)=5`, `describe-div(10,0)=-1`, `describe-half(8)=4`,
`describe-half(7)=-99`). Still real, honest, unstarted: `defenum`, `loop`, `Vec`, struct
*construction* (only `get-field` reads exist) — a real CLI needs `loop` for iteration and likely
`Vec` for building up output before "write a CLI in this" is fully true; this closes the second
real biggest blocker (real error handling), not the last one.

**`loop`/`recur` added to the Go emission target** (2026-09-03, cruise-queue card 9988's own
next-named prerequisite: "loop in particular is a real, necessary prerequisite for any real
iteration a CLI would need," from this same day's earlier match/Result changelog entry). Real,
deliberate v0 boundary, named explicitly, not silently limited: the loop body must be exactly one
top-level `(if cond then else)` with `recur` in exactly one branch (the other branch is the loop's
own terminal value) — the exact real shape every actual `.prn` loop in `stdlib/array.prn` uses
(`product`/`sum`: `(if (>= i n) acc (recur (+ i 1) ...))`). `recur` nested inside a deeper
`if`/`cond`/`match` chain is real, separate, unstarted work — PARENA's own mature `src/emit.c`
needed a `loop_locals` array threaded through `emit_loop_tail`/`emit_match_core`/nested-`if`
dispatch across several distinct bug-fix passes to get that fully general case right for its own
C target, the same real judgment call `match`'s own v0 boundary above already made. Real design:
bindings become real Go locals declared once before a real `for {}`; the terminal branch returns
(boxed through the same `any`/`RetType` mechanism every other construct in this file uses); the
`recur` branch computes every new binding value into its own temp variable BEFORE reassigning any
of them (a real simultaneous-assignment requirement — `(recur acc i)` swapping two loop vars would
silently break without it) and `continue`s. Bindings are evaluated into a cloned local-params map,
same leak-prevention discipline as `let`.

**Real, genuine bug found and fixed live testing this** (a real `(loop [i 0 acc 0] (if (> i n)
acc (recur ...)))` probe): `i := 0` lets Go infer `int`, not `int32`, for the same real reason the
`if` case's own branch-boxing needed an explicit `RetType(...)` conversion — an untyped integer
constant's own Go default doesn't match this emitter's `I32` → `int32` convention. Fixed by
declaring every non-string/non-bool loop binding as `var name int32 = expr` instead of `name :=
expr` (Go allows an untyped constant to implicitly convert to a var's own declared type) — a
real, honestly-named v0 boundary itself: every real loop binding in this stdlib today is I32, so
this is right for the actual current corpus and fails LOUD (a real Go compile error) rather than
silently wrong if a future non-I32 loop binding ever needs this path. `go test`: 88/88 total (7
new — recur in either branch, the no-leak proof, recur-in-both-branches and arity-mismatch
errors, the nested-if-body v0-boundary error, and a real end-to-end `go build`-and-run test).
**Real, live, end-to-end proof, not just unit tests**: a real `sum-to` (triangular-number loop)
compiled via `burrow build`, linked into a real, separate Go module via `go build`, and run —
correct output for `sum-to(0)=0`, `sum-to(1)=1`, `sum-to(10)=55`, `sum-to(100)=5050`. A second,
manual probe (`factorial`/`count-down-check`, exercising both branch orderings and a
single-binding loop) compiled via the real `burrow` CLI and run the same way, also correct
(`factorial(5)=120`, `factorial(0)=1`, `count-down-check(3)=true`). Still real, honest, unstarted:
`defenum`, `Vec`, struct construction, and `loop`/`match` beyond their own current v0 boundaries —
a real "write a CLI in this" bar still needs at least `Vec` for building up output.

**`Vec` added to the Go emission target** (2026-09-03, cruise-queue card 9988's own next-named
prerequisite after `loop`/`recur`). Real, direct port of PARENA's own runtime representation
SHAPE (`parena_runtime.h`'s own `Vec { Arena *arena; void **items; size_t count; size_t
capacity; }`, a boxed void*-array) to Go's own idiomatic equivalent: a bare `[]any` slice, no
wrapper struct at all — Go's own `append()` already does the exact real dynamic-growth job C's
own hand-rolled `vec_push_` does, and `any` already does the exact real "erase to a pointer-ish
box" job `void *` does. **Real, honest consequence, named directly**: this target's own header
comment previously claimed every v0-scope program is "already GC-irrelevant, no heap allocation
possible" — that stops being literally true the moment a real program uses `Vec` (`append` and
its backing array are real Go heap allocations); a host that adopted `debug.SetGCPercent(-1)` on
the strength of the original claim needs to know it no longer holds for `Vec`-using code.

Real, necessary prerequisite found live, not scoped in advance: every real `.prn` function that
builds a `Vec` takes a `dest : Arena @ Region` param and/or returns `(Vec ElemType) @ Region` —
v0 had NO parsing at all for either shape before this pass (a hard "no Arena/region annotations"
rejection). Both are now real, accepted: an `Arena @ Region` param is kept as a real, present Go
param typed `any` (unused for real work, but resolvable as a local so `(vec/new dest)` doesn't
error, and so a real Go host knows to pass a literal `nil` for it); a return type's own trailing
`@ Region` suffix is parsed and skipped, correctly shifting the body's own real AST index.

**Real, deliberate v0 boundary for `vec/get`, named explicitly**: this target has no per-Vec
element-type tracking (the same class of gap `match`'s own scrutinee restriction and `loop`'s own
binding-type fix already named), so `vec/get`'s result is always coerced to `int32` — right for
every current real `.prn` Vec usage this target actually needs (`array.prn`'s own shape/stride
vectors), but a real, named, NOT-yet-supported gap for a `(Vec SomeStruct)` (`bstree.prn`'s own
real `BSTNode`-valued Vec is the known, existing counter-example) — extending this needs a real
per-Vec-defn element-type registry, the same size of undertaking `defnRetInfo` is for
Result/Option. Out-of-bounds returns a real, honest `int32(0)` (this target's closest equivalent
to the C runtime's real NULL-on-OOB), via a comma-ok assertion, never a Go index-out-of-range
panic. `deref` is a real, honest no-op here (documented, not silently dropped): Go's own
`any`-boxed slice element already IS the value, with no separate reference layer C's own pointer
representation needs unwrapped.

**Three real, genuine bugs found and fixed live, not designed in advance**: (1) `append(vec, 1)`
boxes a bare literal as plain Go `int`, not `int32` — the SAME defaulting class already fixed
twice this same day (the `if` case's branches, `loop`'s own bindings), hitting a third real site;
fixed by explicitly wrapping every pushed item in `int32(...)`. (2) The real, common `.prn` idiom
`array.prn`'s own `zeros` uses — a side-effecting `loop` discarded via `_ =`, followed by the
REAL result (`(let [...] (loop [i 0] (when ...)) {...})`) — broke every nested `if`/`loop`'s own
internal `RetType(...)` boxing, which had always unconditionally used the ENCLOSING DEFN's own
return type even for an expression whose value is thrown away a statement later; a discarded
loop's own placeholder terminal value (e.g. bare `0`) then had to satisfy an unrelated return type
like `[]any`, a genuine Go compile error (`cannot convert 0 to type []any`). Fixed by giving every
non-final `do`/`let` statement its own scope with `retType` overridden to `any` — converting
anything to `any` is always legal Go, and asserting `any` back to `any` always trivially succeeds,
so a discarded expression's own internal boxing never needs to agree with whatever it's inside.
(3) `recur` appearing inside a `(do effect... (recur ...))` — the exact real shape a Vec-building
loop needs (`(do (vec/push! ...) (recur ...))`) — wasn't recognized as a recur branch at all
(only a BARE `(recur ...)` was); extended `loop`'s own recur-detection to also unwrap a `do`
whose own last expression is a direct recur call, running the `do`'s other expressions as real
effect statements before the recur reassignment.

`go test`: 93/93 total (5 new — an Arena param + region-annotated return type, real push!/len,
the fused `&v` reference-token shape, the invalid-mutation-target error, and a real end-to-end
`go build`-and-run test). **Real, live, end-to-end proof, not just unit tests**: a real program
building a `(Vec I32)` via a `loop`+`vec/push!` (the exact real `zeros`-style shape above), then
reading it back via `vec/len`+`vec/get`+`deref` in a second `loop`, compiled via `burrow build`,
linked into a real, separate Go module via `go build`, and run — correct output for `sum-of(5)=10`,
`sum-of(0)=0`, `sum-of(10)=45` (hand-computed triangular-number-style sums). Still real, honest,
unstarted: `defenum`, struct construction, `(Vec SomeStruct)` element types, and `loop`/`match`
beyond their own current v0 boundaries.

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

**`defenum` added to the Go emission target** (2026-09-03, cruise-queue card 9988's own
next-named prerequisite after match/Result/loop/Vec). Real, direct generalization of Result/
Option's own already-established `{Tag int; Value any}` shape to an arbitrary user-defined
tagged union — the exact same real design PARENA's own reference C emitter already uses for a
`defenum` (`src/emit.c`'s own `process_defenum`: "a tag enum plus a struct reusing Result/
Option's own {tag; void *value;} shape... a deliberate, honest generalization"), ported here
rather than invented fresh. One real, exported constructor function per variant
(`EnumName_VariantName`), matching the C target's own naming convention (PascalCase instead of
underscore-joined). A registered `defenum`'s own name is folded into the SAME `knownStructs` map
`defstruct` already populates — real, deliberate reuse, since `resolveGoType`'s job for a struct
name and an enum name is identical, so a `defenum` is usable anywhere a `defstruct` already was
(a `Result`'s own `ErrorType`, a param/return type) with zero new plumbing.

**Real, deliberate v0 boundary, named explicitly, not silently limited**: only zero- or
single-field variants are supported (a 2+-field variant is a real, separate, unstarted
extension, the same honest limitation `src/emit.c`'s own comment already names for "every
currently-real single-payload defenum in this stdlib"). `match` against a user-`defenum`-typed
scrutinee is ALSO real, separate, unstarted work — this pass only adds real value
CONSTRUCTION, the same "one bounded slice at a time" discipline `loop`/`Vec` before it already
followed. Still a real, useful increment on its own: a function returning
`(Result Payload MyError)` can now really construct `(Err SomeVariant)`, even though a caller
can't yet `match` on WHICH variant came back (only Ok-vs-Err, matching every real Result/Option
`match` this target already supports).

`go test`: 99/99 total (6 new — a bare zero-payload variant reference, a payload-carrying
variant call, the zero-payload-called-with-an-argument error, the 2+-field-variant v0-boundary
error, a `defenum` used in a `Result`'s own `ErrorType` position, and a real end-to-end
`go build`-and-run test). **Real, live, end-to-end proof, not just unit tests**: a real
`ParseError` `defenum` (`EmptyInput` zero-payload, `Invalid` single-payload) used as a
`Result`'s own `ErrorType`, constructed via both shapes, and consumed via a real, already-shipped
`match` (Ok-vs-Err) — compiled via `burrow build`, linked into a real, separate Go module via
`go build`, and run: correct output for `describe(5)=5` (real `Ok`), `describe(0)=-1`
(`EmptyInput` → `Err`), `describe(500)=-1` (`Invalid` → `Err`). Still real, honest, unstarted:
struct construction, `(Vec SomeStruct)` element types, `match` on a user `defenum`, and `loop`/
`match` beyond their own current v0 boundaries — cruise-queue card 9988's own literal CLI
deliverable still hasn't been started; this and the prior passes are real, necessary language-
feature prerequisites, not the deliverable itself.

**`burrow new` — real "batteries included" scaffolding command shipped** (2026-09-03, kanban
priority-queue card `PXCL-001`: "we need a batteries included cli tool to generate scaffolding
and stuff for us build it into burrow so it can help us manage both the go and prn side of
things"). `burrow new <name>` generates a real, immediately-runnable starter: `<name>.prn` (a
minimal PARENA decision-logic module), `main.go` (a real Go host importing its own compiled
`internal/burrowgen` package, the same real shape `IDUNA_PRO/cmd/idunapro` already established),
and `go.mod` — then actually runs the new `.prn` through `EmitGo` AND a real `go build ./...`
before returning success, so a broken scaffold is a real, honest failure in this command itself,
never silently handed to the user. Real, deliberate v0 scope: the Go target only — BURROW's own
real differentiator over plain `parena` (calling PARENA decision logic directly from a real Go
host, no cgo/FFI) is exactly the pattern worth scaffolding first; refuses to overwrite an
existing directory. `go test`: 101/101 (2 new — a full scaffold-build-and-RUN proof, and the
overwrite-refusal case). Real, live, end-to-end proof, not just unit tests: ran the actual
built `burrow` binary (`burrow new demo_mod`) in a real scratch directory, then `go run .`
against the result — printed the real, correct `Hello from demo_mod!`.

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
