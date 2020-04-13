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
	// NOLOG output no logs
	NOLOG Level = iota
	// ERROR output error logs
	ERROR
	// INFO output info/error logs
	INFO
	// DEBUG output debug/info/error logs
	DEBUG
	// ALL output all logs
	ALL
)

const logFlags = log.Ldate | log.Ltime | log.Lshortfile

var (
	level  Level = INFO // global log level.
	logger       = log.New(os.Stdout, "", logFlags)
)

// Logger type
type Logger Level

// Get Logger for custom log level.
func Get(l Level) Logger {
	return Logger(l)
}

// Level returns logger log level.
func (l Logger) Level() Level {
	return Level(l)
}

// Debugf outputs log for debug
func (l Logger) Debugf(format string, v ...interface{}) {
	if Level(l) >= DEBUG {
		output("[DEBUG] "+format, v...)
	}
}

// Infof outputs log for information
func (l Logger) Infof(format string, v ...interface{}) {
	if Level(l) >= INFO {
		output("[INFO] "+format, v...)
	}
}

// Errorf outouts log for error
func (l Logger) Errorf(format string, v ...interface{}) {
	if Level(l) >= ERROR {
		output("[ERROR] "+format, v...)
	}
}

// SetWriter sets custom log writer.
func SetWriter(w io.Writer) {
	logger = log.New(w, "", logFlags)
}

// CurrentLevel returns global log level
func CurrentLevel() Level {
	return level
}

// SetLevel sets global log level
func SetLevel(l Level) Level {
	level, l = l, level
	return l
}

// Debugf outputs log for debug
func Debugf(format string, v ...interface{}) {
	if level >= DEBUG {
		output("[DEBUG] "+format, v...)
	}
}

// Infof outputs log for information
func Infof(format string, v ...interface{}) {
	if level >= INFO {
		output("[INFO] "+format, v...)
	}
}

// Errorf outputs log for error
func Errorf(format string, v ...interface{}) {
	if level >= ERROR {
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
	switch {
	case l <= NOLOG:
		return "NOLOG"
	case l == ERROR:
		return "ERROR"
	case l == INFO:
		return "INFO"
	case l == DEBUG:
		return "DEBUG"
	}
	return "ALL"
}
