package database

import (
	"context"
	"database/sql"
)

type DB interface {
	Close() error
	Ping() error
	QueryClient() DBQuery
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
	WithTx(tx *sql.Tx) DBQuery
}
