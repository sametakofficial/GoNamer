package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.SugaredLogger
var config zap.Config

type CtxKey int

const loggerKey CtxKey = iota

func init() {
	config = zap.NewDevelopmentConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var err error
	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	log = logger.Sugar()
}

func InjectLogger(ctx context.Context, logger *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func FromContext(ctx context.Context) *zap.SugaredLogger {
	if contextLogger, ok := ctx.Value(loggerKey).(*zap.SugaredLogger); ok {
		return contextLogger
	}
	InjectLogger(ctx, log)
	return log
}

func SetLoggerLevel(level zapcore.Level) {
	config.Level = zap.NewAtomicLevelAt(level)
	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	log = logger.Sugar()
}

func SetLoggerOutput(output zapcore.WriteSyncer) {
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(config.EncoderConfig),
		output,
		config.Level,
	)
	logger := zap.New(core)
	log = logger.Sugar()
}
