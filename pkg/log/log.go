package log

import (
	"log"
	"os"

	"github.com/OutOfStack/game-library-auth/internal/appconf"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/Graylog2/go-gelf.v2/gelf"
)

// InitLogger inits zap logger that writes to console and graylog.
// If graylog isn't available, writes to console only
func InitLogger(graylogAddress string) (*zap.Logger, error) {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.RFC3339TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder

	consoleWriter := zapcore.Lock(os.Stderr)
	cores := []zapcore.Core{
		zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), consoleWriter, zap.InfoLevel),
	}

	gelfWriter, err := gelf.NewTCPWriter(graylogAddress)
	if err != nil {
		log.Printf("can't create gelf writer: %v", err)
	}
	if gelfWriter != nil {
		cores = append(cores,
			zapcore.NewCore(
				zapcore.NewJSONEncoder(encoderCfg),
				zapcore.AddSync(gelfWriter),
				zap.InfoLevel))
	}

	core := zapcore.NewTee(cores...)

	logger := zap.New(core, zap.WithCaller(false)).With(zap.String("service", appconf.ServiceName))

	return logger, nil
}
