package database

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
)

// To be used for testing both in and outside of this package

var _ DBClient = (*MemDBClient)(nil)
var _ DB = (*MemDB)(nil)
var _ DBQuery = (*MemDBQuery)(nil)

type MemDBClient struct {
	q  MemDBQuery
	tx FakeDBTx
}

func NewMemoryDBClient() MemDBClient {
	db := newMemDB()
	return MemDBClient{tx: FakeDBTx{db: &db}, q: MemDBQuery{db: &db}}
}

func (m MemDBClient) InitializeSchema(migrationDir string) error {
	return nil
}

func (m MemDBClient) NewQuery() DBQuery {
	return m.q
}

func (m MemDBClient) DB() DB {
	return m.q.db
}

func (m MemDBClient) NewTransaction() (DBTX, error) {
	return m.tx, nil
}

func (m MemDBClient) NewQueryWithTx() (DBQuery, error) {
	return m.q, nil
}

func (m MemDBClient) CheckDatabaseExists(ctx context.Context, dbName string) (bool, error) {
	return true, nil
}

func newMemDB() MemDB {
	a := make(map[int64]Account)
	t := make(map[int64]Transaction)
	return MemDB{accounts: a, transactions: t}
}

type MemDB struct {
	accounts     map[int64]Account
	transactions map[int64]Transaction
}

func (m MemDB) Ping() error {
	return nil
}

func (m MemDB) Close() error {
	return nil
}

type MemDBQuery struct {
	db *MemDB
}

func (f MemDBQuery) CreateAccount(ctx context.Context, arg CreateAccountParams) (Account, error) {
	l := len(f.db.accounts)
	index := int64(l + 1)
	a := Account{ID: index, Balance: 0, Username: arg.Username, Email: arg.Email}
	f.db.accounts[index] = a
	return a, nil
}

func (f MemDBQuery) CreateTransaction(ctx context.Context, arg CreateTransactionParams) (Transaction, error) {
	l := len(f.db.transactions)
	index := int64(l + 1)
	tx := Transaction{ID: index, FromAccount: arg.FromAccount, ToAccount: arg.ToAccount, Amount: arg.Amount}
	f.db.transactions[index] = tx
	return tx, nil
}

func (f MemDBQuery) DeleteAccount(ctx context.Context, id int64) error {
	return nil
}

func (f MemDBQuery) GetTx(ctx context.Context, id int64) (Transaction, error) {
	tx, ok := f.db.transactions[id]
	if !ok {
		return Transaction{}, fmt.Errorf("not found")
	}
	return tx, nil
}

func (f MemDBQuery) GetUser(ctx context.Context, id int64) (Account, error) {
	a, ok := f.db.accounts[id]
	if !ok {
		return Account{}, fmt.Errorf("not found")
	}
	return a, nil
}

func (f MemDBQuery) GetUserByEmail(ctx context.Context, email sql.NullString) (Account, error) {
	var a Account
	for i := range f.db.accounts {
		if f.db.accounts[i].Email.String == email.String {
			a = f.db.accounts[i]
		}
	}
	return a, nil
}

func (f MemDBQuery) GetUserByUsername(ctx context.Context, username string) (Account, error) {
	var a Account
	for i := range f.db.accounts {
		if f.db.accounts[i].Username == username {
			a = f.db.accounts[i]
		}
	}
	return a, nil
}

func (f MemDBQuery) GetUsers(ctx context.Context) ([]Account, error) {
	var a []Account
	for i := range f.db.accounts {

		a = append(a, f.db.accounts[i])

	}
	sort.Slice(a, func(i, j int) bool { return a[i].ID < a[j].ID })
	return a, nil
}

func (f MemDBQuery) GetUserTransactions(ctx context.Context) ([]GetUserTransactionsRow, error) {
	var txs []GetUserTransactionsRow
	for i := range f.db.transactions {
		tx := f.db.transactions[i]
		txs = append(txs, GetUserTransactionsRow{TransactionID: tx.ID, FromAccountID: tx.FromAccount, ToAccountID: tx.ToAccount, Amount: tx.Amount})
	}
	sort.Slice(txs, func(i, j int) bool { return txs[i].TransactionID < txs[j].TransactionID })
	return txs, nil
}

func (f MemDBQuery) WithTx(tx DBTX) DBQuery {
	return f
}

type FakeDBTx struct {
	db *MemDB
}

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
