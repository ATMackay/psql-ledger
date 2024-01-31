package database

import (
	"context"
	"database/sql"
)

type DB interface {
	InitializeSchema(schemaPath string) error
	Close() error
	Ping() error
	NewQuery() DBQuery
	NewTransaction() (DBQuery, error)
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
	WithTx(tx *sql.Tx) DBQuery
}
