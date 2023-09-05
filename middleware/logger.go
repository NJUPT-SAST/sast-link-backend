package middleware

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func MiddlewareLogging(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Stop timer
		end := time.Now()
		latency := end.Sub(start)

		// Get request status code
		status := c.Writer.Status()

		baseFields := logrus.Fields{
			"status":  status,
			"method":  c.Request.Method,
			"path":    c.Request.URL.Path,
			"ip":      c.ClientIP(),
			"latency": latency,
		}

		switch log.GetLevel() {
		case logrus.DebugLevel:
			// Get params
			var params = url.Values{}
			if c.Request.Method == "GET" {
				params = c.Request.URL.Query()
			} else if c.Request.Method == "POST" {
				err := c.Request.ParseForm()
				if err != nil {
					log.WithFields(baseFields).Error("Error parsing form")
				}
				err = c.Request.ParseMultipartForm(0)
				if err != nil {
					log.WithFields(baseFields).Error("Error parsing form")
				}
				params = c.Request.PostForm
			} else {
				params = c.Request.URL.Query()
			}

			formatParams := formatParams(params)
			// Format headers for readability
			formattedHeaders := formatHeaders(c.Request.Header)

			// Create a formatted log entry
			logEntry := log.WithFields(baseFields)
			logEntry = logEntry.WithField("params", formatParams)
			logEntry = logEntry.WithField("headers", formattedHeaders)
			logEntry.Debug("Request details")
		case logrus.InfoLevel:
			logEntry := log.WithFields(baseFields)
			logEntry.Info("Request details")
		}
	}
}

func formatHeaders(headers http.Header) string {
	var formattedHeaders strings.Builder

	for key, values := range headers {
		formattedHeaders.WriteString(key)
		formattedHeaders.WriteString(": [")
		formattedHeaders.WriteString(strings.Join(values, ", "))
		formattedHeaders.WriteString("] ")
	}

	return formattedHeaders.String()
}

func formatParams(params url.Values) string {
	var formattedParams strings.Builder

	for key, values := range params {
		formattedParams.WriteString(key)
		formattedParams.WriteString(": [")
		formattedParams.WriteString(strings.Join(values, ", "))
		formattedParams.WriteString("] ")
	}

	return formattedParams.String()
}
