package logger

import (
	"fmt"
	"os"
)

type Logger struct {
	color bool
	debug bool
}

func New(color, debug bool) *Logger {
	return &Logger{color, debug}
}

func (l *Logger) Write(data string, args ...interface{}) {
	fmt.Printf(data, args...)
}

func (l *Logger) Debug(data string, args ...interface{}) {
	if l.debug {
		fmt.Printf("DBUG %s\n", fmt.Sprintf(data, args...))
	}
}

func (l *Logger) Info(data string, args ...interface{}) {
	if l.color {
		fmt.Printf("\033[34mINFO\033[0m %s\n", fmt.Sprintf(data, args...))
	} else {
		fmt.Printf("INFO %s\n", fmt.Sprintf(data, args...))
	}
}

func (l *Logger) Warn(data string, args ...interface{}) {
	if l.color {
		fmt.Printf("\033[33mWARN\033[0m %s\n", fmt.Sprintf(data, args...))
	} else {
		fmt.Printf("WARN %s\n", fmt.Sprintf(data, args...))
	}
}

func (l *Logger) Error(data string, args ...interface{}) {
	if l.color {
		fmt.Fprintf(os.Stderr, " \033[31mERR\033[0m %s\n", fmt.Sprintf(data, args...))
	} else {
		fmt.Fprintf(os.Stderr, " ERR %s\n", fmt.Sprintf(data, args...))
	}
}

func (l *Logger) WithError(err error) {
	if l.color {
		fmt.Fprintf(os.Stderr, " \033[31mERR\033[0m %s\n", err.Error())
	} else {
		fmt.Fprintf(os.Stderr, " ERR %s\n", err.Error())
	}
}
