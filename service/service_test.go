package service

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

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
	log, err := NewLogger("info", "plain", false, n)
	if err != nil {
		t.Fatal(err)
	}

	s := New(8080, log, database.NewMemDB())

	apiTests := []struct {
		name             string
		endpoint         string
		body             func() []byte
		method           func(w http.ResponseWriter, req *http.Request)
		expectedResponse any
		expectedCode     int
	}{
		//
		// READ REQUESTS
		//
		{
			"status",
			Status,
			func() []byte { return nil },
			s.Status,
			&StatusResponse{Message: "OK", Version: FullVersion, Service: serviceName},
			http.StatusOK,
		},
		{
			"health",
			Health,
			func() []byte { return nil },
			s.Health,
			&HealthResponse{Version: FullVersion, Service: serviceName, Failures: []string{}},
			http.StatusOK,
		},
		{
			"accounts",
			GetAccount,
			func() []byte { return nil },
			s.Accounts,
			&[]database.Account{database.Account{ID: 1}},
			http.StatusOK,
		},
		{
			"account-by-index",
			GetAccount,
			func() []byte {
				accParams := database.Account{ID: 1}
				b, err := json.Marshal(accParams)
				if err != nil {
					panic(err)
				}
				return b
			},
			s.AccountByIndex,
			&database.Account{ID: 1},
			http.StatusOK,
		},
		{
			"account-by-email",
			GetAccountByEmail,
			func() []byte {
				e := "name@emailprovider.com"
				accParams := database.Account{Email: sql.NullString{String: e}}
				b, err := json.Marshal(accParams)
				if err != nil {
					panic(err)
				}
				return b
			},
			s.AccountByEmail,
			&database.Account{ID: 1, Email: sql.NullString{String: "name@emailprovider.com"}},
			http.StatusOK,
		},
		{
			"account-by-username",
			GetAccountByUsername,
			func() []byte {
				usr := "myusername"
				accParams := database.Account{Username: usr}
				b, err := json.Marshal(accParams)
				if err != nil {
					panic(err)
				}
				return b
			},
			s.AccountByUsername,
			&database.Account{ID: 1, Username: "myusername"},
			http.StatusOK,
		},
		{
			"transaction-by-id",
			GetTransactionByIndex,
			func() []byte {
				accParams := database.Transaction{ID: 1}
				b, err := json.Marshal(accParams)
				if err != nil {
					panic(err)
				}
				return b
			},
			s.TransactionByIndex,
			&database.Transaction{ID: 1},
			http.StatusOK,
		},
		{
			"account-txs",
			GetAccountTransactions,
			func() []byte {
				accParams := database.Account{ID: 1}
				b, err := json.Marshal(accParams)
				if err != nil {
					panic(err)
				}
				return b
			},
			s.TxHistory,
			&[]database.GetUserTransactionsRow{database.GetUserTransactionsRow{TransactionID: 1, FromAccountID: sql.NullInt64{Int64: 1}, ToAccountID: sql.NullInt64{Int64: 2}, Amount: sql.NullInt64{Int64: 1}}},
			http.StatusOK,
		},
		//
		// WRITE REQUESTS
		//
		{
			"create-account",
			CreateAccount,
			func() []byte {
				accParams := database.Account{ID: 1}
				b, err := json.Marshal(accParams)
				if err != nil {
					panic(err)
				}
				return b
			},
			s.CreateAccount,
			&database.Account{ID: 1},
			http.StatusOK,
		},
		{
			"create-transaction",
			CreateTx,
			func() []byte {
				accParams := database.Transaction{ID: 1, FromAccount: sql.NullInt64{Int64: 1}, ToAccount: sql.NullInt64{Int64: 2}, Amount: sql.NullInt64{Int64: 1}}
				b, err := json.Marshal(accParams)
				if err != nil {
					panic(err)
				}
				return b
			},
			s.CreateTx,
			&database.Transaction{ID: 1, FromAccount: sql.NullInt64{Int64: 1}, ToAccount: sql.NullInt64{Int64: 2}, Amount: sql.NullInt64{Int64: 1}},
			http.StatusOK,
		},
	}

	for _, tt := range apiTests {
		t.Run(tt.name, func(t *testing.T) {
			httpReq := httptest.NewRequest(http.MethodGet, tt.endpoint, bytes.NewReader(tt.body()))
			respRecorder := httptest.NewRecorder()
			tt.method(respRecorder, httpReq)
			if g, w := respRecorder.Code, tt.expectedCode; g != w {
				t.Errorf("unexpected response code, want %v got %v", w, g)
			}
			expectedJSON, _ := json.Marshal(tt.expectedResponse)

			if g, w := respRecorder.Body.Bytes(), expectedJSON; !bytes.Equal(g, w) {
				t.Fatalf("unexpected response, want %s, got %s", w, g)
			}
		})
	}
}

/*
EndPoint{
	Path:       CreateTx,
	Handler:    s.CreateTx,
	MethodType: http.MethodPost,
},
EndPoint{
	Path:       CreateAccount,
	Handler:    s.CreateAccount,
	MethodType: http.MethodPost,
*/
