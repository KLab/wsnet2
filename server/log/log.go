package log

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	rootLogger    *zap.Logger
	defaultLogger *zap.Logger
	wrappedLogger *zap.SugaredLogger

	defaultLogLevel = zap.NewAtomicLevel()
)

type Level = zapcore.Level

const (
	// ERROR output error logs
	ERROR = zap.ErrorLevel
	// INFO output info/error logs
	INFO = zap.InfoLevel
	// DEBUG output debug/info/error logs
	DEBUG = zap.DebugLevel
	// ALL output all logs
	ALL = zap.DebugLevel
)

var (
	level Level = INFO // global log level.
)

// Get Logger for custom log level.
func Get(l Level) *zap.Logger {
	return rootLogger.WithOptions(zap.IncreaseLevel(l))
}

// CurrentLevel returns global log level
func CurrentLevel() Level {
	return level
}

// SetLevel sets global log level
func SetLevel(l Level) Level {
	defaultLogLevel.SetLevel(l)
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

func consoleTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("15:04:05.000"))
}

func InitLogger() func() {
	// Consoleに出力するLogger
	consoleEnc := zap.NewDevelopmentEncoderConfig()
	consoleEnc.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEnc.EncodeTime = consoleTimeEncoder
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
