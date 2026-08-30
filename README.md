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

Scoping only. No Go or PARENA code specific to this project exists yet.

## Related

- `PARENA` — the compiler this is a full, parallel rewrite of; `selfhost/*.prn` is the real,
  already-started self-hosting source tree this project leans on directly.
- `GoblinFoxDragon` — the named candidate real Go host for the GC-off-safe Go emission target
  (Phase 5 in `NORTHSTAR.md`, not committed to yet).
- `EMILY` — RSI loop / backlog coordination for cross-repo work.
