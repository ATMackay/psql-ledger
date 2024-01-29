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

	connString := fmt.Sprintf("host=%v port=%v user=%v password=%v dbname=%v sslmode=disable",
		config.PostgresHost,
		config.PostgresPort,
		config.PostgresUser,
		config.PostgresPassword,
		config.PostgresDB)

	db, err := database.NewPSQLClient(connString)
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
	}).Info("creating psqlledger")

	return New(config.Port, l, db), nil
}

func New(port int, l *logrus.Entry, db database.DB) *Service {
	s := &Service{
		logger: l,
		db:     db,
	}
	server := NewHTTPService(port, makeServiceAPIs(s), l)
	s.server = server
	return s
}
