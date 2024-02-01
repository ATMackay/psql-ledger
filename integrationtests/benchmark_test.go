package integrationtests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/ATMackay/psql-ledger/database"
	"github.com/ATMackay/psql-ledger/service"
)

// TODO - fix

func BenchmarkAccountWrite(b *testing.B) {

	s := createStack(b)

	serverURL := fmt.Sprintf("http://0.0.0.0%v", s.psqlLedger.Server().Addr())

	// Setup

	// Healthcheck the stack
	response, err := executeRequest(http.MethodGet, serverURL+service.Health, nil, http.StatusOK)
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

	for n := 0; n < b.N; n++ {
		response, err := executeRequest(http.MethodPost, serverURL+service.CreateAccount, bytes.NewReader(accBytes), http.StatusOK)
		if err != nil {
			b.Fatal(err)
		}
		response.Body.Close()
	}

}

func BenchmarkAccountRead(b *testing.B) {

	s := createStack(b)

	serverURL := fmt.Sprintf("http://0.0.0.0%v", s.psqlLedger.Server().Addr())

	// Setup

	// Healthcheck the stack
	response, err := executeRequest(http.MethodGet, serverURL+service.Health, nil, http.StatusOK)
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
	response, err = executeRequest(http.MethodPost, serverURL+service.CreateAccount, bytes.NewReader(accBytes), http.StatusOK)
	if err != nil {
		b.Fatal(err)
	}
	response.Body.Close()

	queryData := database.Account{ID: 1}
	queryB, err := json.Marshal(queryData)
	if err != nil {
		b.Fatal(err)
	}

	for n := 0; n < b.N; n++ {
		// fetch user account
		if _, err = executeRequest(http.MethodGet, serverURL+service.GetAccount, bytes.NewReader(queryB), http.StatusOK); err != nil {
			b.Fatal(err)
		}
	}

}
