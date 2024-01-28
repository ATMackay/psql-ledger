package database

import (
	"context"
	"database/sql"
)

// For testing purposes

var _ DB = (*MemoryDB)(nil)

type MemoryDB struct {
	client FakeClient
}

func NewMemDB() MemoryDB {
	return MemoryDB{FakeClient{}}
}

func (m MemoryDB) Close() error {
	return nil
}

func (m MemoryDB) QueryClient() DBQuery {
	return m.client
}

type FakeClient struct{}

func (f FakeClient) CreateAccount(ctx context.Context, arg CreateAccountParams) (Account, error) {
	return Account{}, nil
}

func (f FakeClient) CreateTransaction(ctx context.Context, arg CreateTransactionParams) (Transaction, error) {
	return Transaction{}, nil
}

func (f FakeClient) DeleteAccount(ctx context.Context, id int64) error {
	return nil
}

func (f FakeClient) GetTx(ctx context.Context, id int64) (Transaction, error) {
	return Transaction{}, nil
}

func (f FakeClient) GetUser(ctx context.Context, id int64) (Account, error) {
	return Account{}, nil
}

func (f FakeClient) GetUserByEmail(ctx context.Context, email sql.NullString) (Account, error) {
	return Account{}, nil
}

func (f FakeClient) GetUserByUsername(ctx context.Context, username string) (Account, error) {
	return Account{}, nil
}

func (f FakeClient) GetUsers(ctx context.Context) ([]Account, error) {
	return []Account{}, nil
}

func (f FakeClient) WithTx(tx *sql.Tx) *Queries {
	return &Queries{}
}

type FakeDBTx struct{}

func (f FakeDBTx) ExecContext(context.Context, string, ...any) (sql.Result, error) {
	return fakeResult{}, nil
}

func (f FakeDBTx) PrepareContext(context.Context, string) (*sql.Stmt, error) {
	return &sql.Stmt{}, nil
}

func (f FakeDBTx) QueryContext(context.Context, string, ...any) (*sql.Rows, error) {
	return &sql.Rows{}, nil
}

func (f FakeDBTx) QueryRowContext(context.Context, string, ...any) *sql.Row {
	return &sql.Row{}
}

type fakeResult struct{}

func (f fakeResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (f fakeResult) RowsAffected() (int64, error) {
	return 0, nil
}
