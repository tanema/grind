package runner

import (
	"fmt"
	"io"
)

// Logger is a simple logger for prefixing outputs
type Logger struct {
	prefix string
	writer io.Writer
}

func (w *Logger) Printf(msg string, args ...any) (int, error) {
	return w.Write([]byte(fmt.Sprintf(msg, args...)))
}

func (w *Logger) Write(b []byte) (int, error) {
	_, err := w.writer.Write(append([]byte(w.prefix), b...))
	return len(b), err
}
