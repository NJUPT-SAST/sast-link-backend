package log

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/shiena/ansicolor"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	Log = logrus.New()
)

// Fields wraps logrus.Fields, which is a map[string]interface{}
type Fields logrus.Fields

// CustomFormatter is a custom log formatter
type CustomFormatter struct {
	ForceQuote       bool
	DisableQuote     bool
	TimestampFormat  string
	QuoteEmptyFields bool
}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var levelColor string
	switch entry.Level {
	case logrus.DebugLevel:
		levelColor = "\033[36m" // Cyan
	case logrus.InfoLevel:
		levelColor = "\033[32m" // Green
	case logrus.WarnLevel:
		levelColor = "\033[33m" // Yellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = "\033[31m" // Red
	default:
		levelColor = "\033[0m" // Reset
	}

	// Handle timestamp
	var timestamp string
	timestamp = entry.Time.Format(f.TimestampFormat)

	// Handle message quoting
	var message string
	if f.ForceQuote || (!f.DisableQuote && (len(entry.Message) == 0 || f.QuoteEmptyFields)) {
		message = fmt.Sprintf("%q", entry.Message)
	} else {
		message = entry.Message
	}

	// Format the log level
	logLevel := fmt.Sprintf("%s%s%s", levelColor, strings.ToUpper(entry.Level.String()), "\033[0m")

	// Get caller info from data fields
	callerInfo := entry.Data["file"]

	fields := make([]string, 0, len(entry.Data))
	for k, v := range entry.Data {
		if k == "file" {
			continue
		}
		fields = append(fields, fmt.Sprintf("%s=%v", k, v))
	}

	var formattedLog string
	if len(fields) > 0 {
		formattedLog = fmt.Sprintf("[%-16s %s %-18s] %s %s\n", logLevel, timestamp, callerInfo, message, fields)
	} else {
		formattedLog = fmt.Sprintf("[%-16s %s %-18s] %s\n", logLevel, timestamp, callerInfo, message)
	}
	// Combine the formatted log entry
	return []byte(formattedLog), nil
}

func SetupLogger() {
	// use ansicolor to add console color
	// FIXME: Waiting for test.
	// see: https://github.com/sirupsen/logrus/issues/1115
	Log.SetOutput(ansicolor.NewAnsiColorWriter(os.Stdout))
	Log.SetLevel(logLevelSwitcher(viper.GetString("log.level")))
	// add caller message(method and file)
	Log.SetReportCaller(true)
	Log.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		ForceQuote:      true,
		TimestampFormat: time.RFC3339,
		FullTimestamp:   true,
	})

	// file, err := os.OpenFile("log.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	// if err == nil {
	// 	Log.SetOutput(file)
	// } else {
	// 	Log.Info("Failed to log to file, using default stderr")
	// }
}

func SetLevel(level logrus.Level) {
	Log.SetLevel(level)
}

// Debug logs a message at level Debug on the standard logger.
// Usage:
// log.Debug("info")
func Debug(args ...interface{}) {
	if Log.Level >= logrus.DebugLevel {
		entry := Log.WithFields(logrus.Fields{})
		entry.Data["file"] = fileInfo(2)
		entry.Debug(args...)
	}
}

// Usage:
// log.Debugf("info %s", "format")
func Debugf(format string, args ...interface{}) {
	if Log.Level >= logrus.DebugLevel {
		entry := Log.WithFields(logrus.Fields{})
		entry.Data["file"] = fileInfo(2)
		entry.Debugf(format, args...)
	}
}

// Debug logs a message with fields at level Debug on the standard logger.
// Usage:
// log.DebugWithFields("info", log.Fields{"key": "value"})
func DebugWithFields(l interface{}, f Fields) {
	if Log.Level >= logrus.DebugLevel {
		entry := Log.WithFields(logrus.Fields(f))
		entry.Data["file"] = fileInfo(2)
		entry.Debug(l)
	}
}

// Info logs a message at level Info on the standard logger.
func Info(args ...interface{}) {
	if Log.Level >= logrus.InfoLevel {
		entry := Log.WithFields(logrus.Fields{})
		entry.Data["file"] = fileInfo(2)
		entry.Info(args...)
	}
}

func Infof(format string, args ...interface{}) {
	if Log.Level >= logrus.InfoLevel {
		entry := Log.WithFields(logrus.Fields{})
		entry.Data["file"] = fileInfo(2)
		entry.Infof(format, args...)
	}
}

// Debug logs a message with fields at level Debug on the standard logger.
func InfoWithFields(l interface{}, f Fields) {
	if Log.Level >= logrus.InfoLevel {
		entry := Log.WithFields(logrus.Fields(f))
		entry.Data["file"] = fileInfo(2)
		entry.Info(l)
	}
}

// Warn logs a message at level Warn on the standard logger.
func Warn(args ...interface{}) {
	if Log.Level >= logrus.WarnLevel {
		entry := Log.WithFields(logrus.Fields{})
		entry.Data["file"] = fileInfo(2)
		entry.Warn(args...)
	}
}

func Warnf(format string, args ...interface{}) {
	if Log.Level >= logrus.WarnLevel {
		entry := Log.WithFields(logrus.Fields{})
		entry.Data["file"] = fileInfo(2)
		entry.Warnf(format, args...)
	}
}

