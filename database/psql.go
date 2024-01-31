package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/lib/pq"
)

var _ DB = (*PSQLClient)(nil)

type PSQLClient struct {
	host     string
	port     int
	user     string
	password string
	dbName   string
	db       *sql.DB
}

func NewPSQLClient(host string, port int, user, password, dbName string) (*PSQLClient, error) {

	connString := fmt.Sprintf("host=%v port=%v user=%v password=%v dbname=%v sslmode=disable",
		host,
		port,
		user,
		password,
		dbName)

	c, err := pq.NewConnector(connString)
	if err != nil {
		return nil, err
	}

	db := sql.OpenDB(c)

	// Check connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PSQLClient{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		dbName:   dbName,
		db:       db}, nil
}

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

func (p *PSQLClient) Ping() error {
	return p.db.Ping()
}

func (p *PSQLClient) Close() error {
	return p.db.Close()
}

func (p *PSQLClient) NewQuery() DBQuery {
	return New(p.db)
}

func (p *PSQLClient) NewTransaction() (DBQuery, error) {
	conn, err := p.db.Conn(context.Background())
	if err != nil {
		return nil, err
	}
	sqlTx, err := conn.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	return New(sqlTx), nil
}
