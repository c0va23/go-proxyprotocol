package proxyprotocol_test

import (
	"reflect"
	"testing"

	"github.com/c0va23/go-proxyprotocol"
	gomock "github.com/golang/mock/gomock"
)

func TestNewListener(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	rawListener := NewMockListener(mockCtrl)

	listener := proxyprotocol.NewListener(rawListener)

	expectedListener := &proxyprotocol.Listener{
		Listener:            rawListener,
		HeaderParserBuilder: proxyprotocol.DefaultFallbackHeaderParserBuilder(),
	}

	if !reflect.DeepEqual(expectedListener, listener) {
		t.Errorf("NewListener return unexpected result. Expected %#v, got %#v.", expectedListener, listener)
	}
}
