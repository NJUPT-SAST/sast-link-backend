package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/didip/tollbooth/v7"
	"github.com/gin-gonic/gin"
)

// Example: RequestRateLimiter(3, time.Minute) represents 3 requests per minute
func RequestRateLimiter(maxRequests int64, period time.Duration) gin.HandlerFunc {
	rate := float64(maxRequests) / float64(period.Seconds())
	limiter := tollbooth.NewLimiter(rate, nil)
	limiter.SetMessage("Too many requests").
		SetOnLimitReached(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("limit reached")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("429 - Too many requests"))
		})
	return func(c *gin.Context) {
		httpError := tollbooth.LimitByRequest(limiter, c.Writer, c.Request)
		if httpError != nil {
			c.Data(httpError.StatusCode, limiter.GetMessageContentType(), []byte(httpError.Message))
			c.Abort()
		} else {
			c.Next()
		}
	}
}
