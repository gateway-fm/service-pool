package logger

// TODO Move Logger singleton to shared packages

import (
	"sync"

	"go.uber.org/zap"
)

type Zaplog struct {
	*zap.Logger
}

var instance *Zaplog
var once sync.Once

//Log is invoking Zap Logger function
func Log() *Zaplog {
	once.Do(func() {
		logger, _ := zap.NewProduction()
		instance = &Zaplog{logger}
	})
	return instance
}
