# DUNG — the BURROW editor (terminal emulator + editor, unified)

## Where this came from

Founder real-time, 2026-08-30, immediately after BURROW's own Phase 1/2 (lexer + parser parity)
shipped:

> "ok and rewrite pitviper and the parena editor into burrow" → "call it DUNG" → "aka the burrow
> editor" → "and terminal emulator" → "terminal emulator comes down as a visor" → "and there is
> also split pane use i3 primatives"

**DUNG is a new, real sub-component living inside the BURROW repo**: a unified terminal emulator
+ editor, rewriting `PITVIPER`'s own real terminal capability and PARENA's own real editor
(`stdlib/editor/*.prn`) into one thing, in Go + PARENA — matching BURROW's own already-established
"lean on PARENA primitives when it helps" architecture (`NORTHSTAR.md`'s own real, confirmed
direction), not a separate, disconnected idea.

**Scoping only, this pass** — same real discipline every other real scope addition in this
monorepo gets before code, doubly true here given the real scale of both real things being
rewritten (see "Real scale, honestly" below). No DUNG code exists yet.

## Real, explicit non-goal: PITVIPER itself is untouched

**PITVIPER's own Go implementation stays exactly as-is** — this is a real, new, separate rewrite
inside BURROW, not a rewrite-in-place of the existing repo. This is the exact same real precedent
`SAND` already set relative to `PITVIPER` (`/home/fatbaby/CLAUDE.md`'s own SAND row: "PITVIPER's
own Go implementation is unaffected and stays exactly as-is — SAND is a separate, new fork, not a
rewrite-in-place"). DUNG is a second, independent instance of that same real pattern, this time
living inside BURROW instead of a dedicated new repo — real, deliberate placement call: BURROW is
already the real "Go + PARENA, leans on PARENA primitives" project this belongs to structurally,
not a coincidence.

## What DUNG actually is

A single, real, unified application combining:

1. **A real terminal emulator** — the real successor to `PITVIPER`'s own SDL2 terminal (PTY
   handling, glyph rendering, vterm/scrollback), with one real, new UX requirement named directly
   by the founder: **it comes down as a visor** — a real, quake-style drop-down overlay (the same
   real UX `Guake`/`Yakuake`/`Tilda` already establish: hotkey-triggered, slides down from the top
   of the screen, slides back up to hide), not a plain, permanently-windowed terminal the way
   `PITVIPER`'s own current `cmd/pitviper/main.go` is.
2. **A real editor** — the real successor to PARENA's own `stdlib/editor/*.prn` (buffer
   management, TextMate-based syntax highlighting + theming, a Vim-modal keybinding model, a
   Spotlight/`Ctrl+T` command palette with a real, hot-swappable plugin-backend API, a reusable
   widget system) — real, substantial, already-designed PARENA-side logic, not started from a
   blank page.
3. **A real, shared, i3-primitive split-pane layout system** — the founder's own explicit
   correction over `PITVIPER/docs/NORTHSTAR.md`'s own existing, real, but different Milestone 3
   plan (vim-style `Ctrl+W |`/`Ctrl+W -` splits + `Ctrl+W h/j/k/l` navigation) — DUNG's own real
   pane model follows i3's own real tiling primitives instead: a tree of containers, each one
   real, binary horizontal-or-vertical split, panes and splits nested arbitrarily, real focus
   movement between panes in the tree. This one shared layout system serves BOTH the terminal
   panes and the editor panes, since they're now one real, unified application — a real,
   structural reason DUNG benefits from being one program, not two separate ones glued together.

## Why this is architecturally coherent, not a forced merger

