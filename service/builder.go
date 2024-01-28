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
		l.Infof("no config parameters supplied: using default")
	}

	connString := fmt.Sprintf("user=%v password=%v dbname=%v sslmode=disable", config.PostgresUser, config.PostgresPassword, config.PostgresDB)
	db, err := database.NewPSQLClient(connString)
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
	server := NewHTTPService(port, makeServiceAPIs(s), l)
	s.server = server
	return s
}

func makeServiceAPIs(s *Service) *API {
	return MakeAPI([]EndPoint{
		EndPoint{
			Path:       "/create-tx",
			Handler:    s.CreateTx,
			MethodType: "POST",
		},
		EndPoint{
			Path:       "/create-account",
			Handler:    s.CreateAccount,
			MethodType: "POST",
		},
	})
}
