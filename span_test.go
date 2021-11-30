package goagent

import (
	"context"
	"testing"
	"time"

	"github.com/Traceableai/goagent/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartSpan(t *testing.T) {
	cfg := config.Load()
	shutdown := Init(cfg)
	defer shutdown()

	_, s, ender := StartSpan(
		context.Background(),
		"test",
		WithSpanKind(SpanKindClient),
		WithTimestamp(time.Now()),
	)
	defer ender()

	assert.False(t, s.IsNoop())
}

func TestRegisterService(t *testing.T) {
	cfg := config.Load()
	shutdown := Init(cfg)
	defer shutdown()
	ss, err := RegisterService("myservice", nil)
	require.NoError(t, err)

	_, _, closer := ss(context.Background(), "test")
	closer()
}
