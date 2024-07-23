package zap

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"testing"

	"go-micro.dev/v5/logger"
)

func TestName(t *testing.T) {
	l, err := NewLogger()
	if err != nil {
		t.Fatal(err)
	}

	if l.String() != "zap" {
		t.Errorf("name is error %s", l.String())
	}

	t.Logf("test logger name: %s", l.String())
}

func TestLogf(t *testing.T) {
	// skip is 2, because we call logger through logger package
	l, err := NewLogger(logger.WithCallerSkipCount(2))
	if err != nil {
		t.Fatal(err)
	}

	logger.DefaultLogger = l
	logger.Logf(logger.InfoLevel, "test logf: %s", "name")
}

func TestSetLevel(t *testing.T) {
	// skip is 1, because we call logger directly
	l, err := NewLogger(logger.WithCallerSkipCount(1))
	if err != nil {
		t.Fatal(err)
	}
	logger.DefaultLogger = l

	logger.Init(logger.WithLevel(logger.DebugLevel))
	l.Logf(logger.DebugLevel, "test show debug: %s", "debug msg")

	logger.Init(logger.WithLevel(logger.InfoLevel))
	l.Logf(logger.DebugLevel, "test non-show debug: %s", "debug msg")
}

func TestZapLogger(t *testing.T) {
	enc := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	w := zapcore.AddSync(os.Stdout)
	core := zapcore.NewCore(enc, w, zapcore.ErrorLevel)
	zapLogger := zap.New(core)
	l, err := NewLogger(logger.WithLevel(logger.InfoLevel), WithLogger(zapLogger))
	if err != nil {
		t.Fatal(err)
	}
	l.Logf(logger.InfoLevel, "test non-show info: %s", "info msg")
	l.Logf(logger.ErrorLevel, "test show error: %s", "error msg")
	logger.Init(logger.WithLevel(logger.InfoLevel))
	l.Logf(logger.InfoLevel, "test non-show info: %s", "info msg")
}
