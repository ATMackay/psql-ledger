# Alex Mackay 2024

# Git based version
VERSION ?= $(shell git describe --tags)
COMMIT ?= $(shell git rev-parse --short HEAD)
BUILD_DATE ?= $(shell date -u +'%Y-%m-%d %H:%M:%S')
COMMIT_DATE ?= $(shell git show -s --format="%ci" $(shell git rev-parse HEAD))

# Build folder
BUILD_FOLDER = build

build:
	@go build -o $(BUILD_FOLDER)/psqlledger -v -ldflags=" -X 'github.com/ATMackay/psql-ledger/service.commitDate=$(COMMIT_DATE)' -X 'github.com/ATMackay/psql-ledger/service.buildDate=$(BUILD_DATE)' -X 'github.com/ATMackay/psql-ledger/service.version=$(VERSION)' -X 'github.com/ATMackay/psql-ledger/service.gitCommit=$(COMMIT)'" ./cmd/psqlledger

run: build
	@cd build && ./psqlledger

test: 
	@go test -cover ./service

test-stack: 
	@go test -cover ./integrationtests

# Must have Docker installed on the host machine

docker:
	@cd docker && ./build.sh && cd ..

docker-compose: docker
	@cd docker-compose && docker-compose up


# For local stack testing
# Assumes that docker is installed on the host machine 

postgresup:
	@docker run --name postgres -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:latest

postgresdown:
	@docker kill postgres && docker remove postgres

createdb:
	@docker exec -it postgres createdb --username=root --owner=root bank

dropdb: 
	@docker exec -it postgres dropdb --username=root bank

# Requires migrate installation: https://github.com/golang-migrate/migrate/tree/master/cmd/migrate
migrateup:
	@migrate -path sqlc/migrations -database "postgresql://root:secret@localhost:5432/bank?sslmode=disable" -verbose up

migratedown: createdb
	@migrate -path sqlc/migrations -database "postgresql://root:secret@localhost:5432/bank?sslmode=disable" -verbose down

# Requires sqlc installation: https://docs.sqlc.dev/en/stable/overview/install.html
sqlc: 
	@cd sqlc && sqlc generate

.PHONY: build docker postgres createdb dropdb migrateup migratedown sqlc run