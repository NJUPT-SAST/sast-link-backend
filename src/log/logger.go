package log

import (
	"os"
	"strings"
	"sync"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	rotationLog *lumberjack.Logger
	once        sync.Once
	Logger      *zap.Logger // Logger is the global logger for the application
)

func initRotationLog() {
	once.Do(func() {
		rotationLog = &lumberjack.Logger{
			Filename:   viper.GetString("log.file"),
			MaxSize:    viper.GetInt("log.max_size"), // megabytes
			MaxBackups: viper.GetInt("log.max_backups"),
			MaxAge:     viper.GetInt("log.max_age"), // days
			Compress:   viper.GetBool("log.compress"),
		}
	})
}

func NewLogger(options ...zap.Option) *zap.Logger {
	// Initialize the log rotation only once. This is necessary that the log file is not created multiple times.
	initRotationLog()

	return newZap(rotationLog, options...)
}

func WithModule(module string) zap.Option {
	return zap.Fields(zap.String("module", module))
}

func WithLayer(layer string) zap.Option {
	return zap.Fields(zap.String("layer", layer))
}

func WithComponent(component string) zap.Option {
	return zap.Fields(zap.String("component", component))
}

// newZap creates a new zap logger with the given log rotation.
func newZap(rotationLog *lumberjack.Logger, options ...zap.Option) *zap.Logger {
	encodeConfig := zap.NewProductionEncoderConfig()
	encodeConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	fileEncoder := zapcore.NewJSONEncoder(encodeConfig)
	consoleEncoder := zapcore.NewConsoleEncoder(encodeConfig)

	consoleWriter := zapcore.AddSync(os.Stdout)
	rotationWrite := zapcore.AddSync(rotationLog)

	var defaultLogLevel zapcore.Level

	switch strings.ToLower(viper.GetString("log.level")) {
	case "debug":
		defaultLogLevel = zapcore.DebugLevel
	case "info":
		defaultLogLevel = zapcore.InfoLevel
	case "warn":
		defaultLogLevel = zapcore.WarnLevel
	case "error":
		defaultLogLevel = zapcore.ErrorLevel
	default:
		defaultLogLevel = zapcore.InfoLevel
	}

	consoleCore := zapcore.NewCore(consoleEncoder, consoleWriter, defaultLogLevel)
	rotationCore := zapcore.NewCore(fileEncoder, rotationWrite, defaultLogLevel)

	core := zapcore.NewTee(consoleCore, rotationCore)

	options = append(options, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return zap.New(core, options...)
}

func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}
