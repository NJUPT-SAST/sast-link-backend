package log

import (
	"os"

	"github.com/NJUPT-SAST/sast-link-backend/config"
	"github.com/shiena/ansicolor"
	"github.com/sirupsen/logrus"
)

var Log = logrus.New()

func init() {
	initLogger()
}

func initLogger() {
	// use ansicolor to add console color
	Log.SetOutput(ansicolor.NewAnsiColorWriter(os.Stdout))
	Log.SetLevel(logLevelSwitcher(config.Config.GetString("log.level")))
	// add caller message(method and file)
	Log.SetReportCaller(true)
	Log.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		ForceQuote:      true,
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})
	// TODO: implement the `logrus.Formatter` interface
	// as self log format
}

func logLevelSwitcher(level string) logrus.Level {
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
