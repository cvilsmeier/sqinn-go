package sqinn

import (
	"log"
)

// A Logger logs error and debug messages coming from the stderr
// output of the sqinn child process.
type Logger interface {
	Log(s string)
}

// A StdLogger logs to a stdlib log.Logger or to the
// log's standard logger.
type StdLogger struct {

	// Logger will be used for writing log outputs.
	// If Logger is nil, the log package default output will be used.
	Logger *log.Logger
}

// Log logs s to a log.Logger or to the default log output.
func (l StdLogger) Log(s string) {
	if l.Logger != nil {
		l.Logger.Println(s)
	} else {
		log.Println(s)
	}
}

// NoLogger does not log anything.
type NoLogger struct{}

// Log does nothing.
func (l NoLogger) Log(s string) {}
