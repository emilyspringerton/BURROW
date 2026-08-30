## 2026-08-30

- NORTHSTAR scoping pass: PARENA's fourth compilation target (Go, after C/TypeScript/Java). Real answer to the GC-off ask: v0 generated functions never allocate on the Go heap by construction, making debug.SetGCPercent(-1)/GOGC=off a safe host choice. Two Go-specific emitter differences flagged (if -> IIFE, conditional math/rand import). Scoping only, no emit_go code yet. (sess-20260825-1938-f6bd411e)

