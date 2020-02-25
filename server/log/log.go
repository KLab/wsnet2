package log

import (
	"fmt"
	"io"
	"log"
	"os"
)

// Level type of loglevel
type Level int

const (
	// ALL output all logs
	ALL Level = iota
	// DEBUG output debug/info/error logs
	DEBUG
	// INFO output info/error logs
	INFO
	// ERROR output error logs
	ERROR
	// NOLOG output no logs
	NOLOG
)

var (
	logger = newLogger(os.Stdout)
	level  = INFO
)

func newLogger(w io.Writer) *log.Logger {
	return log.New(w, "", log.Ldate|log.Ltime|log.Lshortfile)
}

// SetWriter sets custom log writer
func SetWriter(w io.Writer) {
	logger = newLogger(w)
}

// CurrentLevel returns current log level
func CurrentLevel() Level {
	return level
}

// SetLevel sets log level
func SetLevel(l Level) (old Level) {
	old = level
	level = l
	return old
}

// Debugf outputs log for debug
func Debugf(format string, v ...interface{}) {
	if level <= DEBUG {
		output("[DEBUG] "+format, v...)
	}
}

// Infof outputs log for information
func Infof(format string, v ...interface{}) {
	if level <= INFO {
		output("[INFO] "+format, v...)
	}
}

// Errorf outputs log for error
func Errorf(format string, v ...interface{}) {
	if level <= ERROR {
		output("[ERROR] "+format, v...)
	}
}

func output(format string, v ...interface{}) {
	err := logger.Output(3, fmt.Sprintf(format, v...))
	if err != nil {
		log.Fatalf("logger output error: %v", err)
	}
}

// String implements Stringer interface
func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case ERROR:
		return "ERROR"
	}
	if l <= ALL {
		return "ALL"
	}
	return "NOLOG"
}
