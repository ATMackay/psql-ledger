# A simple transaction ledger implemented in Go and PostgreSQL

This project contains packages for creating a web backend stack for the purpose of recording transactions between users.

Components

* Go HTTP microservice built with [Gorilla mux](https://github.com/gorilla/mux)
* DB interface for PostgreSQL implemented with [sqlc](https://github.com/sqlc-dev/sqlc) - modified to improve test coverage 
* Integration testing with [Testcontainers](https://github.com/testcontainers/testcontainers-go)