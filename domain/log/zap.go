package log

import (
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	ZapWrapper struct {
		logger *zap.SugaredLogger
	}
)

func NewZapWrapper() *ZapWrapper {
	return &ZapWrapper{
		logger: newZap(),
	}
}

func newZap() *zap.SugaredLogger {
	encoderCfg := zapcore.EncoderConfig{
		MessageKey:       "_m",
		NameKey:          "logger",
		LevelKey:         "_l",
		EncodeLevel:      zapcore.LowercaseLevelEncoder,
		TimeKey:          "_t",
		EncodeTime:       zapcore.ISO8601TimeEncoder,
		CallerKey:        "",
		FunctionKey:      "",
		StacktraceKey:    "",
		LineEnding:       "",
		EncodeDuration:   func(duration time.Duration, encoder zapcore.PrimitiveArrayEncoder) {},
		EncodeCaller:     func(caller zapcore.EntryCaller, encoder zapcore.PrimitiveArrayEncoder) {},
		EncodeName:       func(s string, encoder zapcore.PrimitiveArrayEncoder) {},
		ConsoleSeparator: "",
	}
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), os.Stdout, zap.DebugLevel)

	return zap.New(core).Sugar()
}

func (z *ZapWrapper) Info(msg string, args ...interface{}) {
	z.logger.Infow(msg, args...)
}

func (z *ZapWrapper) Error(msg string, args ...interface{}) {
	z.logger.Errorw(msg, args...)
}

func (z *ZapWrapper) CloseWithLog(closer io.Closer) {
	if closer == nil {
		return
	}

	if err := closer.Close(); err != nil {
		z.logger.Infow("failed to close", "err", err)
	}
}

func (z *ZapWrapper) Close() error {
	err := z.logger.Sync()
	if err == nil {
		return nil
	}

	if isSyncInvalidError(err) {
		return nil
	}

	return fmt.Errorf("sync zap logger: %w", err)
}

func isSyncInvalidError(err error) bool {
	var pathErr *os.PathError

	if !errors.As(err, &pathErr) {
		return false
	}

	switch {
	case errors.Is(pathErr.Err, syscall.ENOTTY):
	case errors.Is(pathErr.Err, syscall.EINVAL):
	default:
		return false
	}

	return true
}
