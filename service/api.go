package service

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/ATMackay/psql-ledger/database"
)

const (
	Status = "/status"
	Health = "/health"

	Accounts             = "/accounts"
	GetAccount           = "/account-by-index"
	GetAccountByEmail    = "/account-by-email"
	GetAccountByUsername = "/account-by-username"

	GetAccountTransactions = "/account-txs"

	GetTransactionByIndex = "/tx"

	CreateTx      = "/create-tx"
	CreateAccount = "/create-account"
)

func makeServiceAPIs(s *Service) *API {
	return MakeAPI([]EndPoint{
		EndPoint{
			Path:       Status,
			Handler:    s.Status,
			MethodType: http.MethodGet,
		},
		EndPoint{
			Path:       Health,
			Handler:    s.Health,
			MethodType: http.MethodGet,
		},
		EndPoint{
			Path:       Accounts,
			Handler:    s.Accounts,
			MethodType: http.MethodGet,
		},
		EndPoint{
			Path:       GetAccount,
			Handler:    s.AccountByIndex,
			MethodType: http.MethodGet,
		},
		EndPoint{
			Path:       GetAccountByEmail,
			Handler:    s.AccountByEmail,
			MethodType: http.MethodGet,
		},
		EndPoint{
			Path:       GetAccountByUsername,
			Handler:    s.AccountByUsername,
			MethodType: http.MethodGet,
		},
		EndPoint{
			Path:       GetAccountTransactions,
			Handler:    s.TxHistory,
			MethodType: http.MethodGet,
		},
		EndPoint{
			Path:       GetTransactionByIndex,
			Handler:    s.TransactionByIndex,
			MethodType: http.MethodGet,
		},
		EndPoint{
			Path:       CreateTx,
			Handler:    s.CreateTx,
			MethodType: http.MethodPost,
		},
		EndPoint{
			Path:       CreateAccount,
			Handler:    s.CreateAccount,
			MethodType: http.MethodPost,
		},
	})
}

// StatusResponse contains status response fields.
type StatusResponse struct {
	Message string `json:"message,omitempty"`
	Version string `json:"version,omitempty"`
	Service string `json:"service,omitempty"`
}

func (s *Service) Status(w http.ResponseWriter, r *http.Request) {
	if err := RespondWithJSON(w, http.StatusOK, &StatusResponse{Message: "OK", Version: FullVersion, Service: serviceName}); err != nil {
		s.logger.Error(err)
	}
}

// HealthResponse contains status response fields.
type HealthResponse struct {
	Version  string   `json:"version,omitempty"`
	Service  string   `json:"service,omitempty"`
	Failures []string `json:"failures"`
}

func (s *Service) Health(w http.ResponseWriter, r *http.Request) {
	health := &HealthResponse{
		Service: serviceName,
		Version: FullVersion,
	}
	var failures = []string{}
	var httpCode = http.StatusOK
	if err := s.db.Ping(); err != nil {
		failures = append(failures, fmt.Sprintf("DB: %v", err))
		httpCode = http.StatusServiceUnavailable
	}

	health.Failures = failures

	if err := RespondWithJSON(w, httpCode, health); err != nil {
		s.logger.Error(err)
	}
}

const (
	// This regex defines the regular expression for simple email formats
	//
	// e.g. alex@emailprovider.com -VALID
	//
	//      dhd$@xyz.com - INVALID
	//
	emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	// This regex defines the regular expression for simple email formats
	//
	// e.g. user105 - VALID
	//
	usernameRegex = "^[a-zA-Z0-9]+$"
)

func validAccountParams(c database.Account) error {
	// validate email input
	if c.Email.String != "" {
		if err := isValidString(c.Email.String, emailRegex); err != nil {
			return fmt.Errorf("invalid user email: %v", err)
		}
	}
	if c.Username != "" {
		if err := isValidString(c.Username, usernameRegex); err != nil {
			return fmt.Errorf("invalid user email: %v", err)
		}
	}
	return nil
}

func isValidString(input string, regex string) error {
	re, err := regexp.Compile(regex)
	if err != nil {
		return err
	}

	if !re.MatchString(input) {
		return fmt.Errorf(" '%v' failed to match expression '%v'", input, regex)
	}

	return nil
}

// Read Requests

// Accounts requests the full list if accounts stored in the DB - TODO paginate this request
func (s *Service) Accounts(w http.ResponseWriter, r *http.Request) {
	// Execute Query against PSQL
	acc, err := s.db.NewQuery().GetUsers(context.Background())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err)
		return
	}

	if err := RespondWithJSON(w, http.StatusOK, acc); err != nil {
		s.logger.Error(err)
	}
}

// AccountByIndex requests the account for supplied ID number
func (s *Service) AccountByIndex(w http.ResponseWriter, r *http.Request) {
	var c database.Account
	if err := DecodeJSON(r.Body, &c); err != nil {
		RespondWithError(w, http.StatusBadRequest, err)
		return
	}

	if c.ID == 0 {
		RespondWithError(w, http.StatusBadRequest, fmt.Errorf("cannot supply account ID = 0"))
		return
	}

	// Execute Query against PSQL
	acc, err := s.db.NewQuery().GetUser(context.Background(), c.ID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err)
		return
	}

	if err := RespondWithJSON(w, http.StatusOK, acc); err != nil {
		s.logger.Error(err)
	}
}

