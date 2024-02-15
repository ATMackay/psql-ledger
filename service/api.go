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
			Handler:    s.Status(),
			MethodType: http.MethodGet,
		},
		EndPoint{
			Path:       Health,
			Handler:    s.Health(),
			MethodType: http.MethodGet,
		},
		EndPoint{
			Path:       Accounts,
			Handler:    s.Accounts(),
			MethodType: http.MethodGet,
		},
		EndPoint{
			Path:       GetAccount,
			Handler:    s.AccountByIndex(),
			MethodType: http.MethodPost,
		},
		EndPoint{
			Path:       GetAccountByEmail,
			Handler:    s.AccountByEmail(),
			MethodType: http.MethodPost,
		},
		EndPoint{
			Path:       GetAccountByUsername,
			Handler:    s.AccountByUsername(),
			MethodType: http.MethodPost,
		},
		EndPoint{
			Path:       GetAccountTransactions,
			Handler:    s.TxHistory(),
			MethodType: http.MethodPost,
		},
		EndPoint{
			Path:       GetTransactionByIndex,
			Handler:    s.TransactionByIndex(),
			MethodType: http.MethodPost,
		},
		EndPoint{
			Path:       CreateTx,
			Handler:    s.CreateTx(),
			MethodType: http.MethodPut,
		},
		EndPoint{
			Path:       CreateAccount,
			Handler:    s.CreateAccount(),
			MethodType: http.MethodPut,
		},
	})
}

// HandleAsync wraps the request handler in a go-routine spawner, limited to the number of threads
// represented by the items in the threadpool channel.
// Increasing server throughput at the cost of losing deterministic execution.
//
// NOTE, this should only be used with underling dbClient of type aggregatedClient
// so that the number of active threads can be controlled.
func HandleAsync(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respWriter, req := &responseRecorder{ResponseWriter: w}, r.Clone(context.Background())
		go func() {
			h(respWriter, req) // create go-routine to handle request - TODO implement and test
		}()
	}
}

// StatusResponse contains status response fields.
type StatusResponse struct {
	Message string `json:"message,omitempty"`
	Version string `json:"version,omitempty"`
	Service string `json:"service,omitempty"`
}

// Status implements the status request endpoint. Always returns OK.
func (s *Service) Status() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := RespondWithJSON(w, http.StatusOK, &StatusResponse{Message: "OK", Version: FullVersion, Service: serviceName}); err != nil {
			s.logger.Error(err)
		}
	}

}

// HealthResponse contains status response fields.
type HealthResponse struct {
	Version  string   `json:"version,omitempty"`
	Service  string   `json:"service,omitempty"`
	Failures []string `json:"failures"`
}

// Health pings the connected DB instance.
func (s *Service) Health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		health := &HealthResponse{
			Service: serviceName,
			Version: FullVersion,
		}
		var failures = []string{}
		var httpCode = http.StatusOK
		if err := s.dbClient.DB().Ping(); err != nil {
			failures = append(failures, fmt.Sprintf("DB: %v", err))
			httpCode = http.StatusServiceUnavailable
		}

		health.Failures = failures

		if err := RespondWithJSON(w, httpCode, health); err != nil {
			s.logger.Error(err)
		}
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
	// This regex defines the regular expression for simple username formats
	//
	// e.g. user105 - VALID
	//      u£er101 - INVALID
	//
	usernameRegex = "^[a-zA-Z0-9]+$"
)

func validAccountParams(c database.CreateAccountParams) error {
	// validate email input
	if c.Email.String != "" {
		if err := isValidString(c.Email.String, emailRegex); err != nil {
			return fmt.Errorf("invalid user email: %v", err)
		}
	}
	if c.Username != "" {
		if err := isValidString(c.Username, usernameRegex); err != nil {
			return fmt.Errorf("invalid username: %v", err)
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

// GET Requests

// Accounts requests the full list if accounts stored in the DB - TODO paginate this request
func (s *Service) Accounts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Execute Query against PSQL
		acc, err := s.dbClient.NewQuery().GetUsers(context.Background())
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, err)
			return
		}

		if err := RespondWithJSON(w, http.StatusOK, acc); err != nil {
			s.logger.Error(err)
		}

	}

}

// AccountByIndex requests the account for supplied ID number
func (s *Service) AccountByIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var c database.Account
		if err := DecodeJSON(r.Body, &c); err != nil {
			RespondWithError(w, http.StatusBadRequest, err)
			return
		}

		if c.ID == 0 {
			err := fmt.Errorf("cannot supply account ID = 0")
			RespondWithError(w, http.StatusBadRequest, err)
			return
		}

		// Execute Query against PSQL
		acc, err := s.dbClient.NewQuery().GetUser(context.Background(), c.ID)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, err)
			return
		}

		if err := RespondWithJSON(w, http.StatusOK, acc); err != nil {
			s.logger.Error(err)
		}

	}

}

