package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Traceableai/goagent/config"
	"github.com/Traceableai/goagent/instrumentation/github.com/gorilla/traceablemux"

	"github.com/Traceableai/goagent"
	"github.com/gorilla/mux"
)

func main() {
	cfg := config.Load()
	cfg.Tracing.ServiceName = config.String("http-mux-server")

	flusher := goagent.Init(cfg)
	defer flusher()

	r := mux.NewRouter()
	r.Use(traceablemux.NewMiddleware()) // here we use the mux middleware
	r.HandleFunc("/foo", http.HandlerFunc(fooHandler))
	// Using log.Fatal(http.ListenAndServe(":8081", r)) causes a gosec timeout error.
	// G114 (CWE-676): Use of net/http serve function that has no support for setting timeouts (Confidence: HIGH, Severity: MEDIUM)
	srv := http.Server{
		Addr:              ":8081",
		Handler:           r,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		ReadHeaderTimeout: 60 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

func fooHandler(w http.ResponseWriter, r *http.Request) {
	<-time.After(300 * time.Millisecond)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("{\"message\": \"Hello world\"}"))
}
