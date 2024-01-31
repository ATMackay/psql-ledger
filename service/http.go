package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type HTTPService struct {
	server *http.Server
	logger *logrus.Entry
}

func NewHTTPService(port int, api *API, l *logrus.Entry) HTTPService {

	handler := api.Routes()
	// User logging middleware
	handler.Use(logHTTPRequest(l))

	return HTTPService{
		server: &http.Server{
			Addr:              fmt.Sprintf(":%d", port),
			Handler:           handler,
			ReadHeaderTimeout: 20 * time.Second,
		},
		logger: l,
	}
}

func (h *HTTPService) Addr() string {
	return h.server.Addr
}

func (h *HTTPService) Start() {
	go func() {
		if err := h.server.ListenAndServe(); err != nil {
			h.logger.WithFields(logrus.Fields{"error": err}).Warn("serverTerminated")
		}
	}()
}

func (h *HTTPService) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return h.server.Shutdown(ctx)
}

type EndPoint struct {
	Path       string
	Handler    http.HandlerFunc
	MethodType string
}

func NewEndpoint(path, methodType string, handler http.HandlerFunc) EndPoint {
	return EndPoint{
		Path:       path,
		Handler:    handler,
		MethodType: methodType,
	}
}

type API struct {
	Endpoints []EndPoint
}

func MakeAPI(endpoints []EndPoint) *API {
	r := &API{}
	for _, e := range endpoints {
		r.AddEndpoint(e)
	}
	return r
}

func (a *API) AddEndpoint(e EndPoint) {
	a.Endpoints = append(a.Endpoints, e)
}

func (a *API) Routes() *mux.Router {
	router := mux.NewRouter()
	for _, e := range a.Endpoints {
		router.Handle(e.Path, e.Handler).Methods(e.MethodType)
	}
	return router
}

// HTTP logging middleware

// logHTTPRequest surfaces low level request/response data from the http server.
func logHTTPRequest(entry *logrus.Entry) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		entry := entry
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if entry == nil {
				return
			}
			start := time.Now()
			body, err := readBody(req)
			if err != nil {
				entry.WithError(err)
			}
			statusRecorder := &responseRecorder{ResponseWriter: w}
			h.ServeHTTP(statusRecorder, req)
			elapsed := time.Since(start)
			httpCode := statusRecorder.statusCode
			entry = entry.WithFields(logrus.Fields{
				"http_route":           req.URL.Path,
				"http_method":          req.Method,
				"http_code":            httpCode,
				"elapsed_microseconds": elapsed.Microseconds(),
			})
			// only log full request/reposne data if running in debug mode
			if entry.Logger.Level >= logrus.DebugLevel {
				entry = entry.WithField("body", body)
				entry = entry.WithField("response", string(statusRecorder.response))
			}
			if httpCode > 399 {
				entry.Warn()
			} else {
				entry.Print()
			}
		})
	}
}

type responseRecorder struct {
	http.ResponseWriter

	statusCode int
	response   []byte
}

func (w *responseRecorder) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseRecorder) Write(b []byte) (int, error) {
	w.response = b
	return w.ResponseWriter.Write(b)
}

func readBody(r *http.Request) (map[string]interface{}, error) {
	body := make(map[string]interface{})
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &body); err != nil {
		return nil, err
	}
	defer func() {
		r.Body.Close()
		r.Body = io.NopCloser(bytes.NewBuffer(b))
		r.ContentLength = int64(bytes.NewBuffer(b).Len())
	}()
	return body, nil
}

func RespondWithJSON(w http.ResponseWriter, code int, payload any) error {
	response, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(response)
	return err
}

func DecodeJSON(r io.Reader, v any) error {
	return json.NewDecoder(r).Decode(v)
}

func RespondWithError(w http.ResponseWriter, code int, msg any) {
	var message string
	switch m := msg.(type) {
	case error:
		message = m.Error()
	case string:
		message = m
	}
	_ = RespondWithJSON(w, code, map[string]string{"error": message})
}

func HandleResponseErr(resp *http.Response) error {
	if resp.StatusCode != 200 {
		var v jsonErr
		if err := DecodeJSON(resp.Body, &v); err != nil {
			return fmt.Errorf("cannot parse JSON body from error response: %w", err)
		}
		return fmt.Errorf(v.Err)
	}
	return nil
}

type jsonErr struct {
	Err string `json:"error"`
}
