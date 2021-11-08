package logger

import (
	"log"
	"sync"

	"go.uber.org/zap"
)

var (
	_lMux sync.RWMutex
	_l    *zap.Logger
)

func GetLogger() *zap.Logger {
	if _l == nil {
		InitLogger(zap.NewNop())
	}

	return _l
}

func InitLogger(l *zap.Logger) {
	_lMux.Lock()
	defer _lMux.Unlock()

	if _l != nil {
		log.Println("logger already initialized, ignoring new logger.")
		return
	}

	_l = l
}
