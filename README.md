# go-log
Another custom logger implementation for golang.

Features:

* simple,
* thread-safe, achieved using channels,
* supports log levels.

TODO:

* log rotation

## Example usage

```go
package main

import (
	"fmt"
	log "github.com/jkusniar/go-log"
	"os"
)

func main() {
	// Open file. Must be WriteCloser (no Stdout/err)
	logfile := "application.log"
	file, err := os.OpenFile(logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file %v: %v", logfile, err)
		os.Exit(1)
	}

	// Start logger, defer proper logger shutdown
	Log := log.NewLogger(file, log.LEVEL_DEBUG)
	defer Log.Shutdown()

	Log.Debug("Debug message")
	Log.Infof("Info message %s", "hello")
	Log.SetLevel(log.LEVEL_INFO)
	Log.Debug("Message not printed!")
}
```
