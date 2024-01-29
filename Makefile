# Alex Mackay 2024

build:
	GO111MODULE=on go build -ldflags "-w -linkmode external -extldflags '-static' -X 'github.com/ATMackay/psql-ledger/service.buildDate=$(shell date +"%Y-%m-%d %H:%M:%S")' -X 'github.com/ATMackay/psql-ledger/service.gitCommit=$(shell git rev-parse --short HEAD)'" ./cmd/psqlledger
	mv psqlledger ./build

run: build
	./build/psqlledger

# Must have Docker installed on the host machine

docker:
	cd docker && ./build.sh


# For local stack testing
# Assummes that docker is installed on the host machine 

postgres:
	docker run --name postgres -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:latest

createdb:
	docker exec -it postgres createdb --username=root --owner=root bank

dropdb: 
	docker exec -it postgres dropdb --username=root bank

# Requires migrate installation: https://github.com/golang-migrate/migrate/tree/master/cmd/migrate
migrateup:
	migrate -path database/sqlc/migrations -database "postgresql://root:secret@localhost:5432/bank?sslmode=disable" -verbose up

migratedown:
	migrate -path database/sqlc/migrations -database "postgresql://root:secret@localhost:5432/bank?sslmode=disable" -verbose down

# Requires sqlc installation: https://docs.sqlc.dev/en/stable/overview/install.html
sqlc: 
	cd database/sqlc && sqlc generate && mv ./db/* ..

.PHONY: build docker postgres createdb dropdb migrateup migratedown sqlc run