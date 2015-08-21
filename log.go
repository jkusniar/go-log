/*
Package log is another implementation of logger in golang. It is simple,
supports log levels and is thread safe. File writing synchronization is achieved
using channels. Struct fields thread-safety is achieved using locks. It is
intended to be used in a server application writing logs to a file. Log file
rotation is on a TODO list (using https://github.com/natefinch/lumberjack)
*/
package log

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

// Log level output indicator strings
const (
	labelDebug = "DEBUG"
	labelInfo  = "INFO"
	labelWarn  = "WARN"
	labelError = "ERROR"
)

// Supported log levels.
const (
	LevelError uint8 = iota
	LevelWarn
	LevelInfo
	LevelDebug
)

// Logger represents active logger object. It writes log entries to
// io.WriteCloser. Entries to be written are received trhough entries channel.
// Logger outputs only entries with level equal to, or lower than level of
// the Logger. Logger has to be Shutdown to properly close channels and writer.
type Logger struct {
	mutex    sync.Mutex     // mutex to sync access to other fields
	entries  chan logEntry  // channel for log entries to be written
	done     chan bool      // indicates all log entries were written
	writer   io.WriteCloser // output writer
	minLevel uint8          // minimal log level
}

// New creates new Logger object. Created logger is immediately active and can
// write output to w. Minimal log level is l. If w is nil, logger doesn't write
// anything but can be safely called from application.
func New(w io.WriteCloser, l uint8) *Logger {
	if l > LevelDebug {
		panic(fmt.Sprintf("Log level %v, but maximum allowed is %v",
			l, LevelDebug))
	}

	logger := &Logger{
		entries:  make(chan logEntry, 10),
		done:     make(chan bool),
		writer:   w,
		minLevel: l,
	}

	go logger.listen()

	return logger
}

// channel listening method. Run as goroutine asynchronously. Writes entries
// to l.writer.
func (l *Logger) listen() {
	for {
		entry, more := <-l.entries
		if more {
			fmt.Fprint(l.writer, entry)
		} else {
			l.done <- true
			return
		}
	}
}

// Shutdown closes logger. It closes entries channel, waits to for remaining
// entries to process and closes the writer. Proper usage is to defer Shutdown
// after Logger creation. On server applications, it is better to call shutdown
// from os.Signal handler.
func (l *Logger) Shutdown() {
	close(l.entries)
	<-l.done

	if l.writer != nil {
		l.writer.Close()
	}
}

// canLog checks, if entry with particular level can be written and if writer is
// not nil
func (l *Logger) canLog(level uint8) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	return l.minLevel >= level && l.writer != nil
}

// SetLevel sets new minimal log level for logger. If desired level is higher
// than maximum allowed, method does nothing (returns without warning)
func (l *Logger) SetLevel(level uint8) {
	if level > LevelDebug {
		return
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.minLevel = level
}

// Debug sends log entry with debug level to logger's entries channel.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Debug(v ...interface{}) {
	if l.canLog(LevelDebug) {
		l.entries <- createLogEntry(labelDebug, v...)
	}
}

// Debugf sends log entry with debug level to logger's entries channel.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.canLog(LevelDebug) {
		l.entries <- createLogEntryf(labelDebug, format, v...)
	}
}

// Info sends log entry with info level to logger's entries channel.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Info(v ...interface{}) {
	if l.canLog(LevelInfo) {
		l.entries <- createLogEntry(labelInfo, v...)
	}
}

// Infof sends log entry with info level to logger's entries channel.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Infof(format string, v ...interface{}) {
	if l.canLog(LevelInfo) {
		l.entries <- createLogEntryf(labelInfo, format, v...)
	}
}

// Warn sends log entry with warn level to logger's entries channel.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Warn(v ...interface{}) {
	if l.canLog(LevelWarn) {
		l.entries <- createLogEntry(labelWarn, v...)
	}
}

// Warn sends log entry with warn level to logger's entries channel.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Warnf(format string, v ...interface{}) {
	if l.canLog(LevelWarn) {
		l.entries <- createLogEntryf(labelWarn, format, v...)
	}
}

