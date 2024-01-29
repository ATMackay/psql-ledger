package database

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

var _ DB = (*PSQLClient)(nil)

type PSQLClient struct {
	name string
	db   *sql.DB
}

func NewPSQLClient(connString string) (*PSQLClient, error) {
	c, err := pq.NewConnector(connString)
	if err != nil {
		return nil, err
	}

	db := sql.OpenDB(c)

	// Check connection
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &PSQLClient{name: connString, db: db}, nil
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
