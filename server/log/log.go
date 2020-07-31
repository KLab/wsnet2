package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	rootLogger    *zap.Logger
	defaultLogger *zap.Logger
	wrappedLogger *zap.SugaredLogger

	defaultLogLevel = zap.NewAtomicLevel()
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

var (
	level Level = INFO // global log level.
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
		wrappedLogger.Debugf(format, v...)
	}
}

// Infof outputs log for information
func (l Logger) Infof(format string, v ...interface{}) {
	if Level(l) >= INFO {
		wrappedLogger.Infof(format, v...)
	}
}

// Errorf outouts log for error
func (l Logger) Errorf(format string, v ...interface{}) {
	if Level(l) >= ERROR {
		wrappedLogger.Errorf(format, v...)
	}
}

// CurrentLevel returns global log level
func CurrentLevel() Level {
	return level
}

func toZapLevel(l Level) zapcore.Level {
	switch l {
	case NOLOG:
		return zapcore.PanicLevel
	case ERROR:
		return zapcore.ErrorLevel
	case INFO:
		return zapcore.InfoLevel
	case DEBUG, ALL:
		return zapcore.DebugLevel
	}
	Errorf("Unknown level: %v", l)
	return zapcore.DebugLevel
}

// SetLevel sets global log level
func SetLevel(l Level) Level {
	defaultLogLevel.SetLevel(toZapLevel(l))

	level, l = l, level
	return l
}

// Debugf outputs log for debug
func Debugf(format string, v ...interface{}) {
	if level >= DEBUG {
		wrappedLogger.Debugf(format, v...)
	}
}

// Infof outputs log for information
func Infof(format string, v ...interface{}) {
	if level >= INFO {
		wrappedLogger.Infof(format, v...)
	}
}

// Errorf outputs log for error
func Errorf(format string, v ...interface{}) {
	if level >= ERROR {
		wrappedLogger.Errorf(format, v...)
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

func InitLogger() func() {
	// Consoleに出力するLogger
	consoleEnc := zap.NewDevelopmentEncoderConfig()
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(consoleEnc), os.Stdout, zap.DebugLevel)

	// TODO: 指定されたファイルに出力する。
	// sink, closer, err := zap.Open("/tmp/zaplog.out")
	// if err != nil {
	// 	panic(err)
	// }
	// fileEnc := zap.NewProductionEncoderConfig()
	// core2 := zapcore.NewCore(zapcore.NewJSONEncoder(fileEnc), sink, zap.DebugLevel)
	// core := zapcore.NewTee(core, core2)

	logger := zap.New(core, zap.AddStacktrace(zap.WarnLevel), zap.WithCaller(true))
	rootLogger = logger
	defaultLogger = logger.WithOptions(zap.IncreaseLevel(zap.InfoLevel))
	wrappedLogger = logger.WithOptions(zap.AddCallerSkip(1)).Sugar()

	// zap.S().Debugf() とかで使える logger を設定する。
	zap.ReplaceGlobals(logger)
	// 標準ライブラリの "log" パッケージを使ったログを流し込む。
	zap.RedirectStdLog(logger)

	return func() {
		logger.Sync()
		// closer()
	}
}
