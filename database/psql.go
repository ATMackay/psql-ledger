package database

import (
	"database/sql"

	"github.com/lib/pq"
)

var _ DB = (*PSQLClient)(nil)

type PSQLClient struct {
	name        string
	db          *sql.DB
	queryClient DBQuery // wrapper for read/write queries to PSQL
	//connector *pq.Connector
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
	return &PSQLClient{name: connString, db: db, queryClient: New(db)}, nil
}

func (p *PSQLClient) Close() error {
	return p.db.Close()
}

func (p *PSQLClient) QueryClient() DBQuery {
	return p.queryClient
}
