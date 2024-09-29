package logging

import (
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// GetDefaultFileAndConsoleLogger configures and creates a logr.Logger instance
func GetDefaultFileAndConsoleLogger(filePath string, jsonEncoding bool) (*zap.Logger, logr.Logger, error) {
	zapConfig := zap.NewProductionConfig()
	zapConfig.Sampling = nil

	zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	zapConfig.OutputPaths = []string{"stderr", filePath}
	zapConfig.ErrorOutputPaths = []string{"stderr", filePath}
	zapConfig.DisableStacktrace = false

	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	zapConfig.EncoderConfig.TimeKey = "timestamp"
	zapConfig.EncoderConfig.LevelKey = "level"
	zapConfig.EncoderConfig.MessageKey = "message"

	if !jsonEncoding {
		zapConfig.Encoding = "console"
		zapConfig.EncoderConfig.ConsoleSeparator = "    "
		zapConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(
			time.StampMilli)
	}

	backendLogger, err := zapConfig.Build()
	if err != nil {
		return nil, logr.Logger{}, err
	}

	return backendLogger, zapr.NewLogger(backendLogger), nil
}
