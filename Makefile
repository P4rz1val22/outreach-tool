include .env
export

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Usage: make migrate-create name=migration_name"; \
		exit 1; \
	fi
	migrate create -ext sql -dir db/migrations -seq $(name)

migrate-up:
	migrate -database "$(DATABASE_URL)" -path db/migrations up

migrate-down:
	migrate -database "$(DATABASE_URL)" -path db/migrations down 1

sqlc-gen:
	sqlc generate

mock-gen:
	mockgen -source=db/sqlc/querier.go -destination=mocks/querier_mock.go -package=mocks

gen: sqlc-gen mock-gen

db-update: migrate-up generate

db-up:
	supabase start

db-down:
	supabase stop

build:
	go build -o app ./cmd/server

run:
	go run ./cmd/server

test:
	go test -v ./...