.PHONY: build run-api run-scheduler run-worker run-dashboard clean db-up db-migrate db-reset sqlc-generate test test-integration test-unit

sqlc-generate:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest generate

build: sqlc-generate
	go build -o bin/api-server ./cmd/api-server
	go build -o bin/scheduler ./cmd/scheduler
	go build -o bin/worker ./cmd/worker
	go build -o bin/dashboard ./cmd/dashboard

run-api:
	go run ./cmd/api-server

run-scheduler:
	go run ./cmd/scheduler

run-worker:
	go run ./cmd/worker

run-dashboard:
	go run ./cmd/dashboard

db-up:
	docker-compose up -d

db-migrate:
	go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest -path db/migrations -database "postgres://flux:flux123@localhost:5432/flux?sslmode=disable" up

db-rollback:
	go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest -path db/migrations -database "postgres://flux:flux123@localhost:5432/flux?sslmode=disable" down 1

db-create-migration:
	go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest create -ext sql -dir db/migrations -seq $(name)

db-reset:
	docker-compose down -v && docker-compose up -d && sleep 5 && make db-migrate

test: test-unit test-integration

test-unit:
	go test -v -race -short ./...

test-integration:
	go test -v -race ./internal/adapters/database/... ./internal/adapters/http/...

test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -rf bin/ coverage.out coverage.html