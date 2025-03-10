package service

import (
	"log/slog"
	"os"

	"github.com/ATMackay/psql-ledger/database"
)

// Service represents the main psqllgedger service body with
// HTTP interface and DB connection.
type Service struct {
	dbClient database.DBClient
	server   *HTTPService
}

func (s *Service) Start() {
	slog.Info("starting service", "version", Version, "commitDate", CommitDate, "buildDate", BuildDate, "gitCommitSha", GitCommitHash)
	s.server.Start()
}

func (s *Service) Stop(sig os.Signal) {
	slog.Info("stopping service", "signal", sig)

	if err := s.dbClient.DB().Close(); err != nil {
		slog.Error("error closing db", "error", err)
	}

	if err := s.server.Stop(); err != nil {
		slog.Error("error stopping server")
	}
}

func (s *Service) Server() *HTTPService {
	return s.server
}