// Error sends log entry with error level to logger's entries channel.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Error(v ...interface{}) {
	if l.canLog(LevelError) {
		l.entries <- createLogEntry(labelError, v...)
	}
}

// Errorf sends log entry with error level to logger's entries channel.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Errorf(format string, v ...interface{}) {
	if l.canLog(LevelError) {
		l.entries <- createLogEntryf(labelError, format, v...)
	}
}

// Panic sends log entry with error level to logger's entries channel and
// calls panic() with entry's message. Arguments are handled in the manner
// of fmt.Print
func (l *Logger) Panic(v ...interface{}) {
	entry := createLogEntry(labelError, v...)
	if l.canLog(LevelError) {
		l.entries <- entry
	}

	panic(entry.Message)

}

// Panicf sends log entry with error level to logger's entries channel and
// calls panic() with entry's message. Arguments are handled in the manner
// of fmt.Printf.
func (l *Logger) Panicf(format string, v ...interface{}) {
	entry := createLogEntryf(labelError, format, v...)
	if l.canLog(LevelError) {
		l.entries <- entry
	}

	panic(entry.Message)
}

// Fatal sends log entry with error level to logger's entries channel and
// calls os.Exit(1). Arguments are handled in the manner of fmt.Print.
func (l *Logger) Fatal(v ...interface{}) {
	if l.canLog(LevelError) {
		l.entries <- createLogEntry(labelError, v...)
	}

	os.Exit(1)

}

// Fatalf sends log entry with error level to logger's entries channel and
// calls os.Exit(1). Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Fatalf(format string, v ...interface{}) {
	if l.canLog(LevelError) {
		l.entries <- createLogEntryf(labelError, format, v...)
	}

	os.Exit(1)
}

// DebugEnabled returns true, if logger would print a debug entry
func (l *Logger) DebugEnabled() bool {
	return l.minLevel >= LevelDebug
}

// InfoEnabled returns true, if logger would print an info entry
func (l *Logger) InfoEnabled() bool {
	return l.minLevel >= LevelInfo
}

// WarnEnabled returns true, if logger would print a warn entry
func (l *Logger) WarnEnabled() bool {
	return l.minLevel >= LevelWarn
}

// logEntry struct represents a log message to be written to log file.
// It contains all data necessary to render message.
type logEntry struct {
	Level    string    // level of this entry
	Message  string    // log message
	Filename string    // caller's filename
	Line     int       // caller's line number
	Time     time.Time // time of log event
}

// Stringer interface implementation
func (e logEntry) String() string {
	var format []byte = []byte("[%v] [%v:%v] [%v] %v")

	// Append end-of-line if caller didn't bother.
	if len(e.Message) == 0 || e.Message[len(e.Message)-1] != '\n' {
		format = append(format, '\n')
	}

	return fmt.Sprintf(string(format), e.Time, e.Filename, e.Line, e.Level,
		e.Message)
}

// createLogEntryf is equivalent to createLogEntry, but is using format string
func createLogEntryf(level, format string, v ...interface{}) logEntry {
	now := time.Now()
	file, line := callerInfo()
	return logEntry{
		Level:    level,
		Message:  fmt.Sprintf(format, v...),
		Filename: file,
		Line:     line,
		Time:     now,
	}
}

// createLogEntry prepares logEntry struct prefilled with appropriate data
func createLogEntry(level string, v ...interface{}) logEntry {
	now := time.Now()
	file, line := callerInfo()
	return logEntry{
		Level:    level,
		Message:  fmt.Sprint(v...),
		Filename: file,
		Line:     line,
		Time:     now,
	}
}

// get file and line of logger caller
func callerInfo() (file string, line int) {
	var ok bool
	_, file, line, ok = runtime.Caller(3)
	if !ok {
		file = "???.go"
		line = 0
	}

	short := file
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			short = file[i+1:]
			break
		}
	}
	file = short

	return
}
