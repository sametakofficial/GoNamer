package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.SugaredLogger

func init() {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var err error
	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	log = logger.Sugar()
}

func GetLogger() *zap.SugaredLogger {
	return log
}