Checked before writing this doc, not assumed: `PITVIPER` and PARENA's own editor **already share
the identical real technical foundation** — `stdlib/editor/ui.prn`'s own header comment states it
directly: "gutter/diagnostics/status-bar/popups... render via sdl2's own real primitives..., same
architecture as PITVIPER's own SDL2 terminal-emulator precedent, not a new toolkit." Both are
already SDL2-based, both already share real event/keybinding-model thinking (`stdlib/editor/
events.prn`'s own header comment cites PITVIPER's own real "ConPTY + mouse-drag-selection +
clipboard" work as its own architecture's precedent directly). DUNG isn't inventing a merger
between two unrelated things — it's the real, natural next step of two things that were already
converging.

## Real UX foundation — adopted directly, not invented (`EmilyOS/docs/legacy-archive/gui-v0.1-design-capture.md`)

Founder real-time: "EMILY os design primatives ux and affordances." Found a direct, real match
before writing anything new: `EmilyOS/docs/legacy-archive/gui-v0.1-design-capture.md` is a
complete, already-written GUI v0.1 spec whose own layout model is explicitly named **"tmux × i3
hybrid"** — predating and directly matching the founder's own "use i3 primitives" instruction for
DUNG above. This is adopted as DUNG's own real, authoritative UX foundation directly, not
re-derived or reinvented. The real, load-bearing pieces (full detail in that doc, not repeated
here in full):

- **Visual system**: a real, hard-coded palette — deep navy/near-black background, EGSHELL
  (white-ish) reserved exclusively for directory/container tiles, colored (never white) button
  tiles darker than EGSHELL, all-caps blocky typography, single-shot (never looping) activation
  flashes only — "no animations that imply aliveness."
- **Interaction contract — no single-click actions, ever**: single click only focuses/selects; a
  real, deterministic **fast double-click (≤220ms) activates**, a real, deterministic **slow
  double-click (350-800ms) enters label-edit mode** — the same real two-speed semantics apply to
  both files and directories. Real, exact timing thresholds already specified
  (`DC_FAST_MAX=220ms`, `DC_SLOW_MIN=350ms`/`DC_SLOW_MAX=800ms`), not left to be invented later.
- **Layout — the real "tmux × i3 hybrid" panes**: the screen is a real tiling space; panes hold a
  directory view, a file view, a process/log panel, or a command-line/verb-bar panel; real
  keyboard-first pane operations (vertical/horizontal split, focus movement, resize, non-
  destructive close). No freeform dragging by default — a real, deliberate departure from the
  founder's own later "drag another file onto it" message below, which this doc treats as a real,
  explicit, DUNG-specific EXCEPTION to that default (drag-to-load-file-as-new-chat-pane is exactly
  the kind of "deliberate verb, not casual click-drag" the original spec's own §2.3 already
  permits opting into).
- **Safety/posture hooks, real and already EmilyOS-integrated**: postures (`NORMAL`/`SIEGE`/
  `MERCY`/`INCIDENT`/`GAME`, `EmilyOS/internal/posture/`) gate which actions are AVAILABLE, never
  the visuals themselves; a denied action gets a real, quiet, one-frame "deny flash" — explicitly
  **no modal dialogs, no toast notifications**. Real, direct real relevance to DUNG: EmilyOS is
  already named as PITVIPER's own real, related repo ("PITVIPER is the operator interface" per
  EmilyOS's own CLAUDE.md) — DUNG inheriting this same posture-aware interaction contract is a
  real, natural continuation of that existing relationship, not a new one being forced.

**Real, new addition this pass, reconciled with the original spec's own real "no freeform
dragging by default" rule**: the founder's own later message — "like the editor and the terminal
in the same window and then we like slurp in another file via drag another file onto it or load
it from internet it opens up a new chat window inside the shared window" — confirms DUNG is one
real, shared window (editor + terminal + chat panes together, matching this whole doc's own
"unified application" framing), and names a real, new pane TYPE the original EmilyOS spec doesn't
yet have: a **chat pane** (the real Emily Prime pane `PITVIPER`'s own `docs/NORTHSTAR.md` already
plans as its own Milestone 4), opened by dragging a file onto the DUNG window OR loading one from
a URL — the dropped/loaded file becomes that new chat pane's own real, scoped context. A real,
explicit verb-shaped action (drag-to-load), not silently violating the original spec's own
no-casual-drag default.

## Real architecture: Go + PARENA, same split BURROW's own compiler-rewrite already uses

Same real "PARENA owns the decision logic, Go host owns the plumbing" split every mod in this
monorepo already uses (PAPERCRAFT/GTA7/TYLER's own real mods, applied here to an editor/terminal
instead of a game):

- **Go, real and direct** (matching `PITVIPER`'s own already-proven real stack, not reinvented):
  SDL2 window/event-loop/renderer via cgo, FreeType2 glyph rendering, PTY via `openpty(3)`, the
  real vterm/scrollback state machine. This is real, substantial systems-level plumbing SDL2/cgo
  needs a real host language for — not a PARENA-native target either project has ever attempted.
- **PARENA, where it already fits** — `stdlib/editor/*.prn`'s own real, existing logic is the
  real, direct source of truth for: buffer manipulation semantics, TextMate tokenization + theming
  rules, the Spotlight fuzzy-match/plugin-dispatch algorithm, the widget system's own state
  transitions. Real, honest current blocker, same one BURROW's own Phase 6 (native Go emission
  target) and the TYLER cutscene mod (`PARENA/stdlib/tyler/cutscene_mod.prn`'s own header comment)
  already name: there is no PARENA-to-Go compiler yet, so these `.prn` files can't be compiled
  directly into DUNG's own Go binary today. Two real, honest interim paths, not resolved by this
  doc: (a) hand-port the real logic these files already encode into Go directly, the same real
  fallback BURROW's own Phase 1/2 (lexer/parser) already used successfully, or (b) cgo against the
  real, already-compiled C output PARENA's own existing C emitter produces (`editor-demo`'s own
  real, already-built binary is direct proof this C output already works end to end). **Not
  decided here — a real Phase 0 question for DUNG's own next scoping pass**, same as BURROW's own
  original "selfhost/*.prn vs. hand-write" question was before the founder settled it.

## Real scale, honestly

- `PITVIPER`: **4066+ real lines of already-working Go** — `cmd/pitviper/main.go` (689 lines per
  its own CLAUDE.md status note), `internal/vterm` (842 lines + 555 lines of its own tests),
  `internal/pty`, `internal/font` (218 lines + emoji/shiny variants), `internal/gfdapi` (235
  lines), `internal/mudconn`, `internal/scrollmod`. Real PTY/FreeType2/GFD-API integration already
  proven live, not a prototype.
- PARENA's editor: **13 real `.prn` files** (`buffer`/`render`/`spotlight`/`textmate`/
  `textmate_markdown`/`textmate_parena`/`textmate_loader`/`theme`/`ui`/`widget`/`plugin`/`events`/
  `construct_split`) plus a real, already-built `editor-demo` binary (`PARENA/editor-demo`,
  `make editor-demo`/`editor-demo-smoke` targets already exist).
