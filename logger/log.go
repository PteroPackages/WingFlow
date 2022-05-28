package logger

import (
	"fmt"
	"os"
)

type Logger struct {
	debug bool
	color bool
}

func New(color string, debug bool) *Logger {
	return &Logger{
		color: color == "false",
		debug: debug,
	}
}

func (l *Logger) Info(data string) {
	if l.color {
		fmt.Printf("\x1b[34mINFO\x1b[0m: %s\n", data)
	} else {
		fmt.Printf("INFO: %s\n", data)
	}
}

func (l *Logger) Debug(data string) {
	if l.debug {
		fmt.Printf("DBUG: %s\n", data)
	}
}

func (l *Logger) Warn(data string, args ...interface{}) {
	if l.color {
		fmt.Printf("\x1b[33mWARN\x1b[0m: %s\n", fmt.Sprintf(data, args...))
	} else {
		fmt.Printf("WARN: %s\n", fmt.Sprintf(data, args...))
	}
}

func (l *Logger) Error(data string, args ...interface{}) {
	if l.color {
		fmt.Fprintf(os.Stderr, " \x1b[31mERR\x1b[0m: %s\n", fmt.Sprintf(data, args...))
	} else {
		fmt.Fprintf(os.Stderr, " ERR: %s\n", fmt.Sprintf(data, args...))
	}
}

func (l *Logger) WithError(err error) {
	if l.color {
		fmt.Fprintf(os.Stderr, " \x1b[31mERR\x1b[0m: %v\n", err.Error())
	} else {
		fmt.Fprintf(os.Stderr, " ERR: %v\n", err.Error())
	}
}

func (l *Logger) Fatal(data string) {
	if l.color {
		fmt.Fprintf(os.Stderr, "\x1b[31mFATL\x1b[0m: %s\n", data)
	} else {
		fmt.Fprintf(os.Stderr, "FATL: %s\n", data)
	}
	os.Exit(1)
}

func (l *Logger) WithFatal(err error) {
	if l.color {
		fmt.Fprintf(os.Stderr, "\x1b[31mFATL\x1b[0m: %v\n", err.Error())
	} else {
		fmt.Fprintf(os.Stderr, "FATL: %v\n", err.Error())
	}
	os.Exit(1)
}
