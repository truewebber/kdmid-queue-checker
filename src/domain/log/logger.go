package log

import "io"

type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	CloseWithLog(closer io.Closer)
}
