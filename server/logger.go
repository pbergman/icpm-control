package main

import (
	"github.com/pbergman/logger"
	"os"
)

func GetLogger(config *Config) *logger.Logger {

	var handler = logger.NewWriterHandler(os.Stdout, logger.LogLevelDebug(), true)

	if false == config.Debug {
		handler = logger.NewThresholdHandler(handler, 15, logger.Error, true)
	}

	return logger.NewLogger("app", handler)
}
