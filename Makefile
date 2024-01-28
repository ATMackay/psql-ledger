# Alex Mackay 2024

build:
	GO111MODULE='auto' go build ./cmd/psqlledger
	mv psqlledger ./build

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
	docker exec -it postgres dropdb --username=root --owner=root bank

migrateup:
	migrate --path psql/migration --databse "postgresql://root:secret@localhost:5432/bank?sslmode=disable" --verbose up

migrateup:
	migrate --path psql/migration --databse "postgresql://root:secret@localhost:5432/bank?sslmode=disable" --verbose down

# Requires sqlc installation: https://docs.sqlc.dev/en/stable/overview/install.html
sqlc: 
	cd database/sqlc && sqlc generate && mv ./db/* ..

.PHONY: postgres createdb dropdb migrateup migratedown