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

Phase 0 (architecture) confirmed, a real CLI shell shipped (`main.go`), and Phases 1-2 (lexer +
parser parity) shipped: `lexer.go`/`parser.go`, real, faithful hand-ports of `PARENA/src/lexer.c`
and `PARENA/src/parser.c`+`ast.h`, verified against all 23 real test scenarios from `PARENA/
tests/test_selfhost_lexer.c` + `PARENA/tests/test_lexer_parser.c` (ported to `lexer_test.go`/
`parser_test.go`) — `go build`/`go vet`/`go test` all clean. Also stress-tested against all 111
real `.prn` files in `PARENA/stdlib`+`PARENA/selfhost`: zero parse failures, and structurally
cross-checked node-for-node against the real C reference's own `parena parse` dump (exact match).
Real architecture call made getting there, documented in `NORTHSTAR.md`'s own Phase 1/2 entries: a
hand-port, not a PARENA-compiled-to-Go emitter (the language surface `selfhost/lexer.prn` itself
uses is well beyond any existing PARENA emitter's proven scope). Region-analyzer/emitter parity
(Phases 3-4) not started.

**`DUNG` is its own separate repo** (`github.com/emilyspringerton/DUNG`) — "the BURROW editor," a
unified terminal emulator + editor rewriting `PITVIPER` and PARENA's own `stdlib/editor/*.prn`.
Real, load-bearing relationship to this repo: DUNG's own real build compiles its ground-up PARENA
editor source via the real `burrow` CLI this repo builds — DUNG is BURROW's own real, live,
flagship dogfooding consumer, and its own build is gated on this repo's own Phase 3-4 (region
analyzer + emitter parity, not started) landing enough real emit capability. First scoped inside
this repo (`DUNG.md`, now removed), corrected by the founder into its own standalone, Bazel-built
repo — see `DUNG/NORTHSTAR.md` there for the full real scoping pass.

## Related

- `PARENA` — the compiler this is a full, parallel rewrite of; `selfhost/*.prn` is the real,
  already-started self-hosting source tree this project leans on directly.
- `DUNG` — the real, separate repo whose own build depends on this repo's own emit capability
  (Phase 3-4) — see above.
- `GoblinFoxDragon` — the named candidate real Go host for the GC-off-safe Go emission target
  (Phase 6 in `NORTHSTAR.md`, not committed to yet).
- `EMILY` — RSI loop / backlog coordination for cross-repo work.
