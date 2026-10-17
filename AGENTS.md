# Backend agent instructions

This is the standalone kotoba-api repository. Read README.md and inspect source before edits. Go/Gin serves a separately maintained Next.js frontend (kotoba-web).

Use make api-fmt, make api-test, and make api-check. scripts/go.sh maintains local ignored Go caches. Keep Gin types in internal/httpapi and MongoDB I/O in internal/database; pass standard contexts across boundaries. Never log connection secrets or word payloads. Prefer small concrete Go types and meaningful behavior tests.

Current endpoints: /healthz liveness, /readyz database readiness, POST /api/v1/words/validate. Validation does not persist. MongoDB connection is optional via KOTOBA_MONGO_URI; configured startup must verify connectivity. Local Compose uses persistent storage and loopback-only port binding. No auth or production deployment exists.

English UI/operational messages, Japanese words and readings, Korean meanings. Word capture is three steps with saving on step three. Storage endpoints, indexes, retry-safe operation identity, and actual database integration tests remain pending. Do not deduplicate solely by Japanese spelling or Korean gloss. Do not claim live MongoDB verification until run against a real instance.

Update README.md with actual checks and limitations. Do not commit secrets, generated binaries, caches, or .env files. Preserve existing toolchain pins unless a change is justified.
