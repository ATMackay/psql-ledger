package service

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/ATMackay/psql-ledger/database"
)

const (
	StatusEndPnt = "/status"
	HealthEndPnt = "/health"

	AccountsEndPnt             = "/accounts"
	GetAccountEndPnt           = "/account-by-index"
	GetAccountByEmailEndPnt    = "/account-by-email"
	GetAccountByUsernameEndPnt = "/account-by-username"

	GetAccountTransactionsEndPnt = "/account-txs"

	GetTransactionByIndexEndPnt = "/tx"

	CreateTxEndPnt      = "/create-tx"
	CreateAccountEndPnt = "/create-account"
)

func makeServiceAPIs(dbClient database.DBClient) *API {
	return MakeAPI([]EndPoint{
		{
			Path:       StatusEndPnt,
			Handler:    Status(),
			MethodType: http.MethodGet,
		},
		{
			Path:       HealthEndPnt,
			Handler:    Health(dbClient),
			MethodType: http.MethodGet,
		},
		{
			Path:       AccountsEndPnt,
			Handler:    Accounts(dbClient),
			MethodType: http.MethodGet,
		},
		{
			Path:       GetAccountEndPnt,
			Handler:    AccountByIndex(dbClient),
			MethodType: http.MethodPost,
		},
		{
			Path:       GetAccountByEmailEndPnt,
			Handler:    AccountByEmail(dbClient),
			MethodType: http.MethodPost,
		},
		{
			Path:       GetAccountByUsernameEndPnt,
			Handler:    AccountByUsername(dbClient),
			MethodType: http.MethodPost,
		},
		{
			Path:       GetAccountTransactionsEndPnt,
			Handler:    TxHistory(dbClient),
			MethodType: http.MethodPost,
		},
		{
			Path:       GetTransactionByIndexEndPnt,
			Handler:    TransactionByIndex(dbClient),
			MethodType: http.MethodPost,
		},
		{
			Path:       CreateTxEndPnt,
			Handler:    CreateTx(dbClient),
			MethodType: http.MethodPut,
		},
		{
			Path:       CreateAccountEndPnt,
			Handler:    CreateAccount(dbClient),
			MethodType: http.MethodPut,
		},
	})
}

// GET REQUESTS

// StatusResponse contains status response fields.
type StatusResponse struct {
	Message string `json:"message,omitempty"`
	Version string `json:"version,omitempty"`
	Service string `json:"service,omitempty"`
}

// Status implements the status request endpoint. Always returns OK.
func Status() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := RespondWithJSON(w, http.StatusOK, &StatusResponse{Message: "OK", Version: Version, Service: ServiceName}); err != nil {
			RespondWithError(w, http.StatusInternalServerError, err)
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
func Health(dbClient database.DBClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		health := &HealthResponse{
			Service: ServiceName,
			Version: Version,
		}
		var failures = []string{}
		var httpCode = http.StatusOK
		if dbClient == nil || dbClient.DB() == nil {
			failures = append(failures, "DB: No connection")
			httpCode = http.StatusServiceUnavailable
			health.Failures = failures
			if err := RespondWithJSON(w, httpCode, health); err != nil {
				RespondWithError(w, http.StatusInternalServerError, err)
			}
			return
		}

		if err := dbClient.DB().Ping(); err != nil {
			failures = append(failures, fmt.Sprintf("DB: %v", err))
			httpCode = http.StatusServiceUnavailable
		}

		health.Failures = failures

		if err := RespondWithJSON(w, httpCode, health); err != nil {
			RespondWithError(w, http.StatusInternalServerError, err)
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

// Accounts requests the full list if accounts stored in the DB - TODO paginate this request
func Accounts(dbClient database.DBClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Execute Query against PSQL
		acc, err := dbClient.NewQuery().GetUsers(context.Background())
		if err != nil {
			if err.Error() != database.ErrNotFound.Error() {
				RespondWithError(w, http.StatusInternalServerError, err)
				return
			}
			RespondWithError(w, http.StatusNotFound, err)
			return
		}

		if err := RespondWithJSON(w, http.StatusOK, acc); err != nil {
			RespondWithError(w, http.StatusInternalServerError, err)
		}

	}

}

// POST REQUESTS

// AccountByIndex requests the account for supplied ID number
func AccountByIndex(dbClient database.DBClient) http.HandlerFunc {
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
		acc, err := dbClient.NewQuery().GetUser(context.Background(), c.ID)
		if err != nil {
			if err.Error() != database.ErrNotFound.Error() {
				RespondWithError(w, http.StatusInternalServerError, err)
				return
			}
			RespondWithError(w, http.StatusNotFound, err)
			return
		}

		if err := RespondWithJSON(w, http.StatusOK, acc); err != nil {
			RespondWithError(w, http.StatusInternalServerError, err)
		}

	}

}

// AccountByUsername requests the account for supplied ID number
func AccountByUsername(dbClient database.DBClient) http.HandlerFunc {
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
		acc, err := dbClient.NewQuery().GetUserByUsername(context.Background(), c.Username)
		if err != nil {
			if err.Error() != database.ErrNotFound.Error() {
				RespondWithError(w, http.StatusInternalServerError, err)
				return
			}
			RespondWithError(w, http.StatusNotFound, err)
			return
		}

		// Another zero response is an empty Account struct
		if acc.ID == 0 {
			RespondWithError(w, http.StatusNotFound, database.ErrNotFound)
			return
		}

		if err := RespondWithJSON(w, http.StatusOK, acc); err != nil {
			RespondWithError(w, http.StatusInternalServerError, err)
		}

	}

}

