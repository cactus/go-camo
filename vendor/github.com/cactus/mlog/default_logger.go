package mlog

import (
	"fmt"
	"os"
)

// DefaultLogger is the default package level Logger
var DefaultLogger = New(os.Stderr, Lstd)

// SetEmitter sets the Emitter for the degault logger. See
// Logger.SetEmitter.
func SetEmitter(e Emitter) {
	DefaultLogger.SetEmitter(e)
}

// Flags returns the FlagSet of the default Logger. See Logger.Flags.
func Flags() FlagSet {
	return DefaultLogger.Flags()
}

// SetFlags sets the FlagSet on the default Logger. See Logger.SetFlags.
func SetFlags(flags FlagSet) {
	DefaultLogger.SetFlags(flags)
}

// Debugm logs to the default Logger. See Logger.Debugm
func Debugm(message string, v Map) {
	if DefaultLogger.HasDebug() {
		DefaultLogger.Emit(-1, message, v)
	}
}

// Infom logs to the default Logger. See Logger.Infom
func Infom(message string, v Map) {
	DefaultLogger.Emit(0, message, v)
}

// Printm logs to the default Logger. See Logger.Printm
func Printm(message string, v Map) {
	DefaultLogger.Emit(0, message, v)
}

// Fatalm logs to the default Logger. See Logger.Fatalm
func Fatalm(message string, v Map) {
	DefaultLogger.Emit(1, message, v)
	os.Exit(1)
}

// Debugf logs to the default Logger. See Logger.Debugf
func Debugf(format string, v ...interface{}) {
	if DefaultLogger.HasDebug() {
		DefaultLogger.Emit(-1, fmt.Sprintf(format, v...), nil)
	}
}

// Infof logs to the default Logger. See Logger.Infof
func Infof(format string, v ...interface{}) {
	DefaultLogger.Emit(0, fmt.Sprintf(format, v...), nil)
}

// Printf logs to the default Logger. See Logger.Printf
func Printf(format string, v ...interface{}) {
	DefaultLogger.Emit(0, fmt.Sprintf(format, v...), nil)
}

// Fatalf logs to the default Logger. See Logger.Fatalf
func Fatalf(format string, v ...interface{}) {
	DefaultLogger.Emit(1, fmt.Sprintf(format, v...), nil)
	os.Exit(1)
}

// Debug logs to the default Logger. See Logger.Debug
func Debug(v ...interface{}) {
	if DefaultLogger.HasDebug() {
		DefaultLogger.Emit(-1, fmt.Sprint(v...), nil)
	}
}

// Info logs to the default Logger. See Logger.Info
func Info(v ...interface{}) {
	DefaultLogger.Emit(0, fmt.Sprint(v...), nil)
}

// Print logs to the default Logger. See Logger.Print
func Print(v ...interface{}) {
	DefaultLogger.Emit(0, fmt.Sprint(v...), nil)
}

// Fatal logs to the default Logger. See Logger.Fatal
func Fatal(v ...interface{}) {
	DefaultLogger.Emit(1, fmt.Sprint(v...), nil)
	os.Exit(1)
}
