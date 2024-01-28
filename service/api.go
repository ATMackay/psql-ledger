package service

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/ATMackay/psql-ledger/database"
)

// Read Requests

func (s *Service) User(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func (s *Service) TxHistory(w http.ResponseWriter, r *http.Request) {
	// TODO
}

// Write Requests

type CreateAccountParams struct {
	Username string `json:"username"`
	Email    string `json:"email"`

	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *Service) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var c CreateAccountParams
	if err := DecodeJSON(r.Body, &c); err != nil {
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

	if err := RespondWithJSON(w, http.StatusOK, &CreateAccountParams{ID: acc.ID, Username: acc.Username, Email: acc.Email.String, CreatedAt: acc.CreatedAt.Time}); err != nil {
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

func (s *Service) Tx(w http.ResponseWriter, r *http.Request) {
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