// Debug logs a message with fields at level Debug on the standard logger.
func WarnWithFields(l interface{}, f Fields) {
	if Log.Level >= logrus.WarnLevel {
		entry := Log.WithFields(logrus.Fields(f))
		entry.Data["file"] = fileInfo(2)
		entry.Warn(l)
	}
}

// Error logs a message at level Error on the standard logger.
func Error(args ...interface{}) {
	if Log.Level >= logrus.ErrorLevel {
		entry := Log.WithFields(logrus.Fields{})
		entry.Data["file"] = fileInfo(2)
		entry.Error(args...)
	}
}

func Errorf(format string, args ...interface{}) {
	if Log.Level >= logrus.ErrorLevel {
		entry := Log.WithFields(logrus.Fields{})
		entry.Data["file"] = fileInfo(2)
		entry.Errorf(format, args...)
	}
}

// Debug logs a message with fields at level Debug on the standard logger.
func ErrorWithFields(l interface{}, f Fields) {
	if Log.Level >= logrus.ErrorLevel {
		entry := Log.WithFields(logrus.Fields(f))
		entry.Data["file"] = fileInfo(2)
		entry.Error(l)
	}
}

// Fatal logs a message at level Fatal on the standard logger.
func Fatal(args ...interface{}) {
	if Log.Level >= logrus.FatalLevel {
		entry := Log.WithFields(logrus.Fields{})
		entry.Data["file"] = fileInfo(2)
		entry.Fatal(args...)
	}
}

func Fatalf(format string, args ...interface{}) {
	if Log.Level >= logrus.FatalLevel {
		entry := Log.WithFields(logrus.Fields{})
		entry.Data["file"] = fileInfo(2)
		entry.Fatalf(format, args...)
	}
}

// Debug logs a message with fields at level Debug on the standard logger.
func FatalWithFields(l interface{}, f Fields) {
	if Log.Level >= logrus.FatalLevel {
		entry := Log.WithFields(logrus.Fields(f))
		entry.Data["file"] = fileInfo(2)
		entry.Fatal(l)
	}
}

// Panic logs a message at level Panic on the standard logger.
func Panic(args ...interface{}) {
	if Log.Level >= logrus.PanicLevel {
		entry := Log.WithFields(logrus.Fields{})
		entry.Data["file"] = fileInfo(2)
		entry.Panic(args...)
	}
}

func Panicf(format string, args ...interface{}) {
	if Log.Level >= logrus.PanicLevel {
		entry := Log.WithFields(logrus.Fields{})
		entry.Data["file"] = fileInfo(2)
		entry.Panicf(format, args...)
	}
}

// Debug logs a message with fields at level Debug on the standard logger.
func PanicWithFields(l interface{}, f Fields) {
	if Log.Level >= logrus.PanicLevel {
		entry := Log.WithFields(logrus.Fields(f))
		entry.Data["file"] = fileInfo(2)
		entry.Panic(l)
	}
}

func fileInfo(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		file = "<???>"
		line = 1
	} else {
		slash := strings.LastIndex(file, "/")
		if slash >= 0 {
			file = file[slash+1:]
		}
	}
	return fmt.Sprintf("%s:%d", file, line)
}

func logLevelSwitcher(level string) logrus.Level {
	level = strings.ToLower(level)
	switch level {
	case "trace":
		return logrus.TraceLevel
	case "info":
		return logrus.InfoLevel
	case "debug":
		return logrus.DebugLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "fatal":
		return logrus.FatalLevel
	case "panic":
		return logrus.PanicLevel
	}
	return logrus.DebugLevel
}

// LogReq print request header and body with `debug` level log.
//
// use io.NopCloser to copy request body, so req.Body can be read again
//
// !: need to use this **before** manually read from body using `io.ReadAll(req.Body)`
func LogReq(req *http.Request) {
	// Get request usually have no body
	reqBody := []byte{}
	if req.Body != nil {
		reqBody, _ := io.ReadAll(req.Body)

		req.Body = io.NopCloser(bytes.NewReader(reqBody))
	}

	reqFormat := `
	  Request to be sent:
	`

	reqFormat += fmt.Sprintf(`
			URL: %s %s
	`,
		req.Host,
		req.URL.Path,
	)
	reqFormat += `
		Header:
	`
	for k, v := range req.Header {
		reqFormat += fmt.Sprintf(
			`
				%s: %s
		`,
			k,
			v,
		)
	}
	reqFormat += `
			Body:
				%s
	`
	Log.Warnf(reqFormat, string(reqBody))
}

// LogRes print response header and body with `debug` level log
//
// use io.NopCloser to copy response body, so req.Body can be read again
//
// !: need to use this **before** manually read from body using `io.ReadAll(res.Body)`
func LogRes(res *http.Response) {
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		Log.Errorln("io.ReadAll: ", err)
	}

	res.Body = io.NopCloser(bytes.NewReader(resBody))

	resFormat := `
	    Response received:
	`
	resFormat += `
		Header:
	`
	for k, v := range res.Header {
		resFormat += fmt.Sprintf(
			`
		        %s: %s
		`,
			k,
			v,
		)
	}
	resFormat += `
	        Body:
		        %s
	`
	Log.Warnf(resFormat, string(resBody))
}
