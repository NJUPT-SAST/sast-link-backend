package log

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/NJUPT-SAST/sast-link-backend/config"
	"github.com/shiena/ansicolor"
	"github.com/sirupsen/logrus"
)

var (
	Logger = logrus.New()
	logLevel = config.Config.GetString("log.level")
)

func init() {
	initLogger()
}

func initLogger() {
	// use ansicolor to add console color
	Logger.SetOutput(ansicolor.NewAnsiColorWriter(os.Stdout))
	Logger.SetLevel(logLevelSwitcher(logLevel))
	// add caller message(method and file)
	Logger.SetReportCaller(true)
	Logger.SetFormatter(&logrus.TextFormatter{
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

// LogReq print request header and body with `debug` level log.
// 
// use io.NopCloser to copy request body, so req.Body can be read again
// !: need to use this **before** manually read from body using `io.ReadAll(req.Body)`
func LogReq(req *http.Request) {
	reqBody, _ := io.ReadAll(req.Body)

	req.Body = io.NopCloser(bytes.NewReader(reqBody))

	reqFormat := `
	  Request to be sent:
	`;
	
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
	Logger.Warnf(reqFormat, string(reqBody))
}

// LogRes print response header and body with `debug` level log
//
// use io.NopCloser to copy response body, so req.Body can be read again
// !: need to use this **before** manually read from body using `io.ReadAll(res.Body)`
func LogRes(res *http.Response) {
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		Logger.Errorln("io.ReadAll: ", err)
	}

	res.Body = io.NopCloser(bytes.NewReader(resBody))

	resFormat := `
	    Response received:
	`;
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
	Logger.Warnf(resFormat, string(resBody))
}
