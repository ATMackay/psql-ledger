package integrationtests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ATMackay/psql-ledger/service"
)

func Test_PSQLContainer(t *testing.T) {
	ctx := context.Background()
	psqlContainer, err := startPSQLContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := psqlContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})
}

func Test_Stack(t *testing.T) {

	stack := createStack(t)

	apiTests := []struct {
		name             string
		endpoint         string
		method           func(w http.ResponseWriter, req *http.Request)
		expectedResponse any
		expectedCode     int
	}{
		{
			"status",
			service.Status,
			stack.psqlLedger.Status,
			&service.StatusResponse{Message: "OK", Version: service.FullVersion, Service: "psqlledger"},
			http.StatusOK,
		},
		{
			"health",
			service.Health,
			stack.psqlLedger.Health,
			&service.HealthResponse{Version: service.FullVersion, Service: "psqlledger", Failures: []string{}},
			http.StatusOK,
		},
	}

	for _, tt := range apiTests {
		t.Run(tt.name, func(t *testing.T) {
			httpReq := httptest.NewRequest(http.MethodGet, tt.endpoint, nil)
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
