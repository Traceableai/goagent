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
	log.Fatal(http.ListenAndServe(":8081", r))
}

func fooHandler(w http.ResponseWriter, r *http.Request) {
	<-time.After(300 * time.Millisecond)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"message\": \"Hello world\"}"))
}
