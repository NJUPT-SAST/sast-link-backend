package log

import (
	"testing"
)

func TestInfo(_ *testing.T) {
	Info("info")
	Infof("info %s", "format")
	InfoWithFields("info", Fields{"key": "value"})
	Warn("warn")
	Warnf("warn %s", "format")
	WarnWithFields("warn", Fields{"key": "value"})
	SetLevel(logLevelSwitcher("debug"))
	Debug("ignore")
	Debug("debug")
	Debugf("debug %s", "format")
	DebugWithFields("debug", Fields{"key": "value"})
	Error("error")
	Errorf("error %s", "format")
	ErrorWithFields("error", Fields{"key": "value"})
}
