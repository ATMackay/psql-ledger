package service

import (
	"fmt"

	"github.com/ATMackay/psql-ledger/database"
	"github.com/sirupsen/logrus"
)

const serviceName = "psqlledger"

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
		return nil, err
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

	db, err := database.NewPSQLClient(config.PostgresHost,
		config.PostgresPort,
		config.PostgresUser,
		config.PostgresPassword,
		config.PostgresDB)
	if err != nil {
		return nil, err
	}

	if err := db.InitializeSchema(config.MigrationsPath); err != nil {
		return nil, fmt.Errorf("DB migration up failed: %v", err)
	}
	return db, nil
}

func New(port int, l *logrus.Entry, db database.DB) *Service {
	s := &Service{
		logger: l,
		db:     db,
	}
	h := NewHTTPService(port, makeServiceAPIs(s), l)
	s.server = &h
	return s
}
