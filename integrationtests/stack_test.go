package integrationtests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/ATMackay/psql-ledger/database"
	"github.com/ATMackay/psql-ledger/service"
)

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
			service.StatusEndPnt,
			http.MethodGet,
			&service.StatusResponse{Message: "OK", Version: service.Version, Service: service.ServiceName},
			http.StatusOK,
		},
		{
			"health",
			service.HealthEndPnt,
			http.MethodGet,
			&service.HealthResponse{Version: service.Version, Service: service.ServiceName, Failures: []string{}},
			http.StatusOK,
		},
	}

	for _, tt := range apiTests {
		t.Run(tt.name, func(t *testing.T) {

			response, err := executeRequest(http.MethodGet, fmt.Sprintf("http://0.0.0.0%v%v", stack.psqlLedger.Server().Addr(), tt.endpoint), nil)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if g, w := response.StatusCode, http.StatusOK; g != w {
				t.Fatalf("expected %v, got %v", w, g)
			}

			// Read the response body
			b, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatal(err)
			}
			expectedJSON, _ := json.Marshal(tt.expectedResponse)

			if g, w := b, expectedJSON; !bytes.Equal(g, w) {
				t.Fatalf("unexpected response, want %s, got %s, response = %s", w, g, b)
			}
		})
	}
}

func Test_E2EReadWriteAccount(t *testing.T) {

	s := createStack(t)

	serverURL := fmt.Sprintf("http://0.0.0.0%v", s.psqlLedger.Server().Addr())

	// Healthcheck the stack
	response, err := executeRequest(http.MethodGet, serverURL+service.HealthEndPnt, nil)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if g, w := response.StatusCode, http.StatusOK; g != w {
		t.Fatalf("expected %v, got %v", w, g)
	}

	userName := "myusername"
	email := "myemail@provider.com"

	// Write a User Account to the DB
	accParams := database.CreateAccountParams{Username: userName, Email: sql.NullString{String: email}}
	b, err := json.Marshal(accParams)
	if err != nil {
		t.Fatal(err)
	}

	response, err = executeRequest(http.MethodPut, serverURL+service.CreateAccountEndPnt, bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if g, w := response.StatusCode, http.StatusOK; g != w {
		t.Fatalf("expected %v, got %v", w, g)
	}

	queryData := database.Account{ID: 1}
	queryB, err := json.Marshal(queryData)
	if err != nil {
		t.Fatal(err)
	}

	// Check user account exists
	response, err = executeRequest(http.MethodPost, serverURL+service.GetAccountEndPnt, bytes.NewReader(queryB))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if g, w := response.StatusCode, http.StatusOK; g != w {
		t.Fatalf("expected %v, got %v", w, g)
	}

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
}

// Manual 'benchmark' tests

func Test_MultipleWrites(t *testing.T) {
	s := createStack(t)

	N := 1000 // increase for better test accuracy

	serverURL := fmt.Sprintf("http://0.0.0.0%v", s.psqlLedger.Server().Addr())

	// Setup

	// Healthcheck the stack
	response, err := executeRequest(http.MethodGet, serverURL+service.HealthEndPnt, nil)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if g, w := response.StatusCode, http.StatusOK; g != w {
		t.Fatalf("expected %v, got %v", w, g)
	}

	userName := "myusername"
	email := "myemail@provider.com"

	// Write a User Account to the DB

	// create input data
	reqArray := [][]byte{}
	for n := 0; n < N; n++ {
		accParams := database.CreateAccountParams{Username: userName + fmt.Sprintf("%v", n), Email: sql.NullString{String: fmt.Sprintf("%v", n) + email}}
		accBytes, err := json.Marshal(accParams)
		if err != nil {
			t.Fatal(err)
		}
		reqArray = append(reqArray, accBytes)
	}

	start := time.Now()
	for n := range N {
		response, err := executeRequest(http.MethodPut, serverURL+service.CreateAccountEndPnt, bytes.NewReader(reqArray[n]))
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
	}
	if g, w := response.StatusCode, http.StatusOK; g != w {
		t.Fatalf("expected %v, got %v", w, g)
	}
	elapsed := time.Since(start)
	fmt.Printf("completed %d writes in %v milliseconds (%v/s)", N, elapsed, (float64(N) * 1000.0 / float64(elapsed.Milliseconds())))
}

func Test_MultipleReads(t *testing.T) {
	s := createStack(t)

	N := 100 // increase for better test accuracy

	serverURL := fmt.Sprintf("http://0.0.0.0%v", s.psqlLedger.Server().Addr())

	// Setup

	// Healthcheck the stack
	response, err := executeRequest(http.MethodGet, serverURL+service.HealthEndPnt, nil)
	if err != nil {
		t.Fatal(err)
	}
	b, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if g, w := response.StatusCode, http.StatusOK; g != w {
		t.Fatalf("expected %v, got %v, resp = %s", w, g, b)
	}
	if g, w := response.StatusCode, http.StatusOK; g != w {
		t.Fatalf("expected %v, got %v", w, g)
	}
	userName := "myusername"
	email := "myemail@provider.com"

	// Write a User Account to the DB
	accParams := database.CreateAccountParams{Username: userName, Email: sql.NullString{String: email}}
	accBytes, err := json.Marshal(accParams)
	if err != nil {
		t.Fatal(err)
	}
	response, err = executeRequest(http.MethodPut, serverURL+service.CreateAccountEndPnt, bytes.NewReader(accBytes))
	if err != nil {
		t.Fatal(err)
	}
	b, err = io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if g, w := response.StatusCode, http.StatusOK; g != w {
		t.Fatalf("expected %v, got %v, resp = %s", w, g, b)
	}
	queryData := database.Account{ID: 1}
	queryB, err := json.Marshal(queryData)
	if err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	for range N {
		// fetch user account
		if _, err = executeRequest(http.MethodPost, serverURL+service.GetAccountEndPnt, bytes.NewReader(queryB)); err != nil {
			t.Fatal(err)
		}
		if g, w := response.StatusCode, http.StatusOK; g != w {
			t.Fatalf("expected %v, got %v", w, g)
		}
	}
	elapsed := time.Since(start)
	fmt.Printf("completed %d writes in %v milliseconds (%v/s)", N, elapsed, (float64(N) * 1000.0 / float64(elapsed.Milliseconds())))
}
