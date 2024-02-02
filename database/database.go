package database

import (
	"context"
	"database/sql"
)

// DBClient represents a database client.
type DBClient interface {
	CheckDatabaseExists(ctx context.Context, dbName string) (bool, error)
	InitializeSchema(schemaPath string) error

	DB() DB
	NewQuery() DBQuery
	NewTransaction() (DBTX, error)
	NewQueryWithTx() (DBQuery, error)
}

// DB represents basic database operations.
type DB interface {
	Close() error
	Ping() error
}

// DBQuery is an interface for executing queries on the database.
type DBQuery interface {
	CreateAccount(ctx context.Context, arg CreateAccountParams) (Account, error)
	CreateTransaction(ctx context.Context, arg CreateTransactionParams) (Transaction, error)
	DeleteAccount(ctx context.Context, id int64) error
	GetTx(ctx context.Context, id int64) (Transaction, error)
	GetUser(ctx context.Context, id int64) (Account, error)
	GetUserByEmail(ctx context.Context, email sql.NullString) (Account, error)
	GetUserByUsername(ctx context.Context, username string) (Account, error)
	GetUsers(ctx context.Context) ([]Account, error)
	GetUserTransactions(ctx context.Context) ([]GetUserTransactionsRow, error)
	WithTx(tx DBTX) DBQuery
}
