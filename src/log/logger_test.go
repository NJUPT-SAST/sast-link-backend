package log

import (
	"testing"

	"go.uber.org/zap"
)

func TestLogger(t *testing.T) {
	logger := NewLogger(WithLayer("test_layer"), WithModule("test_module"))
	logger.Info("test info", zap.String("key", "value"))

	logger = NewLogger(WithModule("test_module"), WithLayer("test_layer"))
	logger.Info("test info 2", zap.String("key", "value"))
}
