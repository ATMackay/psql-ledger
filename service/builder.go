package service

import (
	"github.com/ATMackay/psql-ledger/database"
	"github.com/sirupsen/logrus"
)

const serviceName = "psqlledger"

func BuildService(cfg Config) (*Service, error) {

	config, defaultUsed := sanitizeConfig(cfg)

	l, err := NewLogger(Level(config.LogFormat), Format(config.LogFormat), config.LogToFile, serviceName)
	if err != nil {
		return nil, err
	}

	if defaultUsed {
		l.Infof("no config parameters supplied: using default")
	}

	db, err := database.NewPSQLClient("")
	if err != nil {
		return nil, err
	}

	return New(cfg.Port, l, db), nil
}

func New(port int, l *logrus.Entry, db database.DB) *Service {
	s := &Service{
		logger: l,
		db:     db,
	}
	server := NewHTTPService(8080, makeServiceAPIs(s), l)
	s.server = server
	return s
}

func makeServiceAPIs(s *Service) *API {
	return MakeAPI([]EndPoint{
		EndPoint{
			Path:       "/tx",
			Handler:    s.Tx,
			MethodType: "POST",
		},
		EndPoint{
			Path:       "/create-account",
			Handler:    s.CreateAccount,
			MethodType: "POST",
		},
	})
}
