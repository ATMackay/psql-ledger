package service

import (
	"context"
	"fmt"

	"github.com/ATMackay/psql-ledger/database"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func makePostgresDBClient(l *logrus.Entry, config Config) (database.DBClient, error) {

	d, err := makeClientSet(l, config)
	if err != nil {
		return nil, err
	}
	// check DB exists
	exists, err := d.CheckDatabaseExists(context.Background(), config.PostgresDB)
	if err != nil {
		return nil, fmt.Errorf("CheckDatabaseExists err: %v", err)
	}

	if !exists {
		l.Debugf("DB %v not found", config.PostgresDB)
		// TODO - attempt DB creation again..
		return nil, fmt.Errorf("DB %v does not exist", config.PostgresDB)
	} else {
		l.Debugf("found DB %v", config.PostgresDB)
	}

	if err := d.InitializeSchema(config.MigrationsPath); err != nil {
		return nil, fmt.Errorf("InitializeSchema failed: %v", err)
	}
	l.Debugf("migrated DB using schema path '%v'", config.MigrationsPath)
	return d, nil
}

type aggregatedClient struct {
	clients chan database.DBClient
}

func makeClientSet(l *logrus.Entry, config Config) (aggregatedClient, error) {
	n := config.MaxThreads
	clients := make(chan database.DBClient, n)
	a := aggregatedClient{clients: clients}
	for i := 0; i < n; i++ {
		// creates n new connections
		c, err := pq.NewConnector(fmt.Sprintf("host=%v port=%v user=%v password=%v dbname=%v sslmode=disable",
			config.PostgresHost,
			config.PostgresPort,
			config.PostgresUser,
			config.PostgresPassword,
			config.PostgresDB))
		if err != nil {
			return a, fmt.Errorf("NewConnector err: %v", err)
		}
		dbClient, err := database.NewPSQLClient(config.PostgresDB, c)
		if err != nil {
			return a, fmt.Errorf("NewPSQLClient err: %v", err)
		}
		l.WithField("client", i).Debug("new client")
		a.clients <- dbClient
	}
	return a, nil

}

func (a aggregatedClient) CheckDatabaseExists(ctx context.Context, dbName string) (bool, error) {
	cl := <-a.clients
	defer func() {
		a.clients <- cl
	}()
	return cl.CheckDatabaseExists(ctx, dbName)
}

func (a aggregatedClient) InitializeSchema(schemaPath string) error {
	cl := <-a.clients
	defer func() {
		a.clients <- cl
	}()
	return cl.InitializeSchema(schemaPath)
}

func (a aggregatedClient) DB() database.DB {
	cl := <-a.clients
	defer func() {
		a.clients <- cl
	}()
	return cl.DB()
}

func (a aggregatedClient) NewQuery() database.DBQuery {
	cl := <-a.clients
	defer func() {
		a.clients <- cl
	}()
	return cl.NewQuery()
}

func (a aggregatedClient) NewTransaction() (database.DBTX, error) {
	cl := <-a.clients
	defer func() {
		a.clients <- cl
	}()
	return cl.NewTransaction()
}

func (a aggregatedClient) NewQueryWithTx() (database.DBQuery, error) {
	cl := <-a.clients
	defer func() {
		a.clients <- cl
	}()
	return cl.NewQueryWithTx()
}
