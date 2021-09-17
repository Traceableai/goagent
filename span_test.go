package goagent

import (
	"context"
	"testing"

	"github.com/Traceableai/goagent/config"
	"github.com/stretchr/testify/assert"
)

func TestStartSpan(t *testing.T) {
	cfg := config.Load()
	shutdown := Init(cfg)
	defer shutdown()

	_, s, ender := StartSpan(context.Background(), "test")
	defer ender()

	assert.False(t, s.IsNoop())
}
