// Package gogitlog implements a very simple wrapper around the
// Go "log" package, providing support for a toggle-able debug flag
// and a couple of functions that log or not based on that flag.
package gologit

import (
	"log"
	"os"
	"os/signal"
	"sync"
)

var Logger = New(false)

// A DebugLogger represents a logging object, that embeds log.Logger, and
// provides support for a toggle-able debug flag.
type DebugLogger struct {
	*log.Logger
	debug bool
	mx sync.Mutex
}

// New creates a new DebugLogger.
// The debug argument specifies whether debug should be set or not.
func New(debug bool) *DebugLogger {
	return &DebugLogger{log.New(os.Stderr, "", log.LstdFlags), debug, sync.Mutex{}}
}

// Toggles the debug state.
// If debug is true, sets it to false.
// If debug is false, sets it to true.
func (l *DebugLogger) Toggle() {
	l.mx.Lock()
	defer l.mx.Unlock()
	if l.debug == false {
		l.debug = true
	} else {
		l.debug = false
	}
}

func (l *DebugLogger) ToggleOnSignal(sig os.Signal) {
	debugSig := make(chan os.Signal, 1)
	// spawn goroutine to handle signal/toggle of debug logging
	go func() {
		for {
			<-debugSig
			l.Toggle()
			if l.State() {
				l.Printf("Debug logging enabled")
			} else {
				l.Printf("Debug logging disabled")
			}
		}
	}()
	// notify send to debug sign channel on signusr1
	signal.Notify(debugSig, sig)
}


func (l *DebugLogger) State() bool {
	return l.debug
}

// Set the debug state directly.
func (l *DebugLogger) Set(debug bool) {
	l.mx.Lock()
	defer l.mx.Unlock()
	l.debug = debug
}

// Debugf calls log.Printf if debug is true.
// If debug is false, does nothing.
func (l *DebugLogger) Debugf(format string, v ...interface{}) {
	if l.debug == true {
		l.Printf(format, v...)
	}
}

// Debug calls log.Print if debug is true.
// If debug is false, does nothing.
func (l *DebugLogger) Debug(v ...interface{}) {
	if l.debug == true {
		l.Print(v...)
	}
}

// Debugln calls log.Println if debug is true.
// If debug is false, does nothing.
func (l *DebugLogger) Debugln(v ...interface{}) {
	if l.debug == true {
		l.Println(v...)
	}
}


// These functions call the default Logger

// Toggles the debug state of the default Logger. See Logger.Toggle
func Toggle() {
	Logger.Toggle()
}

func ToggleOnSignal(sig os.Signal) {
	Logger.ToggleOnSignal(sig)
}

// Gets the state of the default Logger. See Logger.State
func State() bool {
	return Logger.State()
}

// Sets the state of the default Logger. See Logger.Set
func Set(debug bool) {
	Logger.Set(debug)
}

// Logs to the default Logger. See Logger.Debugf
func Debugf(format string, v ...interface{}) {
	Logger.Debugf(format, v...)
}

// Logs to the default Logger. See Logger.Debug
func Debug(v ...interface{}) {
	Logger.Debug(v...)
}

// Logs to the default Logger. See Logger.Debugln
func Debugln(v ...interface{}) {
	Logger.Debugln(v...)
}

// Logs to the default Logger. See Logger.Print
func Print(v ...interface{}) {
	Logger.Print(v...)
}

// Logs to the default Logger. See Logger.Printf
func Printf(format string, v ...interface{}) {
	Logger.Printf(format, v...)
}

// Logs to the default Logger. See Logger.Println
func Println(v ...interface{}) {
	Logger.Println(v...)
}

// Logs to the default Logger. See Logger.Fatal
func Fatal(v ...interface{}) {
	Logger.Fatal(v...)
}

// Logs to the default Logger. See Logger.Fatalf
func Fatalf(format string, v ...interface{}) {
	Logger.Fatalf(format, v...)
}

// Logs to the default Logger. See Logger.Fatalln
func Fatalln(v ...interface{}) {
	Logger.Fatalln(v...)
}

// Logs to the default Logger. See Logger.Panic
func Panic(v ...interface{}) {
	Logger.Panic(v...)
}

// Logs to the default Logger. See Logger.Panicf
func Panicf(format string, v ...interface{}) {
	Logger.Panicf(format, v...)
}

// Logs to the default Logger. See Logger.Panicln
func Panicln(v ...interface{}) {
	Logger.Panicln(v...)
}
