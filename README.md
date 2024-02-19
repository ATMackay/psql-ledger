# A simple transaction ledger implemented in Go and PostgreSQL

Web backend tooling for creating user accounts and recording transactions between users.

## Components

* Go HTTP microservice built with [httprouter](https://github.com/julienschmidt/httprouter)
* DB interface for PostgreSQL implemented with [sqlc](https://github.com/sqlc-dev/sqlc) - modified to improve test coverage 
* Integration testing with [Testcontainers](https://github.com/testcontainers/testcontainers-go)

## Getting Started

Start the postgres server
```
~/go/src/github.com/ATMackay/psql-ledger$ make postgresup
```

Create 'bank' database
```
~/go/src/github.com/ATMackay/psql-ledger$ make createdb
```

Start service
```
~/go/src/github.com/ATMackay/psql-ledger$ make run
```

Use a new terminal to interact with the application. Healthcheck the stack (an empty failures list indicates that the service is healthy and ready to take requests).
```
~$ curl localhost:8080/health
{"version":"v0.1.0-17379d11","service":"psql-ledger","failures":[]}
```

Create an account
```
~$ curl -X PUT -H "Content-Type: application/json" -d '{"username": "exampleuser", "email": {"String": "user@example.com", "Valid": true} }' http://localhost:8080/create-account
{"id":1,"username":"exampleuser","balance":0,"email":{"String":"user@example.com","Valid":true},"created_at":{"Time":"2024-02-01T13:12:27.782459Z","Valid":true}}
```

Fetch create account details (by-index)
```
~$ curl -X POST -H "Content-Type: application/json" -d '{"id":1}' http://localhost:8080/account-by-index
{"id":1,"username":"exampleuser","balance":0,"email":{"String":"user@example.com","Valid":true},"created_at":{"Time":"2024-02-01T13:12:27.782459Z","Valid":true}}
```