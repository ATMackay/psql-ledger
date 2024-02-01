package database

import (
	"context"
	"database/sql"
)

type DBClient interface {
	InitializeSchema(schemaPath string) error
	CheckDatabaseExists(ctx context.Context, dbName string) (bool, error)
	DB() DB
	NewQuery() DBQuery
	NewTransaction() (DBTX, error)
	NewQueryWithTx() (DBQuery, error)
}

type DB interface {
	Close() error
	Ping() error
}

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
