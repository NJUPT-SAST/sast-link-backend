package router

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitRouter(t *testing.T) {
	router := InitRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}

// Test limiter
func TestLimiter(t *testing.T) {
	router := InitRouter()
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/api/v1/user/info", nil)
			router.ServeHTTP(w, req)
			//t.Log("result:" + w.Body.String())
			assert.Equal(t, http.StatusOK, w.Code)
		}()
	}
	wg.Wait()
}
