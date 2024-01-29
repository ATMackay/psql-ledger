package service

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/ATMackay/psql-ledger/database"
)

const (
	Status = "/status"
	Health = "/health"

	GetAccount           = "/account-by-index"
	GetAccountByEmail    = "/account-by-email"
	GetAccountByUsername = "/account-by-username"

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
	resp := StatusResponse{Message: "OK", Version: FullVersion, Service: serviceName}
	if err := RespondWithJSON(w, http.StatusOK, resp); err != nil {
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

type AccountParams struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Balance  int64  `json:"balance"`

	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
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

func validAccountParams(c AccountParams) error {
	// validate email input
	if c.Email != "" {
		if err := isValidString(c.Email, emailRegex); err != nil {
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

// AccountByIndex requests the account for supplied ID number
func (s *Service) AccountByIndex(w http.ResponseWriter, r *http.Request) {
	var c AccountParams
	if err := DecodeJSON(r.Body, &c); err != nil {
		RespondWithError(w, http.StatusBadRequest, err)
		return
	}

	if c.ID == 0 {
		RespondWithError(w, http.StatusBadRequest, fmt.Errorf("cannot supply account ID = 0"))
		return
	}

	// Execute Query against PSQL
	acc, err := s.db.QueryClient().GetUser(context.Background(), c.ID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err)
		return
	}

	if err := RespondWithJSON(w, http.StatusOK, &AccountParams{ID: acc.ID, Username: acc.Username, Email: acc.Email.String, CreatedAt: acc.CreatedAt.Time}); err != nil {
		s.logger.Error(err)
	}
}

// AccountByUsername requests the account for supplied ID number
func (s *Service) AccountByUsername(w http.ResponseWriter, r *http.Request) {
	var c AccountParams
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
	acc, err := s.db.QueryClient().GetUserByUsername(context.Background(), c.Username)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err)
		return
	}

	if err := RespondWithJSON(w, http.StatusOK, &AccountParams{ID: acc.ID, Username: acc.Username, Email: acc.Email.String, CreatedAt: acc.CreatedAt.Time}); err != nil {
		s.logger.Error(err)
	}
}

// AccountByUsername requests the account for supplied ID number
func (s *Service) AccountByEmail(w http.ResponseWriter, r *http.Request) {
	var c AccountParams
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
	acc, err := s.db.QueryClient().GetUserByEmail(context.Background(), sql.NullString{String: c.Email})
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err)
		return
	}

	if err := RespondWithJSON(w, http.StatusOK, &AccountParams{ID: acc.ID, Username: acc.Username, Email: acc.Email.String, CreatedAt: acc.CreatedAt.Time}); err != nil {
		s.logger.Error(err)
	}
}

// Write Requests

func (s *Service) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var c AccountParams
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
	acc, err := s.db.QueryClient().CreateAccount(context.Background(), database.CreateAccountParams{
		Email:    sql.NullString{String: c.Email},
		Username: c.Username,
		Balance:  0,
	})
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err)
		return
	}

	if err := RespondWithJSON(w, http.StatusOK, &AccountParams{ID: acc.ID, Username: acc.Username, Email: acc.Email.String, CreatedAt: acc.CreatedAt.Time}); err != nil {
		s.logger.Error(err)
	}

}

type TxParams struct {
	FromAccount int64 `json:"from_account"`
	ToAccount   int64 `json:"to_account"`
	Amount      int64 `json:"amount"`

	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

// Read Requests

func (s *Service) TxHistory(w http.ResponseWriter, r *http.Request) {
	// TODO
}

// Write Requests

func (s *Service) CreateTx(w http.ResponseWriter, r *http.Request) {
	var txParams TxParams
	if err := DecodeJSON(r.Body, &txParams); err != nil {
		RespondWithError(w, http.StatusBadRequest, err)
		return
	}

	// Validate data
	if txParams.Amount <= 0 {
		RespondWithError(w, http.StatusBadRequest, fmt.Errorf("cannot send negative amount '%v'", txParams.Amount))
	}

	// Check to and from account exist

	// Execute Query against PSQL
	tx, err := s.db.QueryClient().CreateTransaction(context.Background(), database.CreateTransactionParams{
		FromAccount: sql.NullInt64{Int64: txParams.FromAccount},
		ToAccount:   sql.NullInt64{Int64: txParams.FromAccount},
		Amount:      sql.NullInt64{Int64: txParams.Amount},
	})
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err)
		return
	}

	if err := RespondWithJSON(w, http.StatusOK, &TxParams{ID: tx.ID, CreatedAt: tx.CreatedAt.Time, FromAccount: tx.FromAccount.Int64, ToAccount: tx.ToAccount.Int64}); err != nil {
		s.logger.Error(err)
	}
}