// AccountByUsername requests the account for supplied ID number
func AccountByEmail(dbClient database.DBClient) http.HandlerFunc {
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
		acc, err := dbClient.NewQuery().GetUserByEmail(context.Background(), c.Email)
		if err != nil {
			if err.Error() != database.ErrNotFound.Error() {
				RespondWithError(w, http.StatusInternalServerError, err)
				return
			}
			RespondWithError(w, http.StatusNotFound, err)
			return
		}

		// Another zero response is an empty Account struct
		if acc.ID == 0 {
			RespondWithError(w, http.StatusNotFound, database.ErrNotFound)
			return
		}

		if err := RespondWithJSON(w, http.StatusOK, acc); err != nil {
			RespondWithError(w, http.StatusInternalServerError, err)
		}
	}

}

// TransactionByIndex requests the transaction for supplied ID number - TODO paginate these requests
func TransactionByIndex(dbClient database.DBClient) http.HandlerFunc {
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
		tx, err := dbClient.NewQuery().GetTx(context.Background(), txParams.ID)
		if err != nil {
			if err.Error() != database.ErrNotFound.Error() {
				RespondWithError(w, http.StatusInternalServerError, err)
				return
			}
			RespondWithError(w, http.StatusNotFound, err)
			return
		}

		if tx.ID == 0 {
			RespondWithError(w, http.StatusNotFound, database.ErrNotFound)
			return
		}

		if err := RespondWithJSON(w, http.StatusOK, tx); err != nil {
			RespondWithError(w, http.StatusInternalServerError, err)
		}
	}
}

// TxHistory returns the full list of to and from transactions from the database
func TxHistory(dbClient database.DBClient) http.HandlerFunc {
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
		txs, err := dbClient.NewQuery().GetUserTransactions(context.Background())
		if err != nil {
			if err.Error() != database.ErrNotFound.Error() {
				RespondWithError(w, http.StatusInternalServerError, err)
				return
			}
			RespondWithError(w, http.StatusNotFound, err)
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
			RespondWithError(w, http.StatusInternalServerError, err)
		}
	}

}

// PUT Requests

// CreateAccount validates then writes a new account to the database
// Once registered the new account will have a unique ID number.
func CreateAccount(dbClient database.DBClient) http.HandlerFunc {
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
		if u, _ := dbClient.NewQuery().GetUserByUsername(context.Background(), c.Username); u.ID != 0 {
			RespondWithError(w, http.StatusBadRequest, fmt.Errorf("username already exists"))
			return
		}

		if u, _ := dbClient.NewQuery().GetUserByEmail(context.Background(), c.Email); u.Email.Valid {
			RespondWithError(w, http.StatusBadRequest, fmt.Errorf("email already exists"))
			return
		}

		// Execute Query against PSQL
		acc, err := dbClient.NewQuery().CreateAccount(context.Background(), database.CreateAccountParams{
			Email:    c.Email,
			Username: c.Username,
			Balance:  0,
		})
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, err)
			return
		}

		if err := RespondWithJSON(w, http.StatusOK, acc); err != nil {
			RespondWithError(w, http.StatusInternalServerError, err)
		}

	}

}

// CreateTx posts a new transaction to the DB. Transaction fields
// are validated before the tx is registered
func CreateTx(dbClient database.DBClient) http.HandlerFunc {
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
		if _, err := dbClient.NewQuery().GetUser(context.Background(), txParams.FromAccount.Int64); err != nil {
			RespondWithError(w, http.StatusBadRequest, err)
			return
		}
		if _, err := dbClient.NewQuery().GetUser(context.Background(), txParams.ToAccount.Int64); err != nil {
			RespondWithError(w, http.StatusBadRequest, err)
			return
		}

		// Execute Query against PSQL
		tx, err := dbClient.NewQuery().CreateTransaction(context.Background(), database.CreateTransactionParams{
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
			RespondWithError(w, http.StatusInternalServerError, err)
		}

	}

}
