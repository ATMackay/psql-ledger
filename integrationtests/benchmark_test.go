package integrationtests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/ATMackay/psql-ledger/database"
	"github.com/ATMackay/psql-ledger/service"
)

// These tests require a locally running stack
//
// make postgresup
// make createdb
// make run
func BenchmarkAccountWrite(b *testing.B) {

	// Setup

	serverURL := "http://0.0.0.0:8080"

	// Healthcheck the stack
	response, err := executeRequest(http.MethodGet, serverURL+service.HealthEndPnt, nil, http.StatusOK)
	if err != nil {
		b.Fatal(err)
	}
	response.Body.Close()

	userName := "myusername"
	email := "myemail@provider.com"

	// Write a User Account to the DB
	rand.New(rand.NewSource(time.Now().UnixNano()))

	for n := 0; n < b.N; n++ {
		r := rand.Int63()
		id := strconv.FormatInt(r, 10)
		accParams := database.CreateAccountParams{Username: userName + id, Email: sql.NullString{String: id + email}}
		accBytes, err := json.Marshal(accParams)
		if err != nil {
			b.Fatal(err)
		}
		response, err := executeRequest(http.MethodPut, serverURL+service.CreateAccountEndPnt, bytes.NewReader(accBytes), http.StatusOK)
		if err != nil {
			b.Fatal(err)
		}
		response.Body.Close()
	}

}

func BenchmarkAccountRead(b *testing.B) {

	// Setup

	serverURL := "http://0.0.0.0:8080"

	// Healthcheck the stack
	response, err := executeRequest(http.MethodGet, serverURL+service.HealthEndPnt, nil, http.StatusOK)
	if err != nil {
		b.Fatal(err)
	}
	response.Body.Close()

	userName := "myusername"
	email := "myemail@provider.com"

	// Write a User Account to the DB
	accParams := database.CreateAccountParams{Username: userName, Email: sql.NullString{String: email}}
	accBytes, err := json.Marshal(accParams)
	if err != nil {
		b.Fatal(err)
	}
	response, err = executeRequest(http.MethodPut, serverURL+service.CreateAccountEndPnt, bytes.NewReader(accBytes), http.StatusOK)
	if err != nil {
		b.Logf("account may already exist: %v", err)
	}
	if response != nil {
		response.Body.Close()
	}

	queryData := database.Account{ID: 1}
	queryB, err := json.Marshal(queryData)
	if err != nil {
		b.Fatal(err)
	}

	for n := 0; n < b.N; n++ {
		// fetch user account
		if _, err = executeRequest(http.MethodPost, serverURL+service.GetAccountEndPnt, bytes.NewReader(queryB), http.StatusOK); err != nil {
			b.Fatal(err)
		}
	}

}

func BenchmarkTransactionWrite(b *testing.B) {

	// Setup

	serverURL := "http://0.0.0.0:8080"

	// Healthcheck the stack
	response, err := executeRequest(http.MethodGet, serverURL+service.HealthEndPnt, nil, http.StatusOK)
	if err != nil {
		b.Fatal(err)
	}
	response.Body.Close()

	// Write a User 1 Account to the DB
	accParams := database.CreateAccountParams{Username: "myusername", Email: sql.NullString{String: "myemail@provider.com"}}
	accBytes, err := json.Marshal(accParams)
	if err != nil {
		b.Fatal(err)
	}
	response, _ = executeRequest(http.MethodPut, serverURL+service.CreateAccountEndPnt, bytes.NewReader(accBytes), http.StatusOK)
	if response != nil {
		response.Body.Close()
	}

	// Write a User 2 Account to the DB
	accParams = database.CreateAccountParams{Username: "yourusername", Email: sql.NullString{String: "youremail@provider.com"}}
	accBytes, err = json.Marshal(accParams)
	if err != nil {
		b.Fatal(err)
	}
	response, _ = executeRequest(http.MethodPut, serverURL+service.CreateAccountEndPnt, bytes.NewReader(accBytes), http.StatusOK)
	if response != nil {
		response.Body.Close()
	}

	for n := 0; n < b.N; n++ {
		// Prepare tx
		TxData := database.Transaction{FromAccount: sql.NullInt64{Int64: 1}, ToAccount: sql.NullInt64{Int64: 2}, Amount: sql.NullInt64{Int64: 1}, CreatedAt: sql.NullTime{Time: time.Now()}}
		TxB, err := json.Marshal(TxData)
		if err != nil {
			b.Fatal(err)
		}
		// Write transaction
		if _, err = executeRequest(http.MethodPut, serverURL+service.CreateTxEndPnt, bytes.NewReader(TxB), http.StatusOK); err != nil {
			b.Fatal(err)
		}
	}

}
