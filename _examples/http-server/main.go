package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Traceableai/goagent"
	"github.com/Traceableai/goagent/config"
	"github.com/Traceableai/goagent/instrumentation/net/traceablehttp"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	port = ":8081"
)

var logger *zap.Logger

func main() {
	cfgFilePath, found := os.LookupEnv("TA_CONFIG_FILE")
	if !found {
		cfgFilePath = "./config.yaml"
	}

	cfg := config.LoadFromFile(filepath.Clean(cfgFilePath))

	closer := goagent.Init(cfg)
	defer closer()
	logger, _ = zap.NewProduction()
	logger = zap.New(zapcore.NewTee(logger.Core(), goagent.NewZapCore("http-server", cfg.Tracing.GetTelemetry().GetLogs())))
	r := mux.NewRouter()
	r.Handle("/foo", traceablehttp.NewHandler(
		http.HandlerFunc(fooHandler),
		"/foo",
	))
	// Using log.Fatal(http.ListenAndServe(":8081", r)) causes a gosec timeout error.
	// G114 (CWE-676): Use of net/http serve function that has no support for setting timeouts (Confidence: HIGH, Severity: MEDIUM)
	srv := http.Server{
		Addr:              port,
		Handler:           r,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		ReadHeaderTimeout: 60 * time.Second,
	}
	log.Println("Starting HTTP server on " + port)
	log.Fatal(srv.ListenAndServe())
}

type person struct {
	Name string `json:"name"`
}

func fooHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info("Received foo request")
	sBody, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	p := &person{}
	err = json.Unmarshal(sBody, p)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	<-time.After(300 * time.Millisecond)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(fmt.Sprintf("{\"message\": \"Hello %s\"}", p.Name)))
}
