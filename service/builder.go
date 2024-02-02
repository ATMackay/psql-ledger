package service

import (
	"context"
	"fmt"

	"github.com/ATMackay/psql-ledger/database"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

const serviceName = "psqlledger"

// BuildService constructs a service with PostgreSQL DB client using the
// supplied configuration parameters.
func BuildService(cfg Config) (*Service, error) {

	config, defaultUsed := sanitizeConfig(cfg)

	l, err := NewLogger(Level(config.LogLevel), Format(config.LogFormat), config.LogToFile, serviceName)
	if err != nil {
		return nil, err
	}

	if defaultUsed {
		l.Warnf("no config parameters supplied: using default")
	}

	db, err := makePostgresDB(config)
	if err != nil {
		return nil, fmt.Errorf("could not make postgres DB: %v", err)
	}

	l.WithFields(logrus.Fields{
		"port":     config.Port,
		"loglevel": config.LogLevel,
		"DBHost":   config.PostgresHost,
		"DBPort":   config.PostgresPort,
		"DBUser":   config.PostgresUser,
		"DBName":   config.PostgresDB,
	}).Info("connected to postgresDB")

	return New(config.Port, l, db), nil
}

func makePostgresDB(config Config) (*database.PSQLClient, error) {

	c, err := pq.NewConnector(fmt.Sprintf("host=%v port=%v user=%v password=%v dbname=%v sslmode=disable",
		config.PostgresHost,
		config.PostgresPort,
		config.PostgresUser,
		config.PostgresPassword,
		config.PostgresDB))
	if err != nil {
		return nil, fmt.Errorf("NewConnector err: %v", err)
	}
	dbClient, err := database.NewPSQLClient(config.PostgresDB, c)
	if err != nil {
		return nil, fmt.Errorf("new PSQLClient err: %v", err)
	}

	// check DB exists
	exists, err := dbClient.CheckDatabaseExists(context.Background(), config.PostgresDB)
	if err != nil {
		return nil, fmt.Errorf("CheckDatabaseExists err: %v", err)
	}

	if !exists {
		// TODO - attempt DB creation again..
		return nil, fmt.Errorf("DB %v does not exist", config.PostgresDB)
	}

	if err := dbClient.InitializeSchema(config.MigrationsPath); err != nil {
		return nil, fmt.Errorf("InitializeSchema failed: %v", err)
	}
	return dbClient, nil
}

func New(port int, l *logrus.Entry, dbClient database.DBClient) *Service {
	s := &Service{
		logger:   l,
		dbClient: dbClient,
	}
	h := NewHTTPService(port, makeServiceAPIs(s), l)
	s.server = &h
	return s
}
