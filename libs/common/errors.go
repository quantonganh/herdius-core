package common

import (
	"fmt"
	"runtime"
)

//Error ...
type Error interface {
	Error() string
	Stacktrace() Error
	Trace(offset int, format string, args ...interface{}) Error
	Data() interface{}
}

// Captures a stacktrace if one was not already captured.
func (err *cmnError) Stacktrace() Error {
	if err.stacktrace == nil {
		var offset = 3
		var depth = 32
		err.stacktrace = captureStacktrace(offset, depth)
	}
	return err
}

// Implements error.
func (err *cmnError) Error() string {
	return fmt.Sprintf("%v", err)
}

// Add tracing information with msg.
// Set n=0 unless wrapped with some function, then n > 0.
func (err *cmnError) Trace(offset int, format string, args ...interface{}) Error {
	msg := fmt.Sprintf(format, args...)
	return err.doTrace(msg, offset)
}

func (err *cmnError) doTrace(msg string, n int) Error {
	pc, _, _, _ := runtime.Caller(n + 2) // +1 for doTrace().  +1 for the caller.
	// Include file & line number & msg.
	// Do not include the whole stack trace.
	err.msgtraces = append(err.msgtraces, msgtraceItem{
		pc:  pc,
		msg: msg,
	})
	return err
}

// Return the "data" of this error.
// Data could be used for error handling/switching,
// or for holding general error/debug information.
func (err *cmnError) Data() interface{} {
	return err.data
}

//----------------------------------------
// stacktrace & msgtraceItem

func captureStacktrace(offset int, depth int) []uintptr {
	var pcs = make([]uintptr, depth)
	n := runtime.Callers(offset, pcs)
	return pcs[0:n]
}

type msgtraceItem struct {
	pc  uintptr
	msg string
}

type FmtError struct {
	format string
	args   []interface{}
}

// New Error with formatted message.
// The Error's Data will be a FmtError type.
func NewError(format string, args ...interface{}) Error {
	err := FmtError{format, args}
	return newCmnError(err)
}

type cmnError struct {
	data       interface{}    // associated data
	msgtraces  []msgtraceItem // all messages traced
	stacktrace []uintptr      // first stack trace
}

// NOTE: do not expose.
func newCmnError(data interface{}) *cmnError {
	return &cmnError{
		data:       data,
		msgtraces:  nil,
		stacktrace: nil,
	}
}

// PanicCrisis ...
// A panic here means something has gone horribly wrong, in the form of data corruption or
// failure of the operating system. In a correct/healthy system, these should never fire.
// If they do, it's indicative of a much more serious problem.
func PanicCrisis(v interface{}) {
	panic(fmt.Sprintf("Panicked on a Crisis: %v", v))
}
