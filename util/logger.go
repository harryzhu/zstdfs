package util

import (
	"fmt"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Zapper struct {
	Level  string
	File   string
	Logger *zap.Logger
}

var GlobalZapper *Zapper
var once sync.Once

func (z *Zapper) GetLogger() *Zapper {
	once.Do(func() {
		var logLevel zapcore.Level
		switch strings.ToLower(z.Level) {
		case "debug":
			logLevel = zapcore.DebugLevel
		case "info":
			logLevel = zapcore.InfoLevel
		case "warn":
			logLevel = zapcore.WarnLevel
		case "error":
			logLevel = zapcore.ErrorLevel
		case "fatal":
			logLevel = zapcore.FatalLevel
		default:
			logLevel = zapcore.DebugLevel

		}
		z.Logger = SetZapLogger(z.File, logLevel)
	})
	GlobalZapper = z
	return z
}

func SetZapLogger(LogFile string, LogLevel zapcore.Level) *zap.Logger {
	ec := zap.NewProductionEncoderConfig()
	ec.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")

	cfgProduction := zap.NewProductionConfig()
	cfgProduction.EncoderConfig = ec
	cfgProduction.DisableStacktrace = true
	cfgProduction.DisableCaller = true
	cfgProduction.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)

	if LogFile != "" {
		cfgProduction.OutputPaths = []string{LogFile, "stdout"}
		cfgProduction.ErrorOutputPaths = []string{LogFile, "stderr"}
	} else {
		cfgProduction.OutputPaths = []string{"stdout"}
		cfgProduction.ErrorOutputPaths = []string{"stderr"}
	}

	cfgProduction.InitialFields = map[string]interface{}{}

	zl, err := cfgProduction.Build()
	defer zl.Sync()

	if err != nil {
		panic(fmt.Sprintf("logger init failed: %v", err))
	}
	return zl
}
