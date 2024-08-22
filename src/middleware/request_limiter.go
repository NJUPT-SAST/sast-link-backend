package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/didip/tollbooth/v7"
	"github.com/labstack/echo/v4"
)

// Example: RequestRateLimiter(3, time.Minute) represents 3 requests per minute
func RequestRateLimiter(maxRequests int64, period time.Duration) echo.MiddlewareFunc {
	rate := float64(maxRequests) / float64(period.Seconds())
	limiter := tollbooth.NewLimiter(rate, nil)
	limiter.SetMessage("Too many requests").
		SetOnLimitReached(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("limit reached")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("429 - Too many requests"))
		})
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if err := tollbooth.LimitByRequest(limiter, c.Response().Writer, c.Request()); err != nil {
				return err
			}
			return next(c)
		}
	}
}
