# Integration testing with Testcontainers

[Testcontainers](https://github.com/testcontainers/testcontainers-go) is a package that automates the creation and cleanup of container-based dependencies for integration tests. Here we can create a stack with postgreSQL and our `psqlledger` HTTP service, then execute test scenarios using the Go test framework.