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

func TestIsValidString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		regex    string
		expected error
	}{
		{"valid email", "alex@emailprovider.com", emailRegex, nil},
		{"invalid email", "dhd$@xyz.com", emailRegex, fmt.Errorf(" 'dhd$@xyz.com' failed to match expression '%v'", emailRegex)},
		{"valid username", "user105", usernameRegex, nil},
		{"invalid username", "u£er101", usernameRegex, fmt.Errorf(" 'u£er101' failed to match expression '%v'", usernameRegex)},
		{"empty string", "", emailRegex, fmt.Errorf(" '' failed to match expression '%v'", emailRegex)},
		{"invalid regex", "test123", "invalid[regex", fmt.Errorf("error parsing regexp: missing closing ]: `[regex`")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := isValidString(test.input, test.regex)

			if test.expected == nil && err != nil {
				t.Fatalf("Expected success, but got error: %v", err)
			}

			if test.expected != nil && err == nil {
				t.Fatalf("Expected error: %v, but got success", test.expected)
			}

			if test.expected != nil && err != nil && test.expected.Error() != err.Error() {
				t.Fatalf("Expected error: %v, but got: %v", test.expected, err)
			}
		})
	}
}

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
	service := New(8080, 1, l, database.NewMemoryDBClient())

	service.Start()

	service.Stop(os.Kill)
}

func Test_API(t *testing.T) {
	n := "test"
	log, err := NewLogger("error", "plain", false, n)
	if err != nil {
		t.Fatal(err)
	}

	s := New(8080, 1, log, database.NewMemoryDBClient())
	s.Start()
	t.Cleanup(func() {
		s.Stop(os.Kill)
	})
	time.Sleep(50 * time.Millisecond) // TODO - smell

	testAccount := database.Account{ID: 1, Username: "myusername", Email: sql.NullString{String: "myname@emailprovider.com"}}
	testAccount2 := database.Account{ID: 2, Username: "yourusername", Email: sql.NullString{String: "yourname@emailprovider.com"}}
	testTx := database.Transaction{ID: 1, FromAccount: sql.NullInt64{Int64: 1}, ToAccount: sql.NullInt64{Int64: 2}, Amount: sql.NullInt64{Int64: 1}}

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
		//
		// WRITE REQUESTS
		//
		{
			"create-account",
			CreateAccount,
			http.MethodPut,
			func() []byte {
				b, err := json.Marshal(testAccount)
				if err != nil {
					panic(err)
				}
				return b
			},
			testAccount,
			http.StatusOK,
		},
		{
			"create-account-2",
			CreateAccount,
			http.MethodPut,
			func() []byte {
				b, err := json.Marshal(testAccount2)
				if err != nil {
					panic(err)
				}
				return b
			},
			testAccount2,
			http.StatusOK,
		},
		//
		// READ REQUESTS WITH DB LOOKUP
		//
		{
			"create-transaction",
			CreateTx,
			http.MethodPut,
			func() []byte {
				b, err := json.Marshal(testTx)
				if err != nil {
					panic(err)
				}
				return b
			},
			testTx,
			http.StatusOK,
		},
		{
			"accounts",
			Accounts,
			http.MethodGet,
			func() []byte { return nil },
			&[]database.Account{testAccount, testAccount2},
			http.StatusOK,
		},
		{
			"account-by-index",
			GetAccount,
			http.MethodPost,
			func() []byte {
				accParams := database.Account{ID: 1}
				b, err := json.Marshal(accParams)
				if err != nil {
					panic(err)
				}
				return b
			},
			testAccount,
			http.StatusOK,
		},
		{
			"account-by-email",
			GetAccountByEmail,
			http.MethodPost,
			func() []byte {
				accParams := &database.Account{Email: sql.NullString{String: "myname@emailprovider.com"}}
				b, err := json.Marshal(accParams)
				if err != nil {
					panic(err)
				}
				return b
			},
			testAccount,
			http.StatusOK,
		},
		{
			"account-by-username",
			GetAccountByUsername,
			http.MethodPost,
			func() []byte {
				accParams := database.Account{Username: "myusername"}
				b, err := json.Marshal(accParams)
				if err != nil {
					panic(err)
				}
				return b
			},
			testAccount,
			http.StatusOK,
		},
		{
			"transaction-by-id",
			GetTransactionByIndex,
			http.MethodPost,
			func() []byte {
				accParams := database.Transaction{ID: 1}
				b, err := json.Marshal(accParams)
				if err != nil {
					panic(err)
				}
				return b
			},
			testTx,
			http.StatusOK,
		},
		{
			"account-txs",
			GetAccountTransactions,
			http.MethodPost,
			func() []byte {
				accParams := database.Account{ID: 1}
				b, err := json.Marshal(accParams)
				if err != nil {
					panic(err)
				}
				return b
			},
			&[]database.GetUserTransactionsRow{database.GetUserTransactionsRow{TransactionID: testTx.ID, FromAccountID: testTx.FromAccount, ToAccountID: testTx.ToAccount, Amount: testTx.Amount}},
			http.StatusOK,
		},
	}

	for _, tt := range apiTests {
		req, err := http.NewRequestWithContext(context.Background(), tt.methodType, fmt.Sprintf("http://0.0.0.0%v%v", s.server.Addr(), tt.endpoint), bytes.NewReader(tt.body()))
		if err != nil {
			t.Fatalf("%v: %v", tt.name, err)
		}
		req.Header.Set("Content-Type", "application/json")

		response, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("%v: %v", tt.name, err)
		}
		defer response.Body.Close()

		// Read the response body
		b, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}
		if g, w := response.StatusCode, tt.expectedCode; g != w {
			t.Errorf("%v unexpected response code, want %v got %v", tt.name, w, g)
		}
		expectedJSON, _ := json.Marshal(tt.expectedResponse)

		if g, w := b, expectedJSON; !bytes.Equal(g, w) {
			t.Errorf("%v unexpected response, want %s, got %s", tt.name, w, g)
		}

	}
}
