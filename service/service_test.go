package service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/ATMackay/psql-ledger/database"
	yaml "gopkg.in/yaml.v3"
)

// TODO

func Test_SantizeConfig(t *testing.T) {

	tests := []struct {
		name           string
		initialConfig  func() Config
		expectedConfig func() Config
		expectDefault  bool
	}{
		{
			"empty",
			func() Config {
				return emptyConfig
			},
			func() Config {
				return DefaultConfig
			},
			true,
		},
		{
			"empty-with-port",
			func() Config {
				cfg := emptyConfig
				cfg.Port = 1
				return cfg
			},
			func() Config {
				cfg := DefaultConfig
				cfg.Port = 1
				return cfg
			},
			false,
		},
		{
			"empty-with-log-level",
			func() Config {
				cfg := emptyConfig
				cfg.LogLevel = string(Info)
				return cfg
			},
			func() Config {
				cfg := DefaultConfig
				cfg.LogLevel = string(Info)
				return cfg
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, defaultUsed := sanitizeConfig(tt.initialConfig())
			if defaultUsed != tt.expectDefault {
				t.Errorf("unexpected result, expected 'defaultUsed' to be %v, got %v", tt.expectDefault, defaultUsed)
			}
			b, _ := yaml.Marshal(c)
			e, _ := yaml.Marshal(tt.expectedConfig())
			if !bytes.Equal(b, e) {
				t.Errorf("returned config not equal to default")
			}
		})
	}
}

func Test_ServiceStartStop(t *testing.T) {
	l, err := NewLogger("error", "plain", false, "test")
	if err != nil {
		t.Fatal(err)
	}
	service := New(8080, l, database.NewMemDB())

	service.Start()

	service.Stop(os.Kill)
}

func Test_API(t *testing.T) {
	n := "test"
	log, err := NewLogger("error", "plain", false, n)
	if err != nil {
		t.Fatal(err)
	}

	s := New(8080, log, database.NewMemDB())
	s.Start()
	t.Cleanup(func() {
		s.Stop(os.Kill)
	})
	time.Sleep(50 * time.Millisecond) // TODO - smell

	apiTests := []struct {
		name             string
		endpoint         string
		methodType       string
		body             func() []byte
		expectedResponse any
		expectedCode     int
	}{
		//
		// READ REQUESTS
		//
		{
			"status",
			Status,
			http.MethodGet,
			func() []byte { return nil },
			&StatusResponse{Message: "OK", Version: FullVersion, Service: serviceName},
			http.StatusOK,
		},
		{
			"health",
			Health,
			http.MethodGet,
			func() []byte { return nil },
			&HealthResponse{Version: FullVersion, Service: serviceName, Failures: []string{}},
			http.StatusOK,
		},
		{
			"accounts",
			Accounts,
			http.MethodGet,
			func() []byte { return nil },
			&[]database.Account{database.Account{ID: 1}},
			http.StatusOK,
		},
		{
			"account-by-index",
			GetAccount,
			http.MethodGet,
			func() []byte {
				accParams := database.Account{ID: 1}
				b, err := json.Marshal(accParams)
				if err != nil {
					panic(err)
				}
				return b
			},
			&database.Account{ID: 1},
			http.StatusOK,
		},
		{
			"account-by-email",
			GetAccountByEmail,
			http.MethodGet,
			func() []byte {
				e := "name@emailprovider.com"
				accParams := database.Account{Email: sql.NullString{String: e}}
				b, err := json.Marshal(accParams)
				if err != nil {
					panic(err)
				}
				return b
			},
			&database.Account{ID: 1, Email: sql.NullString{String: "name@emailprovider.com"}},
			http.StatusOK,
		},
		{
			"account-by-username",
			GetAccountByUsername,
			http.MethodGet,
			func() []byte {
				usr := "myusername"
				accParams := database.Account{Username: usr}
				b, err := json.Marshal(accParams)
				if err != nil {
					panic(err)
				}
				return b
			},
			&database.Account{ID: 1, Username: "myusername"},
			http.StatusOK,
		},
		{
			"transaction-by-id",
			GetTransactionByIndex,
			http.MethodGet,
			func() []byte {
				accParams := database.Transaction{ID: 1}
				b, err := json.Marshal(accParams)
				if err != nil {
					panic(err)
				}
				return b
			},
			&database.Transaction{ID: 1},
			http.StatusOK,
		},
		{
			"account-txs",
			GetAccountTransactions,
			http.MethodGet,
			func() []byte {
				accParams := database.Account{ID: 1}
				b, err := json.Marshal(accParams)
				if err != nil {
					panic(err)
				}
				return b
			},
			&[]database.GetUserTransactionsRow{database.GetUserTransactionsRow{TransactionID: 1, FromAccountID: sql.NullInt64{Int64: 1}, ToAccountID: sql.NullInt64{Int64: 2}, Amount: sql.NullInt64{Int64: 1}}},
			http.StatusOK,
		},
		//
		// WRITE REQUESTS
		//
		{
			"create-account",
			CreateAccount,
			http.MethodPost,
			func() []byte {
				accParams := database.Account{ID: 1}
				b, err := json.Marshal(accParams)
				if err != nil {
					panic(err)
				}
				return b
			},
			&database.Account{ID: 1},
			http.StatusOK,
		},
		{
			"create-transaction",
			CreateTx,
			http.MethodPost,
			func() []byte {
				accParams := database.Transaction{ID: 1, FromAccount: sql.NullInt64{Int64: 1}, ToAccount: sql.NullInt64{Int64: 2}, Amount: sql.NullInt64{Int64: 1}}
				b, err := json.Marshal(accParams)
				if err != nil {
					panic(err)
				}
				return b
			},
			&database.Transaction{ID: 1, FromAccount: sql.NullInt64{Int64: 1}, ToAccount: sql.NullInt64{Int64: 2}, Amount: sql.NullInt64{Int64: 1}},
			http.StatusOK,
		},
	}

	for _, tt := range apiTests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequestWithContext(context.Background(), tt.methodType, fmt.Sprintf("http://0.0.0.0%v%v", s.server.Addr(), tt.endpoint), bytes.NewReader(tt.body()))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			response, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()

			// Read the response body
			b, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatal(err)
			}
			if g, w := response.StatusCode, tt.expectedCode; g != w {
				t.Errorf("unexpected response code, want %v got %v", w, g)
			}
			expectedJSON, _ := json.Marshal(tt.expectedResponse)

			if g, w := b, expectedJSON; !bytes.Equal(g, w) {
				t.Fatalf("unexpected response, want %s, got %s", w, g)
			}
		})
	}
}
