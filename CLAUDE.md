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

Real CLI shell shipped (`main.go`, full command parity with `parena`'s own surface). Phase 1
(lexer parity) shipped: `lexer.go` + `lexer_test.go`, a real hand-port of `PARENA/src/lexer.c`
verified against `PARENA/tests/test_selfhost_lexer.c`'s own real assertions — see `NORTHSTAR.md`'s
own Phase 1 entry for the real architecture call (hand-port, not a PARENA-Go emitter — the
language surface `selfhost/lexer.prn` needs is well beyond any existing PARENA emitter's proven
scope). Parser/region-analyzer/emitter parity not started.

## Related Repos

- `PARENA` — the language and compiler this is a fourth compilation target for; `src/emit_ts.c`
  and `src/emit_java.c` are the real, proven v0 emitter template this project follows.
- `GoblinFoxDragon` — the named candidate real Go host for a GC-off-safe PARENA-compiled decision
  layer (Phase 2 in `NORTHSTAR.md`, not committed to yet).
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
