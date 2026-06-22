.PHONY: dev down logs test fmt build

dev:
	docker compose up --build

down:
	docker compose down

logs:
	docker compose logs -f api worker

test:
	cd backend && go test ./...
	cd frontend && npm run lint && npm run build

fmt:
	cd backend && gofmt -w ./cmd ./internal

build:
	cd backend && go build ./cmd/api ./cmd/worker
	cd frontend && npm run build

