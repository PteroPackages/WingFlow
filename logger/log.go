package logger

import (
	"fmt"
	"os"
)

type Logger struct {
	debug bool
}

func New(debug bool) *Logger {
	return &Logger{debug: debug}
}

func (*Logger) Info(data string) {
	fmt.Printf("INFO: %s\n", data)
}

func (l *Logger) Debug(data string) {
	if l.debug {
		fmt.Printf("DBUG: %s\n", data)
	}
}

func (*Logger) Warn(data string) {
	fmt.Printf("WARN: %s\n", data)
}

func (*Logger) Error(data string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, " ERR: %s\n", fmt.Sprintf(data, args...))
}

func (*Logger) WithError(err error) {
	fmt.Fprintf(os.Stderr, " ERR: %v\n", err.Error())
}

func (*Logger) Fatal(data string) {
	fmt.Fprintf(os.Stderr, "FATL: %s\n", data)
	os.Exit(1)
}

func (*Logger) WithFatal(err error) {
	fmt.Fprintf(os.Stderr, "FATL: %v\n", err.Error())
	os.Exit(1)
}
