package service

import (
	"os"

	"github.com/ATMackay/psql-ledger/database"
	"github.com/sirupsen/logrus"
)

var versionFields = logrus.Fields{"buildDate": Date, "gitCommitSha": GitCommitHash}

type Service struct {
	logger *logrus.Entry
	db     database.DB
	server *HTTPService
}

func (s *Service) Start() {
	s.logger.WithFields(versionFields).Infof("starting %v service", serviceName)
	s.server.Start()
}

func (s *Service) Stop(sig os.Signal) {
	s.logger.WithFields(logrus.Fields{"signal": sig}).Infof("stopping %v service", serviceName)

	if err := s.db.Close(); err != nil {
		s.logger.WithFields(logrus.Fields{"error": err}).Error("error closing db")
	}

	if err := s.server.Stop(); err != nil {
		s.logger.WithFields(logrus.Fields{"error": err}).Error("error stopping server")
	}
}

func (s *Service) Server() *HTTPService {
	return s.server
}
