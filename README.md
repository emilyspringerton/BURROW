# BURROW

PARENA's fourth real compilation target: Go, designed around a GC-off-safe host.

New repo (2026-08-30), NORTHSTAR only — no emitter code yet. See `NORTHSTAR.md` for the full real
scoping pass: what "designed to run with the gc turned off" actually means for an emitter (not a
special codegen mode — a scope discipline: v0-scope generated functions never allocate on the Go
heap, so a host that disables GC around calls into them is making a real, safe, informed choice),
the real Go-specific emitter differences from PARENA's own existing TypeScript/Java targets
(`if`-as-immediately-invoked-closure, conditional `math/rand` import), and the phased plan.

## Status

Scoping only. `PARENA/src/emit_ts.c` and `PARENA/src/emit_java.c` are the real, proven v0 emitter
template this project follows — read `NORTHSTAR.md` before writing any `emit_go` code.

## Related

- `PARENA` — the language and compiler this is a fourth target for.
- `GoblinFoxDragon` — the named candidate real Go host (not committed to yet, see `NORTHSTAR.md`).
- `EMILY` — RSI loop / backlog coordination for cross-repo work.