- **A full rewrite of both, in one sitting, is not realistic** — the same honest scale assessment
  BURROW's own `NORTHSTAR.md` already makes for the compiler-rewrite itself ("this is a genuinely
  large, multi-phase undertaking, not a same-afternoon addition"). This doc's own real job is
  scoping the real shape and the real first slice, not claiming a finished design.

## Real, phased plan (a real, separate track from BURROW's own compiler-parity phases 1-6)

**DUNG Phase 0 — settle the real Go-vs-PARENA boundary question** named above (hand-port vs. cgo
against `editor-demo`'s own real C output) — a real, load-bearing decision this doc deliberately
does not make unilaterally.

**DUNG Phase 1 — the real, smallest proof point**: a real SDL2 window that (a) renders a visor-
style drop-down terminal pane (PTY + vterm, hand-ported or vendored from `PITVIPER`'s own real,
already-working `internal/vterm`/`internal/pty`) triggered by one real hotkey, and (b) can split
that one pane into two via one real i3-primitive (horizontal or vertical) — matching the same
"smallest real proof point" discipline (`NORTHSTAR.md`'s own original "a player can log in and
spawn... nothing else") every other real Phase 0/1 in this monorepo already follows. No editor
pane yet — proving the visor + i3-split mechanics first, on the simpler of the two real domains.

**DUNG Phase 2+ (design only, not detailed here)**: the real editor pane (buffer + TextMate
highlighting + Spotlight), real plugin-API parity with `stdlib/editor/plugin.prn`'s own design,
Emily Prime pane integration (matching `PITVIPER`'s own already-planned Milestone 4).

## Real risks and open questions, named honestly

- **The real Go-vs-PARENA boundary (Phase 0) is genuinely unresolved** — flagged, not guessed at.
- **cgo linking against real PARENA-emitted C** (if that path is chosen) needs the same real
  `parena_runtime.h`/`parena_runtime.c` bridging every other real PARENA-mod host in this monorepo
  already uses (PAPERCRAFT/GTA7/SHANKPIT's own real precedent) — real, proven pattern, not
  invented, but not yet applied to a GUI/SDL2 host specifically.
- **i3's own real tiling model is more general than a single visor-terminal-plus-editor app
  strictly needs** (i3 manages whole windows/workspaces across an entire desktop session) — DUNG
  only needs the real CONTAINER-SPLIT primitive (binary horizontal/vertical tree, real focus
  movement), not i3's own window-manager-level concerns (multi-monitor, workspace switching,
  floating windows) — named so "use i3 primitives" isn't silently over-scoped into "reimplement
  i3."
- **No real acceptance bar named yet for DUNG** (unlike BURROW's own compiler-rewrite, which has
  the founder's own explicit "pass all that parena c tests" bar) — a real, open question for
  DUNG's own next scoping pass, not invented here.
- **The chat-pane's own real backend contract isn't specified yet** — presumably the same real
  `EMILY_AGENT_URL` / Emily Prime agent connection `PITVIPER/docs/NORTHSTAR.md`'s own Milestone 4
  already plans, but not confirmed for DUNG specifically, and "load [a file] from internet" as a
  chat-context source (vs. a local drag-drop) has no real fetch/security model specified yet
  either — flagged, not designed here.

## Related

- `PITVIPER/CLAUDE.md`, `PITVIPER/docs/NORTHSTAR.md` — the real, existing terminal emulator DUNG
  rewrites (unaffected, stays as-is — see this doc's own "explicit non-goal" section above); its
  own Milestone 4 (Emily Prime pane) is the real precedent for DUNG's own new chat-pane feature.
- `PARENA/stdlib/editor/*.prn` — the real, existing editor logic DUNG rewrites; `PARENA/
  editor-demo` is the real, already-built proof this logic already compiles and runs.
- `EmilyOS/docs/legacy-archive/gui-v0.1-design-capture.md` — the real, already-written,
  authoritative UX/interaction spec DUNG adopts directly (visual system, tmux×i3 pane layout,
  double-click-speed interaction contract, posture-aware non-modal safety feedback).
- `EmilyOS/internal/posture/` — the real posture state machine (`NORMAL`/`SIEGE`/`MERCY`/
  `INCIDENT`/`GAME`) DUNG's own denial-feedback UX (per the design-capture doc above) would need
  to read from for a real, live integration, not just visually mimic.
- `BURROW/NORTHSTAR.md` — the real "lean on PARENA primitives when it helps" architecture DUNG
  inherits directly, and the real hand-port-fallback precedent (Phases 1-2) DUNG's own Phase 0
  question will likely resolve the same way.
- `SAND` — the real precedent for "a new, separate fork of an existing mission, not a
  rewrite-in-place," applied here a second time (BURROW/DUNG relative to PITVIPER, the same real
  shape SAND is relative to PITVIPER for the editor-only half of that same mission).
