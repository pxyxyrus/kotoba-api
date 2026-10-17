# Kotoba API

Shared product requirements, implementation status, and milestones: [Project plan](docs/project-plan.md).

Dedicated Go/Gin backend for a personal Japanese wordbook with Korean meanings and an English interface. The frontend belongs in the separate kotoba-web repository.

## Run

Install Go with toolchain download support, then from this repository:

```sh
make api-run
```

This starts validation-only mode on http://127.0.0.1:8080. scripts/go.sh uses ignored project-local caches and the toolchain in go.mod. Go 1.26.0 is the module minimum; selected toolchain is 1.26.8.

For local MongoDB, install/start a Docker engine, then:

```sh
make db-up
make api-run-db
```

Compose runs mongo:8.0 on loopback port 27017 with a persistent named volume. The image tracks patch releases; a digest is not pinned. This is unauthenticated local development only. make db-down preserves the volume. Do not remove volumes unless you intend to erase data.

KOTOBA_ADDR overrides the HTTP address. KOTOBA_MONGO_URI enables MongoDB; api-run-db defaults it to mongodb://127.0.0.1:27017. The example environment file is not automatically loaded. Connection startup is bounded, a single client is reused, and shutdown disconnects it. Invalid/unreachable configured databases prevent startup. Secrets are not included in connection errors.

## API

- GET /healthz: process liveness.
- GET /readyz: database ping; 200 when ready, 503 when unconfigured/unavailable.
- POST /api/v1/words/validate: normalize/check a draft, without saving.

```sh
curl -H 'Content-Type: application/json' \
  -d '{"japanese":"懐かしい","reading":"なつかしい","meaning_ko":"그립다","pronunciation_confirmed":true}' \
  http://127.0.0.1:8080/api/v1/words/validate
```

Success returns status=validated, persisted=false, and the normalized word. Optional context is supported. Strings are trimmed; limits in Unicode code points are Japanese 100, reading 200, meaning 500, context 2000. JSON bodies are capped at 16 KiB. Errors use an error object containing code/message and optional fields. Invalid binding returns 400; domain errors return 422. Unknown JSON fields are ignored. Confirmation is a user assertion, not proof of linguistic correctness.

## Code map

- cmd/api: startup, configuration, lifecycle.
- internal/httpapi: Gin transport, validation endpoint, readiness, error responses.
- internal/vocabulary: pure domain validation.
- internal/database: official MongoDB driver connection lifecycle.
- compose.yaml: local development database.

## Checks

```sh
make api-fmt
make api-test
make api-check
```

Verified in the source workspace: build, vet, race tests, Staticcheck, Compose configuration, and govulncheck (zero reachable findings; one imported-package and one required-module finding outside detected calls). MongoDB live verification is pending: Docker engine was unavailable and the local Docker application failed to launch. Unit tests cover readiness states, bounded contexts, configuration failure, and cancellation; these are not substitutes for a real database integration test.

## Next slice

No persistence, auth, save/list endpoints, pronunciation/speech providers, review scheduling, or deployment exists. Proposed first collection: kotoba.words, with stable record ID, client operation ID for retry safety, Japanese text, reading, Korean sense, context, schema version, and UTC timestamps. Add a unique operation-ID index and a created-at/ID list index with the first storage operations. Do not make Japanese spelling or Korean gloss unique: distinct senses must remain possible.

Reference: [official MongoDB Go driver](https://www.mongodb.com/docs/drivers/go/current/connect/mongoclient/).
