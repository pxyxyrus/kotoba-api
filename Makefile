.PHONY: api-run api-fmt api-test api-check

api-run:
	sh scripts/go.sh run ./cmd/api

api-fmt:
	sh scripts/go.sh fmt ./...

api-test:
	sh scripts/go.sh test -race ./...

api-check:
	sh scripts/check-api.sh

.PHONY: db-up db-down db-status api-run-db
db-up:
	docker compose up -d --wait mongo

db-down:
	docker compose down

db-status:
	docker compose ps

api-run-db:
	KOTOBA_MONGO_URI="$${KOTOBA_MONGO_URI:-mongodb://127.0.0.1:27017}" sh scripts/go.sh run ./cmd/api
