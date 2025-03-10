package service

import (
	"fmt"
	"log/slog"

	"github.com/ATMackay/psql-ledger/database"
)

const ServiceName = "psql-ledger"

// BuildService constructs a service with PostgreSQL DB client using the
// supplied configuration parameters.
func BuildService(cfg Config) (*Service, error) {

	config, defaultUsed := sanitizeConfig(cfg)

	if err := InitLogging(config.LogLevel, config.LogFormat, config.LogToFile); err != nil {
		return nil, err
	}

	if defaultUsed {
		slog.Warn("no config parameters supplied: using default")
	}

	db, err := makePostgresDBClient(config)
	if err != nil {
		return nil, fmt.Errorf("could not make postgres DB: %v", err)
	}

	slog.Info("connected to postgresDB",
		"DBHost", config.PostgresHost,
		"DBPort", config.PostgresPort,
		"DBUser", config.PostgresUser,
		"DBName", config.PostgresDB)

	return New(config.Port, config.MaxThreads, db), nil
}

func New(port, threads int, dbClient database.DBClient) *Service {
	s := &Service{
		dbClient: dbClient,
	}
	h := NewHTTPService(port, makeServiceAPIs(dbClient))
	s.server = &h
	return s
}
