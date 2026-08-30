# BURROW

A feature-for-feature parallel implementation of the PARENA compiler (`parena-c`'s own real
lexer → parser → region analyzer → every emit target), written in a real combination of Go and
PARENA itself — not a new backend bolted onto the existing C compiler, a full rewrite proven
against it.

New repo (2026-08-30), NORTHSTAR only — no code yet. See `NORTHSTAR.md` for the full real scoping
pass: the real, founder-named acceptance bar ("can we write a pure golang and parena tool that
still pass all that parena c tests" — real behavioral parity against `parena-c`'s own existing
test corpus, not a freshly-invented one), the real relationship to PARENA's own already-in-progress
self-hosting effort (`selfhost/*.prn`), the GC-off design for BURROW's own new native Go emission
target, the "dogfood it" directive extending that same low-allocation discipline to `burrow`'s own
Go implementation, and the phased plan.

## Status

Phase 0 (architecture) confirmed, a real CLI shell shipped (`main.go`), and Phases 1-4 (lexer,
parser, region analyzer, and a real v0 C emitter) shipped: `lexer.go`, `parser.go`, `region.go`,
`emit_c.go`, real, faithful hand-ports of `PARENA/src/lexer.c`/`parser.c`+`ast.h`/`region.c`, plus
a new v0 C emitter matching `emit_ts.c`/`emit_java.c`'s own proven narrow scope — verified against
all 30 real test scenarios ported from PARENA's own C reference test suite, plus a full real-corpus
stress test (all 111 `.prn` files in `stdlib`+`selfhost` parse and region-analyze clean) and a real
end-to-end proof: `burrow build` on `PAPERCRAFT/stdlib/papercraft/level_mod.prn`, actually compiled
with `gcc`, actually run against `level_mod_test.c`'s own real assertions — all pass. `go build`/
`go vet`/`go test` all clean, 38/38. `burrow` is installed on `PATH`. Full `emit.c` parity
(structs/enums/`match`/`loop`/`Result`/`Vec`) and the TypeScript/Java targets remain real, honest,
unstarted work.

**`DUNG` is its own separate repo** (`github.com/emilyspringerton/DUNG`) — "the BURROW editor," a
unified terminal emulator + editor rewriting `PITVIPER` and PARENA's own `stdlib/editor/*.prn`.
Real, load-bearing relationship to this repo: DUNG's own real build compiles its ground-up PARENA
editor source via the real `burrow` CLI this repo builds. Its own real build now works for the
same narrow, scalar v0 scope this repo's own Phase 4 just proved — a real, concrete design
constraint DUNG's own editor port needs to stay within until full `emit.c` parity lands. First
scoped inside this repo (`DUNG.md`, now removed), corrected by the founder into its own
standalone, Bazel-built repo — see `DUNG/NORTHSTAR.md` there for the full real scoping pass.

## Related

- `PARENA` — the compiler this is a full, parallel rewrite of; `selfhost/*.prn` is the real,
  already-started self-hosting source tree this project leans on directly.
- `DUNG` — the real, separate repo whose own build depends on this repo's own emit capability —
  see above.
- `GoblinFoxDragon` — the named candidate real Go host for the GC-off-safe Go emission target
  (Phase 6 in `NORTHSTAR.md`, not committed to yet).
- `EMILY` — RSI loop / backlog coordination for cross-repo work.
