package log_test

import (
	"fmt"
	"os"

	log "github.com/jkusniar/go-log"
)

func Example() {
	// Open file. Must be WriteCloser (no Stdout/err)
	logfile := "application.log"
	file, err := os.OpenFile(logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file %v: %v", logfile, err)
		os.Exit(1)
	}

	// Start logger, defer proper logger shutdown
	Log := log.New(file, log.LevelDebug)
	defer Log.Shutdown()

	Log.Debug("Debug message")
	Log.Infof("Info message %s", "hello")
	Log.SetLevel(log.LevelInfo)
	Log.Debug("Message not printed!")
}