// AccountByUsername requests the account for supplied ID number
func (s *Service) AccountByUsername(w http.ResponseWriter, r *http.Request) {
	var c database.Account
	if err := DecodeJSON(r.Body, &c); err != nil {
		RespondWithError(w, http.StatusBadRequest, err)
		return
	}

	// validate inputs
	if err := validAccountParams(c); err != nil {
		RespondWithError(w, http.StatusBadRequest, err)
		return
	}

	// Execute Query against PSQL
	acc, err := s.db.NewQuery().GetUserByUsername(context.Background(), c.Username)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err)
		return
	}

	if err := RespondWithJSON(w, http.StatusOK, acc); err != nil {
		s.logger.Error(err)
	}
}

// AccountByUsername requests the account for supplied ID number
func (s *Service) AccountByEmail(w http.ResponseWriter, r *http.Request) {
	var c database.Account
	if err := DecodeJSON(r.Body, &c); err != nil {
		RespondWithError(w, http.StatusBadRequest, err)
		return
	}
	// validate inputs
	if err := validAccountParams(c); err != nil {
		RespondWithError(w, http.StatusBadRequest, err)
		return
	}
	// Execute Query against PSQL
	acc, err := s.db.NewQuery().GetUserByEmail(context.Background(), c.Email)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err)
		return
	}

	if err := RespondWithJSON(w, http.StatusOK, acc); err != nil {
		s.logger.Error(err)
	}
}

// Write Requests

func (s *Service) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var c database.Account
	if err := DecodeJSON(r.Body, &c); err != nil {
		RespondWithError(w, http.StatusBadRequest, err)
		return
	}

	// validate inputs
	if err := validAccountParams(c); err != nil {
		RespondWithError(w, http.StatusBadRequest, err)
		return
	}

	// Execute Query against PSQL
	acc, err := s.db.NewQuery().CreateAccount(context.Background(), database.CreateAccountParams{
		Email:    c.Email,
		Username: c.Username,
		Balance:  0,
	})
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err)
		return
	}

	if err := RespondWithJSON(w, http.StatusOK, acc); err != nil {
		s.logger.Error(err)
	}

}

// Read Requests

// TransactionByIndex requests the transaction for supplied ID number
func (s *Service) TransactionByIndex(w http.ResponseWriter, r *http.Request) {
	var txParams database.Transaction
	if err := DecodeJSON(r.Body, &txParams); err != nil {
		RespondWithError(w, http.StatusBadRequest, err)
		return
	}

	if txParams.ID == 0 {
		RespondWithError(w, http.StatusBadRequest, fmt.Errorf("cannot supply account ID = 0"))
		return
	}

	// Execute Query against PSQL
	tx, err := s.db.NewQuery().GetTx(context.Background(), txParams.ID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err)
		return
	}

	if err := RespondWithJSON(w, http.StatusOK, tx); err != nil {
		s.logger.Error(err)
	}
}

func (s *Service) TxHistory(w http.ResponseWriter, r *http.Request) {
	var c database.Account
	if err := DecodeJSON(r.Body, &c); err != nil {
		RespondWithError(w, http.StatusBadRequest, err)
		return
	}

	if c.ID == 0 {
		RespondWithError(w, http.StatusBadRequest, fmt.Errorf("cannot supply account ID = 0"))
		return
	}

	// Execute Query against PSQL
	txs, err := s.db.NewQuery().GetUserTransactions(context.Background())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err)
		return
	}

	// TODO - this is not efficient. Need to create the correct query as opposed to dumping the entire tx set
	var filteredTx []*database.GetUserTransactionsRow
	for i := range txs {

		tx := txs[i]

		if tx.FromAccountID.Int64 == c.ID || tx.ToAccountID.Int64 == c.ID {
			filteredTx = append(filteredTx, &tx)
		}
	}

	if err := RespondWithJSON(w, http.StatusOK, filteredTx); err != nil {
		s.logger.Error(err)
	}
}

// Write Requests

func (s *Service) CreateTx(w http.ResponseWriter, r *http.Request) {
	var txParams database.Transaction
	if err := DecodeJSON(r.Body, &txParams); err != nil {
		RespondWithError(w, http.StatusBadRequest, err)
		return
	}

	// Validate amount
	if txParams.Amount.Int64 <= 0 {
		RespondWithError(w, http.StatusBadRequest, fmt.Errorf("cannot send negative amount '%v'", txParams.Amount))
		return
	}

	if txParams.FromAccount.Int64 == txParams.ToAccount.Int64 {
		RespondWithError(w, http.StatusBadRequest, fmt.Errorf("to and from account cannot match"))
		return
	}

	// Check to and from account exist
	if _, err := s.db.NewQuery().GetUser(context.Background(), txParams.FromAccount.Int64); err != nil {
		RespondWithError(w, http.StatusBadRequest, err)
		return
	}
	if _, err := s.db.NewQuery().GetUser(context.Background(), txParams.ToAccount.Int64); err != nil {
		RespondWithError(w, http.StatusBadRequest, err)
		return
	}

	// Execute Query against PSQL
	tx, err := s.db.NewQuery().CreateTransaction(context.Background(), database.CreateTransactionParams{
		FromAccount: txParams.FromAccount,
		ToAccount:   txParams.ToAccount,
		Amount:      txParams.Amount,
	})
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err)
		return
	}
	// TODO, update user balance - requires using DB Tx (with rollback)

	if err := RespondWithJSON(w, http.StatusOK, tx); err != nil {
		s.logger.Error(err)
	}
}
