package service

import (
	"fmt"

	"github.com/ATMackay/psql-ledger/database"
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

	db, err := makePostgresDBClient(l, config)
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

	return New(config.Port, config.MaxThreads, l, db), nil
}

func New(port, threads int, l *logrus.Entry, dbClient database.DBClient) *Service {
	s := &Service{
		logger:   l,
		dbClient: dbClient,
	}
	h := NewHTTPService(port, makeServiceAPIs(s), l)
	s.server = &h
	return s
}
