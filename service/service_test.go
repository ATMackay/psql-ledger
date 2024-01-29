package service

import (
	"bytes"
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
	cfg := emptyConfig

	c, defaultUsed := sanitizeConfig(cfg)
	if !defaultUsed {
		t.Errorf("unexpected result, expected 'defaultUsed' to be true")
	}

	b, _ := yaml.Marshal(c)
	e, _ := yaml.Marshal(DefaultConfig)
	if !bytes.Equal(b, e) {
		t.Errorf("returned config not equal to default")
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

func Test_APIGet(t *testing.T) {
	n := "test"
	log, err := NewLogger("fatal", "plain", false, n)
	if err != nil {
		t.Fatal(err)
	}
	s := New(8080, log, database.NewMemDB())

	apiTests := []struct {
		name             string
		endpoint         string
		method           func(w http.ResponseWriter, req *http.Request)
		expectedResponse any
		expectedCode     int
	}{
		{
			"status",
			Status,
			s.Status,
			&StatusResponse{Message: "OK", Version: FullVersion, Service: serviceName},
			http.StatusOK,
		},
		{
			"health",
			Health,
			s.Health,
			&HealthResponse{Version: FullVersion, Service: serviceName, Failures: []string{}},
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
