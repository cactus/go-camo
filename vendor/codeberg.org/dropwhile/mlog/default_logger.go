package mlog

import (
	"fmt"
	"io"
	"os"
)

// DefaultLogger is the default package level Logger
var DefaultLogger = New(os.Stderr, Lstd)

func SetOutput(writer io.Writer) {
	DefaultLogger.SetOutput(writer)
}

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

// HasDebug returns true if the default Logger has debug logging FlagSet enabled.
// See Logger.HasDebug
func HasDebug() bool {
	return DefaultLogger.HasDebug()
}

// Debugx logs to the default Logger. See Logger.Debugm
func Debugx(message string, attrs ...*Attr) {
	if DefaultLogger.HasDebug() {
		DefaultLogger.EmitAttrs(-1, message, attrs...)
	}
}

// Infox logs to the default Logger. See Logger.Infom
func Infox(message string, attrs ...*Attr) {
	DefaultLogger.EmitAttrs(0, message, attrs...)
}

// Printx logs to the default Logger. See Logger.Printm
func Printx(message string, attrs ...*Attr) {
	DefaultLogger.EmitAttrs(0, message, attrs...)
}

// Fatalx logs to the default Logger. See Logger.Fatalm
func Fatalx(message string, attrs ...*Attr) {
	DefaultLogger.EmitAttrs(1, message, attrs...)
	os.Exit(1)
}

// Panicx logs to the default Logger. See Logger.Panicm
func Panicx(message string, attrs ...*Attr) {
	DefaultLogger.EmitAttrs(1, message, attrs...)
	panic(message)
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

// Panicm logs to the default Logger. See Logger.Panicm
func Panicm(message string, v Map) {
	DefaultLogger.Emit(1, message, v)
	panic(message)
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

// Panicf is equivalent to Printf() followed by a call to panic().
// See Logger.Panicf
func Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	DefaultLogger.Emit(1, s, nil)
	panic(s)
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

// Panic is equivalent to Print() followed by a call to panic().
// See Logger.Panic
func Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	DefaultLogger.Emit(1, s, nil)
	panic(s)
}
