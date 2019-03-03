package proxyprotocol_test

import (
	"reflect"
	"testing"

	"github.com/c0va23/go-proxyprotocol"
	"github.com/golang/mock/gomock"
)

func TestTextHeaderParserBuilder_Build(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := NewMockLogger(mockCtrl)

	headerParser := proxyprotocol.TextHeaderParserBuilder.Build(logger)

	expectedHeaderParser := proxyprotocol.NewTextHeaderParser(logger)

	if headerParser != expectedHeaderParser {
		t.Errorf("Unexpected header parser %v", headerParser)
	}
}

func TestBinaryHeaderParserBuilder_Build(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := NewMockLogger(mockCtrl)

	headerParser := proxyprotocol.BinaryHeaderParserBuilder.Build(logger)

	expectedHeaderParser := proxyprotocol.NewBinaryHeaderParser(logger)

	if headerParser != expectedHeaderParser {
		t.Errorf("Unexpected header parser %v", headerParser)
	}
}

func TestStubHeaderParserBuilder_Build(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := NewMockLogger(mockCtrl)

	headerParser := proxyprotocol.StubHeaderParserBuilder.Build(logger)

	expectedHeaderParser := proxyprotocol.NewStubHeaderParser()

	if headerParser != expectedHeaderParser {
		t.Errorf("Unexpected header parser %v", headerParser)
	}
}

func TestNewDefaultListener(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	listener := NewMockListener(mockCtrl)

	defaultListener := proxyprotocol.NewDefaultListener(listener)

	if defaultListener.Listener != listener {
		t.Errorf("Unexpected listener %v", defaultListener)
	}

	if !reflect.DeepEqual(defaultListener.HeaderParserBuilder, proxyprotocol.DefaultFallbackHeaderParserBuilder) {
		t.Errorf("Unexpected header parser builder %v", defaultListener.HeaderParserBuilder)
	}
}
