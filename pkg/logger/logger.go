package logger

import (
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log = zap.NewNop()

func Init(level, format string) error {
	var logLevel zapcore.Level
	if err := logLevel.UnmarshalText([]byte(strings.ToLower(level))); err != nil {
		return fmt.Errorf("parse log level: %w", err)
	}
	built, err := buildLogger(logLevel, strings.ToLower(format), zapcore.Lock(os.Stdout), zapcore.Lock(os.Stderr))
	if err != nil {
		return fmt.Errorf("build logger: %w", err)
	}
	previous := log
	log = built
	_ = previous.Sync()
	return nil
}

func buildLogger(level zapcore.Level, format string, output, errorOutput zapcore.WriteSyncer) (*zap.Logger, error) {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:          "time",
		LevelKey:         "level",
		CallerKey:        "caller",
		MessageKey:       "message",
		StacktraceKey:    "stacktrace",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeDuration:   zapcore.StringDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		ConsoleSeparator: " ",
	}
	newEncoder := func() (zapcore.Encoder, error) {
		switch format {
		case "console":
			encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
			encoderConfig.EncodeTime = consoleTimeEncoder
			return zapcore.NewConsoleEncoder(encoderConfig), nil
		case "json":
			encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
			encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
			return zapcore.NewJSONEncoder(encoderConfig), nil
		default:
			return nil, fmt.Errorf("unsupported log format %q", format)
		}
	}
	standardEncoder, err := newEncoder()
	if err != nil {
		return nil, err
	}
	errorEncoder, err := newEncoder()
	if err != nil {
		return nil, err
	}
	standardPriority := zap.LevelEnablerFunc(func(current zapcore.Level) bool {
		return current >= level && current < zapcore.ErrorLevel
	})
	errorPriority := zap.LevelEnablerFunc(func(current zapcore.Level) bool {
		return current >= level && current >= zapcore.ErrorLevel
	})
	core := zapcore.NewTee(
		zapcore.NewCore(standardEncoder, output, standardPriority),
		zapcore.NewCore(errorEncoder, errorOutput, errorPriority),
	)
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)).With(zap.String("service", "ai-argus")), nil
}

func consoleTimeEncoder(value time.Time, encoder zapcore.PrimitiveArrayEncoder) {
	encoder.AppendString(value.Local().Format("2006-01-02 15:04:05.000"))
}

func L() *zap.Logger {
	return log
}

func Sync() {
	_ = log.Sync()
}
