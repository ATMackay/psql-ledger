package service

import (
	"os"

	"github.com/ATMackay/psql-ledger/database"
	"github.com/sirupsen/logrus"
)

var versionFields = logrus.Fields{"buildDate": Date, "gitCommitSha": GitCommitHash}

// Service represents the main psqllgedger service boddy with
// HTTP interface and DB connection.
type Service struct {
	logger   *logrus.Entry
	dbClient database.DBClient
	server   *HTTPService
}

func (s *Service) Start() {
	s.logger.WithFields(versionFields).Infof("starting %v service", serviceName)
	s.server.Start()
}

func (s *Service) Stop(sig os.Signal) {
	s.logger.WithFields(logrus.Fields{"signal": sig}).Infof("stopping %v service", serviceName)

	if err := s.dbClient.DB().Close(); err != nil {
		s.logger.WithFields(logrus.Fields{"error": err}).Error("error closing db")
	}

	if err := s.server.Stop(); err != nil {
		s.logger.WithFields(logrus.Fields{"error": err}).Error("error stopping server")
	}
}

func (s *Service) Server() *HTTPService {
	return s.server
}