// AccountByUsername requests the account for supplied ID number
func (s *Service) AccountByUsername() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var c database.Account
		if err := DecodeJSON(r.Body, &c); err != nil {
			RespondWithError(w, http.StatusBadRequest, err)
			return
		}

		// validate inputs
		if err := validAccountParams(database.CreateAccountParams{Username: c.Username}); err != nil {
			RespondWithError(w, http.StatusBadRequest, err)
			return
		}

		// Execute Query against PSQL
		acc, err := s.dbClient.NewQuery().GetUserByUsername(context.Background(), c.Username)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, err)
			return
		}

		if err := RespondWithJSON(w, http.StatusOK, acc); err != nil {
			s.logger.Error(err)
		}

	}

}

// AccountByUsername requests the account for supplied ID number
func (s *Service) AccountByEmail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var c database.Account
		if err := DecodeJSON(r.Body, &c); err != nil {
			RespondWithError(w, http.StatusBadRequest, err)
			return
		}
		// validate inputs
		if err := validAccountParams(database.CreateAccountParams{Email: c.Email}); err != nil {
			RespondWithError(w, http.StatusBadRequest, err)
			return
		}
		// Execute Query against PSQL
		acc, err := s.dbClient.NewQuery().GetUserByEmail(context.Background(), c.Email)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, err)
			return
		}

		if err := RespondWithJSON(w, http.StatusOK, acc); err != nil {
			s.logger.Error(err)
		}
	}

}

// TransactionByIndex requests the transaction for supplied ID number - TODO paginate these requests
func (s *Service) TransactionByIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		tx, err := s.dbClient.NewQuery().GetTx(context.Background(), txParams.ID)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, err)
			return
		}

		if err := RespondWithJSON(w, http.StatusOK, tx); err != nil {
			s.logger.Error(err)
		}
	}
}

// TxHistory returns the full list of to and from transactions from the database
func (s *Service) TxHistory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var c database.Account
		if err := DecodeJSON(r.Body, &c); err != nil {
			RespondWithError(w, http.StatusBadRequest, err)
			return
		}

		if c.ID == 0 {
			err := fmt.Errorf("cannot supply account ID = 0")
			RespondWithError(w, http.StatusBadRequest, err)
			return
		}

		// Execute Query against PSQL
		txs, err := s.dbClient.NewQuery().GetUserTransactions(context.Background())
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

}

// POST Requests

// CreateAccount validates then writes a new account to the database
// Once registered the new account will have a unique ID number.
func (s *Service) CreateAccount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var c database.CreateAccountParams
		if err := DecodeJSON(r.Body, &c); err != nil {
			RespondWithError(w, http.StatusBadRequest, err)
			return
		}

		// validate inputs
		if err := validAccountParams(c); err != nil {
			RespondWithError(w, http.StatusBadRequest, err)
			return
		}

		// Verify uniqueness
		if u, _ := s.dbClient.NewQuery().GetUserByUsername(context.Background(), c.Username); u.ID != 0 {
			RespondWithError(w, http.StatusBadRequest, fmt.Errorf("username already exists"))
			return
		}

		if u, _ := s.dbClient.NewQuery().GetUserByEmail(context.Background(), c.Email); u.Email.Valid {
			RespondWithError(w, http.StatusBadRequest, fmt.Errorf("email already exists"))
			return
		}

		// Execute Query against PSQL
		acc, err := s.dbClient.NewQuery().CreateAccount(context.Background(), database.CreateAccountParams{
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

}

// CreateTx posts a new transaction to the DB. Transaction fields
// are validated before the tx is registered
func (s *Service) CreateTx() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var txParams database.CreateTransactionParams
		if err := DecodeJSON(r.Body, &txParams); err != nil {
			RespondWithError(w, http.StatusBadRequest, err)
			return
		}

		// Validate amount
		if txParams.Amount.Int64 <= 0 {
			err := fmt.Errorf("cannot send negative amount '%v'", txParams.Amount)
			RespondWithError(w, http.StatusBadRequest, err)
			return
		}

		if txParams.FromAccount.Int64 == txParams.ToAccount.Int64 {
			err := fmt.Errorf("to and from account cannot match")
			RespondWithError(w, http.StatusBadRequest, err)
			return
		}

		// Check to and from account exist
		if _, err := s.dbClient.NewQuery().GetUser(context.Background(), txParams.FromAccount.Int64); err != nil {
			RespondWithError(w, http.StatusBadRequest, err)
			return
		}
		if _, err := s.dbClient.NewQuery().GetUser(context.Background(), txParams.ToAccount.Int64); err != nil {
			RespondWithError(w, http.StatusBadRequest, err)
			return
		}

		// Execute Query against PSQL
		tx, err := s.dbClient.NewQuery().CreateTransaction(context.Background(), database.CreateTransactionParams{
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

}
