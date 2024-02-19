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
	response, err := executeRequest(http.MethodGet, serverURL+service.Health, nil, http.StatusOK)
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
		response, err := executeRequest(http.MethodPut, serverURL+service.CreateAccount, bytes.NewReader(accBytes), http.StatusOK)
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
	response, err = executeRequest(http.MethodPut, serverURL+service.CreateAccount, bytes.NewReader(accBytes), http.StatusOK)
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
		if _, err = executeRequest(http.MethodPost, serverURL+service.GetAccount, bytes.NewReader(queryB), http.StatusOK); err != nil {
			b.Fatal(err)
		}
	}

}
