.PHONY: dev down logs test lint fmt build migrate seed

dev:
	docker compose up --build

down:
	docker compose down

logs:
	docker compose logs -f api worker

test:
	cd backend && go test ./...
	cd frontend && npm run lint && npm run build

lint:
	cd backend && test -z "$$(gofmt -l ./cmd ./internal ./migrations)" && go vet ./...
	cd frontend && npm run lint

fmt:
	cd backend && gofmt -w ./cmd ./internal ./migrations

build:
	cd backend && go build ./cmd/...
	cd frontend && npm run build

migrate:
	docker compose run --rm --build migrate

seed:
	docker compose run --rm --build seed
