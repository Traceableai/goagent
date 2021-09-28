package logger

import (
	"sync"

	"go.uber.org/zap"
)

var (
	_globalMu sync.RWMutex
	_globalL  = zap.NewNop()
)

func Logger() *zap.Logger {
	_globalMu.RLock()
	defer _globalMu.RUnlock()
	return _globalL
}

func InitLogger(l *zap.Logger) {
	_globalMu.Lock()
	defer _globalMu.Unlock()
	_globalL = l
}
