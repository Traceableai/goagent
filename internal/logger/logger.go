package logger

import (
	"log"
	"strings"
	"sync"

	"go.uber.org/zap"
)

var (
	_lMux sync.RWMutex
	_l    *zap.Logger
)

func GetLogger() *zap.Logger {
	if _l == nil {
		setLogger(zap.NewNop())
	}

	return _l
}

func setLogger(l *zap.Logger) {
	_lMux.Lock()
	defer _lMux.Unlock()

	if _l != nil {
		log.Println("logger already initialized, ignoring new logger.")
		return
	}

	_l = l
}

func InitLogger(logLevel string) func() {
	var lvl = zap.ErrorLevel
	switch strings.ToLower(logLevel) {
	case "debug":
		lvl = zap.DebugLevel
	case "info":
		lvl = zap.InfoLevel
	case "warn":
		lvl = zap.WarnLevel
	}

	l, err := zap.Config{
		Level:            zap.NewAtomicLevelAt(lvl),
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}.Build()
	if err != nil {
		log.Printf("Failed to init logger: %v", err)
		return func() {}
	}

	setLogger(l)

	return func() {
		if err := l.Sync(); err != nil {
			log.Printf("Failed sync logger: %v", err)
		}
	}
}
