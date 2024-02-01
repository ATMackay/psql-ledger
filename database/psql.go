package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
)

var _ DBClient = (*PSQLClient)(nil)

type PSQLClient struct {
	dbName string
	db     *sql.DB
}

func NewPSQLClient(dbName string, c driver.Connector) (*PSQLClient, error) {

	// Open DB with sql
	db := sql.OpenDB(c)

	// Check connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PSQLClient{
		dbName: dbName,
		db:     db}, nil
}

// CheckDatabaseExists checks where a postgres instance has the specified dbName relation
func (p *PSQLClient) CheckDatabaseExists(ctx context.Context, dbName string) (bool, error) {
	// Use the 'pgx' driver-specific method to query for the database existence
	var exists bool
	if err := p.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

// InitializeSchema migrates up the connected postgres instance with the schema contained in the
// supplied migration directory.
func (p *PSQLClient) InitializeSchema(migrationDir string) error {

	driver, err := postgres.WithInstance(p.db, &postgres.Config{DatabaseName: p.dbName, MigrationsTable: migrationDir})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationDir),
		"postgres", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

// DB returns the underlying DB interface with Open and Close commands.
func (p *PSQLClient) DB() DB {
	return p.db
}

// NewQuery returns DBQuery interface
func (p *PSQLClient) NewQuery() DBQuery {
	return New(p.db)
}

// NewQueryWithTx creates a database transaction with query methods
func (p *PSQLClient) NewQueryWithTx() (DBQuery, error) {
	tx, err := p.NewTransaction()
	if err != nil {
		return nil, err
	}
	return New(tx), nil
}

func (p *PSQLClient) NewTransaction() (DBTX, error) {
	conn, err := p.db.Conn(context.Background())
	if err != nil {
		return nil, err
	}
	sqlTx, err := conn.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	return sqlTx, nil
}
