TEST_DSN ?= postgres://postgres:postgres@localhost:5433/taskservice_test?sslmode=disable

.PHONY: test-db-up test-db-down test-db-reset test-repo test-repo-one

test-db-up:
	docker compose -f docker-compose.test.yml up -d

test-db-down:
	docker compose -f docker-compose.test.yml down

test-db-reset:
	docker compose -f docker-compose.test.yml down -v
	docker compose -f docker-compose.test.yml up -d

test-repo:
	mkdir -p .gocache
	TEST_DATABASE_DSN="$(TEST_DSN)" GOCACHE="$$(pwd)/.gocache" go test ./internal/repository/postgres -v

test-repo-one:
	mkdir -p .gocache
	TEST_DATABASE_DSN="$(TEST_DSN)" GOCACHE="$$(pwd)/.gocache" go test ./internal/repository/postgres -run TestRecurrenceCreate_SpecificDates -v
