package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/Traceableai/goagent"
	"github.com/Traceableai/goagent/config"
	"github.com/Traceableai/goagent/instrumentation/net/traceablehttp"
)

type message struct {
	Content string `json:"message"`
}

func main() {
	cfg := config.LoadFromFile("./config.yaml")

	closer := goagent.Init(cfg)
	defer closer()

	client := http.Client{
		Transport: traceablehttp.NewTransport(http.DefaultTransport),
	}

	req, err := http.NewRequest("GET", "http://localhost:8081/foo", bytes.NewBufferString(`{"name":"Dave"}`))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		log.Fatalf("failed to create the request: %v", err)
	}

	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("failed to perform the request: %v", err)
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("failed to read the response body: %v", err)
	}
	defer res.Body.Close()

	m := &message{}
	err = json.Unmarshal(resBody, m)
	if err != nil {
		log.Fatalf("failed to unmarshal the response body: %v", err)
		return
	}

	fmt.Println(m.Content)
}
