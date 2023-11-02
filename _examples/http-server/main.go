package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/Traceableai/goagent"
	"github.com/Traceableai/goagent/config"
	"github.com/Traceableai/goagent/instrumentation/net/traceablehttp"
	"github.com/gorilla/mux"
)

const (
	port = ":8081"
)

func main() {
	cfg := config.LoadFromFile("./config.yaml")

	closer := goagent.Init(cfg)
	defer closer()

	r := mux.NewRouter()
	r.Handle("/foo", traceablehttp.NewHandler(
		http.HandlerFunc(fooHandler),
		"/foo",
	))
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

type person struct {
	Name string `json:"name"`
}

func fooHandler(w http.ResponseWriter, r *http.Request) {
	sBody, err := ioutil.ReadAll(r.Body)
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
