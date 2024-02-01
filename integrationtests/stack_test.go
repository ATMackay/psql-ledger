package integrationtests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/ATMackay/psql-ledger/database"
	"github.com/ATMackay/psql-ledger/service"
)

const serviceName = "psqlledger"

func Test_StackAPI(t *testing.T) {

	stack := createStack(t)

	apiTests := []struct {
		name             string
		endpoint         string
		methodType       string
		expectedResponse any
		expectedCode     int
	}{
		{
			"status",
			service.Status,
			http.MethodGet,
			&service.StatusResponse{Message: "OK", Version: service.FullVersion, Service: serviceName},
			http.StatusOK,
		},
		{
			"health",
			service.Health,
			http.MethodGet,
			&service.HealthResponse{Version: service.FullVersion, Service: serviceName, Failures: []string{}},
			http.StatusOK,
		},
	}

	for _, tt := range apiTests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequestWithContext(context.Background(), tt.methodType, fmt.Sprintf("http://0.0.0.0%v%v", stack.psqlLedger.Server().Addr(), tt.endpoint), nil)
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

func Test_E2EReadWriteAccount(t *testing.T) {

	s := createStack(t)

	serverURL := fmt.Sprintf("http://0.0.0.0%v", s.psqlLedger.Server().Addr())

	// Healthcheck the stack
	response, err := executeRequest(http.MethodGet, serverURL+service.Health, nil, http.StatusOK)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()

	userName := "myusername"
	email := "myemail@provider.com"

	// Write a User Account to the DB
	accParams := database.CreateAccountParams{Username: userName, Email: sql.NullString{String: email}}
	b, err := json.Marshal(accParams)
	if err != nil {
		t.Fatal(err)
	}

	response, err = executeRequest(http.MethodPost, serverURL+service.CreateAccount, bytes.NewReader(b), http.StatusOK)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()

	queryData := database.Account{ID: 1}
	queryB, err := json.Marshal(queryData)
	if err != nil {
		t.Fatal(err)
	}

	// Check user account exists
	response, err = executeRequest(http.MethodGet, serverURL+service.GetAccount, bytes.NewReader(queryB), http.StatusOK)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	respB, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	responseData := &database.Account{}
	if err := json.Unmarshal(respB, responseData); err != nil {
		t.Fatal(err)
	}

	// Verify returned data

	if g, w := responseData.ID, int64(1); g != w {
		t.Fatalf("unexpected accountID, want %v got %v", w, g)
	}

	if g, w := responseData.Username, accParams.Username; g != w {
		t.Fatalf("unexpected account username, want %v got %v", w, g)
	}

	if g, w := responseData.Email, accParams.Email; g != w {
		t.Fatalf("unexpected account email, want %v got %v", w, g)
	}
}

func executeRequest(methodType, url string, body io.Reader, expectedCode int) (*http.Response, error) {
	req, err := http.NewRequestWithContext(context.Background(), methodType, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if g, w := response.StatusCode, expectedCode; g != w {
		return nil, fmt.Errorf("unexpected response code, want %v got %v", w, g)
	}
	return response, nil

}
