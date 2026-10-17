.PHONY: api-run api-fmt api-test api-check

api-run:
	sh scripts/go.sh run ./cmd/api

api-fmt:
	sh scripts/go.sh fmt ./...

api-test:
	sh scripts/go.sh test -race ./...

api-check:
	sh scripts/check-api.sh

