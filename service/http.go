package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

type HTTPService struct {
	server *http.Server
}

func NewHTTPService(port int, api *API) HTTPService {

	handler := api.Routes()

	return HTTPService{
		server: &http.Server{
			Addr:              fmt.Sprintf(":%d", port),
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
		},
	}
}

func (h *HTTPService) Addr() string {
	return h.server.Addr
}

func (h *HTTPService) Start() {
	go func() {
		slog.Info(fmt.Sprintf("server listening on http://0.0.0.0%v", h.Addr()))
		if err := h.server.ListenAndServe(); err != nil {
			slog.Warn("serverTerminated", "error", err)
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

func (a *API) Routes() *httprouter.Router {

	router := httprouter.New()

	for _, e := range a.Endpoints {

		router.Handler(e.MethodType, e.Path, logHTTPRequest(e.Handler))

	}
	return router
}

// HTTP logging middleware

// logHTTPRequest provides logging middleware. It surfaces low level request/response data from the http server.
func logHTTPRequest(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		statusRecorder := &responseRecorder{ResponseWriter: w}
		h.ServeHTTP(statusRecorder, req)
		elapsed := time.Since(start)
		httpCode := statusRecorder.statusCode
		if httpCode > 499 {
			slog.Error(req.URL.Path, "http_method", req.Method,
				"http_code", httpCode,
				"elapsed_microseconds", elapsed.Microseconds())
			return
		}
		if httpCode > 399 {
			slog.Warn(req.URL.Path, "http_method", req.Method,
				"http_code", httpCode,
				"elapsed_microseconds", elapsed.Microseconds())
			return
		}
		slog.Info(req.URL.Path, "http_method", req.Method,
			"http_code", httpCode,
			"elapsed_microseconds", elapsed.Microseconds())
	})
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
		return errors.New(v.Err)
	}
	return nil
}

type jsonErr struct {
	Err string `json:"error"`
}
