package main

import (
	"log"
	"net/http"

	"github.com/Traceableai/goagent/hypertrace/goagent/config"
	"github.com/Traceableai/goagent/hypertrace/goagent/instrumentation/hypertrace"
	"github.com/Traceableai/goagent/hypertrace/goagent/instrumentation/hypertrace/github.com/gin-gonic/hypergin"
	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	// Disable Console Color
	// hypergin.DisableConsoleColor()
	r := gin.Default()
	db := make(map[string]string)
	db["john"] = "doe"
	db["jane"] = "smith"
	db["bob"] = "builder"

	r.Use(hypergin.Middleware())

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.Header("ping-response-header-1", "ping")
		c.Writer.Header().Set("ping-response-header-2", "ping")
		c.JSON(200, gin.H{
			"code":    http.StatusOK,
			"message": "pong",
		})
	})

	// Get user value
	r.GET("/user/:name", func(c *gin.Context) {
		name := c.Params.ByName("name")
		value, ok := db[name]
		if ok {
			c.Header("user-response-header-1", "valexists")
			c.Writer.Header().Set("user-response-header-2", "valexists")
			c.JSON(http.StatusOK, gin.H{"user": name, "value": value})
		} else {
			c.Header("user-response-header-1", "novalue")
			c.Writer.Header().Set("user-response-header-2", "novalue")
			c.JSON(http.StatusNotFound, gin.H{"user": name, "status": "no value"})
		}
	})

	return r
}

func main() {
	cfg := config.Load()
	cfg.ServiceName = config.String("gin-example-server")
	cfg.Reporting.Endpoint = config.String("localhost:5442")
	cfg.Reporting.TraceReporterType = config.TraceReporterType_OTLP

	flusher := hypertrace.Init(cfg)
	defer flusher()

	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	err := r.Run(":8080")
	if err != nil {
		log.Fatalf("gin server failed with error: %v", err)
	}
}
