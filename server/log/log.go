package log

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"wsnet2/config"
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
	NOLOG Level = iota + 1
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

// Get Logger for custom log level.
func Get(l Level) *zap.Logger {
	return rootLogger.WithOptions(zap.IncreaseLevel(toZapLevel(l)))
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
	switch l {
	case NOLOG:
		return "NOLOG"
	case ERROR:
		return "ERROR"
	case INFO:
		return "INFO"
	case DEBUG:
		return "DEBUG"
	case ALL:
		return "ALL"
	}
	return fmt.Sprintf("Level(%d)", l)
}

func consoleTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("15:04:05.000"))
}

func InitLogger(logconf *config.LogConf) func() {
	// stdoutに出力するLogger
	var stdoutEnc zapcore.Encoder
	if logconf.LogStdoutConsole {
		// ローカル開発用 コンソール出力
		conf := zap.NewDevelopmentEncoderConfig()
		conf.EncodeLevel = zapcore.CapitalColorLevelEncoder
		conf.EncodeTime = consoleTimeEncoder
		stdoutEnc = zapcore.NewConsoleEncoder(conf)
	} else {
		conf := zap.NewProductionEncoderConfig()
		stdoutEnc = zapcore.NewJSONEncoder(conf)
	}
	core := zapcore.NewCore(stdoutEnc, os.Stdout, toZapLevel(Level(logconf.LogStdoutLevel)))

	// Fileに出力するLogger
	closer := func() {}
	if logconf.LogPath != "" {
		ljackLogger := &lumberjack.Logger{
			Filename:   logconf.LogPath,
			MaxSize:    logconf.LogMaxSize,
			MaxBackups: logconf.LogMaxBackups,
			MaxAge:     logconf.LogMaxAge,
			Compress:   logconf.LogCompress,
		}
		sink := zapcore.AddSync(ljackLogger)
		closer = func() {
			ljackLogger.Close()
		}
		conf := zap.NewProductionEncoderConfig()
		core2 := zapcore.NewCore(zapcore.NewJSONEncoder(conf), sink, zap.DebugLevel)
		core = zapcore.NewTee(core, core2)
	}

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
		closer()
	}
}
