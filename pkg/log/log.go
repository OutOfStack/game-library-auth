package log

import (
	"os"

	"github.com/OutOfStack/game-library-auth/internal/appconf"
	gelf "github.com/snovichkov/zap-gelf"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New returns new zap logger instance
func New(logLvl string, graylogAddress string) *zap.Logger {
	// log level
	logLevel := zap.InfoLevel
	parsedLevel, lErr := zapcore.ParseLevel(logLvl)
	if lErr == nil {
		logLevel = parsedLevel
	}

	// log format
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.RFC3339TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder

	// set console output
	consoleWriter := zapcore.Lock(os.Stderr)
	cores := []zapcore.Core{
		zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), consoleWriter, logLevel),
	}

	// set Graylog output
	host, _ := os.Hostname()
	gelfCore, gelfErr := gelf.NewCore(gelf.Addr(graylogAddress), gelf.Host(host), gelf.Level(logLevel))
	if gelfCore != nil {
		cores = append(cores, gelfCore)
	}

	core := zapcore.NewTee(cores...)

	logger := zap.New(core, zap.WithCaller(false)).With(zap.String("service", appconf.ServiceName))

	// log deferred errors if any
	if lErr != nil {
		logger.Error("parse log level", zap.Error(lErr), zap.String("level", logLvl))
	}
	if gelfErr != nil {
		logger.Error("create gelf logger", zap.Error(gelfErr))
	}

	return logger
}
