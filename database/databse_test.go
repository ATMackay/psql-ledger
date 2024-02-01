package database

import (
	"context"
	"database/sql"
	"testing"
)

func TestMemDBQuery_CreateAccount(t *testing.T) {
	dbClient := NewMemoryDBClient()
	ctx := context.Background()
	userName := "testuser"
	email := sql.NullString{String: "test@example.com", Valid: true}

	params := CreateAccountParams{
		Username: userName,
		Email:    email,
	}

	// Test CreateAccount
	if _, err := dbClient.NewQuery().CreateAccount(ctx, params); err != nil {
		t.Errorf("Error creating user: %v", err)
	}

	// Test GetUserByEmail
	user, err := dbClient.NewQuery().GetUserByEmail(ctx, email)
	if err != nil {
		t.Errorf("Error retrieving user by email: %v", err)
	}

	// Verify the retrieved user
	if user.ID != 1 || user.Email.String != email.String || !user.Email.Valid {
		t.Errorf("Unexpected retrieved user data: %+v", user)
	}

	// Test GetUserByUsername
	user, err = dbClient.NewQuery().GetUserByUsername(ctx, userName)
	if err != nil {
		t.Errorf("Error retrieving user by email: %v", err)
	}

	// Verify the retrieved user
	if user.ID != 1 || user.Username != userName || !user.Email.Valid {
		t.Errorf("Unexpected retrieved user data: %+v", user)
	}
}

func TestMemDBClient_CreateTransaction(t *testing.T) {
	dbClient := NewMemoryDBClient()
	ctx := context.Background()
	params := CreateTransactionParams{
		FromAccount: sql.NullInt64{Int64: 1},
		ToAccount:   sql.NullInt64{Int64: 2},
		Amount:      sql.NullInt64{Int64: 100},
	}

	// Test CreateTransaction
	transaction, err := dbClient.NewQuery().CreateTransaction(ctx, params)
	if err != nil {
		t.Errorf("Error creating transaction: %v", err)
	}

	if transaction.ID != 1 || transaction.FromAccount.Int64 != 1 || transaction.ToAccount.Int64 != 2 || transaction.Amount.Int64 != 100 {
		t.Errorf("Unexpected transaction data: %+v", transaction)
	}

	retrievedTransaction, err := dbClient.NewQuery().GetTx(ctx, 1)
	if err != nil {
		t.Errorf("Error retrieving transaction: %v", err)
	}

	if retrievedTransaction.ID != 1 || retrievedTransaction.FromAccount.Int64 != 1 || retrievedTransaction.ToAccount.Int64 != 2 || retrievedTransaction.Amount.Int64 != 100 {
		t.Errorf("Unexpected retrieved transaction data: %+v", retrievedTransaction)
	}
}

func TestMemDBQuery_DeleteAccount(t *testing.T) {
	dbClient := NewMemoryDBClient()
	ctx := context.Background()

	// Test DeleteAccount
	err := dbClient.NewQuery().DeleteAccount(ctx, 1)
	if err != nil {
		t.Errorf("Error deleting account: %v", err)
	}
}

func TestMemDBClient_CheckDatabaseExists(t *testing.T) {
	dbClient := NewMemoryDBClient()
	ctx := context.Background()

	// Test CheckDatabaseExists
	exists, err := dbClient.CheckDatabaseExists(ctx, "bank")
	if err != nil {
		t.Errorf("Error checking database existence: %v", err)
	}

	// Verify the existence
	if !exists {
		t.Errorf("Expected database 'bank' to exist, but it does not")
	}
}
