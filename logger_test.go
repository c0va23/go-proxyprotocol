package proxyprotocol_test

import (
	"reflect"
	"testing"

	"github.com/c0va23/go-proxyprotocol"
	"github.com/golang/mock/gomock"
)

func TestLoggerFunc(t *testing.T) {
	var format string
	var args []interface{}
	innerLogger := func(f string, v ...interface{}) {
		format = f
		args = v
	}

	logger := proxyprotocol.LoggerFunc(innerLogger)

	expectedFormat := "format %s %d"
	expectedArgs := []interface{}{"test", 123}
	logger.Printf(expectedFormat, expectedArgs...)

	if format != expectedFormat {
		t.Errorf("Uexpected format %s", format)
	}

	if !reflect.DeepEqual(args, expectedArgs) {
		t.Errorf("Unexpected args %v", args)
	}
}

func TestFallbackLogger_Printf(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	t.Run("when inner logger is nil", func(t *testing.T) {
		logger := proxyprotocol.FallbackLogger{Logger: nil}

		logger.Printf("Test")
	})

	t.Run("when inner logger valid logger", func(t *testing.T) {
		innerLogger := NewMockLogger(mockCtrl)
		innerLogger.EXPECT().Printf("Debug %s", 123)

		logger := proxyprotocol.FallbackLogger{Logger: innerLogger}

		logger.Printf("Debug %s", 123)
	})
}
