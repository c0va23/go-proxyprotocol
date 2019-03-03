package proxyprotocol_test

import (
	"errors"
	"net"
	"reflect"
	"testing"

	"github.com/c0va23/go-proxyprotocol"
	gomock "github.com/golang/mock/gomock"
)

func TestNewListener(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	rawListener := NewMockListener(mockCtrl)

	builder := NewMockHeaderParserBuilder(mockCtrl)
	listener := proxyprotocol.NewListener(rawListener, builder)

	expectedListener := proxyprotocol.Listener{
		Listener:            rawListener,
		HeaderParserBuilder: builder,
	}

	if !reflect.DeepEqual(expectedListener, listener) {
		t.Errorf("NewListener return unexpected result. Expected %#v, got %#v.", expectedListener, listener)
	}
}

func TestListener_WithLogger(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	rawListener := NewMockListener(mockCtrl)

	builder := NewMockHeaderParserBuilder(mockCtrl)
	listener := proxyprotocol.NewListener(rawListener, builder)

	logger := proxyprotocol.LoggerFunc(t.Logf)
	withLogger := listener.WithLogger(logger)

	if reflect.ValueOf(withLogger.Logger).Pointer() != reflect.ValueOf(logger).Pointer() {
		t.Errorf("Unexpected LoggerFn")
	}
}

func TestListener_WithSourceChecker(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	rawListener := NewMockListener(mockCtrl)

	builder := NewMockHeaderParserBuilder(mockCtrl)
	listener := proxyprotocol.NewListener(rawListener, builder)

	sourceChecker := func(net.Addr) (bool, error) {
		return false, nil
	}

	withSourceChecker := listener.WithSourceChecker(sourceChecker)

	if reflect.ValueOf(withSourceChecker.SourceChecker).Pointer() != reflect.ValueOf(sourceChecker).Pointer() {
		t.Errorf("Unexpected SourceChecker")
	}
}

func TestListener_WithHeaderParserBuilder(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	rawListener := NewMockListener(mockCtrl)

	builder := NewMockHeaderParserBuilder(mockCtrl)
	listener := proxyprotocol.NewListener(rawListener, builder)

	otherBuilder := NewMockHeaderParserBuilder(mockCtrl)
	withHeaderParserBuilder := listener.WithHeaderParserBuilder(otherBuilder)

	if withHeaderParserBuilder.HeaderParserBuilder != otherBuilder {
		t.Errorf(
			"Unexpected HeaderParserBuilder. Expect %v, got %v.",
			otherBuilder,
			withHeaderParserBuilder.HeaderParserBuilder,
		)
	}
}

func TestListener_Close(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	rawListener := NewMockListener(mockCtrl)

	builder := NewMockHeaderParserBuilder(mockCtrl)
	listener := proxyprotocol.NewListener(rawListener, builder)

	expectedErr := errors.New("Closer error")
	rawListener.EXPECT().Close().Return(expectedErr)

	err := listener.Close()
	if expectedErr != err {
		t.Errorf("Unexpected close result. Expect %s, got %s", expectedErr, err)
	}
}

func TestListener_Addr(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	rawListener := NewMockListener(mockCtrl)

	builder := NewMockHeaderParserBuilder(mockCtrl)
	listener := proxyprotocol.NewListener(rawListener, builder)

	expectedAddr := &net.TCPAddr{
		IP:   net.IPv4(1, 2, 3, 4),
		Port: 8080,
	}

	rawListener.EXPECT().Addr().Return(expectedAddr)

	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok || !expectedAddr.IP.Equal(addr.IP) || expectedAddr.Port != addr.Port {
		t.Errorf("Unexpected addr result. Expect %s, got %s", expectedAddr, addr)
	}
}

func TestListener_Accept(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	rawListener := NewMockListener(mockCtrl)

	builder := NewMockHeaderParserBuilder(mockCtrl)
	listener := proxyprotocol.NewListener(rawListener, builder)

	t.Run("when raw listener accept return error", func(t *testing.T) {
		acceptErr := errors.New("Accept error")
		rawListener.EXPECT().Accept().Return(nil, acceptErr)

		conn, err := listener.Accept()

		if nil != conn {
			t.Errorf("Expect nil conn, but got %s", conn)
		}

		if err != acceptErr {
			t.Errorf("Unexpected error. Expect %s, got %s", acceptErr, err)
		}
	})

	t.Run("when raw listner accept return connection", func(t *testing.T) {
		rawConn := NewMockConn(mockCtrl)
		rawListener.EXPECT().Accept().Return(rawConn, nil).AnyTimes()

		remoteAddr := net.TCPAddr{
			IP:   net.IPv4(1, 2, 3, 4),
			Port: 12345,
		}
		rawConn.EXPECT().RemoteAddr().Return(&remoteAddr).AnyTimes()

		t.Run("listener not have SourceChecker", func(t *testing.T) {
			logger := proxyprotocol.FallbackLogger{Logger: listener.Logger}
			headerParser := NewMockHeaderParser(mockCtrl)
			builder.EXPECT().Build(logger).Return(headerParser)

			conn, err := listener.Accept()

			if nil != err {
				t.Errorf("Expect nil error, but got %s", err)
			}

			expectedConn := proxyprotocol.NewConn(rawConn, logger, headerParser)
			if !reflect.DeepEqual(expectedConn, conn) {
				t.Errorf("Unexpected connection %s", conn)
			}
		})

		t.Run("listener have SourceChecker", func(t *testing.T) {
			var sourceCheckErr error
			var sourceCheckResult bool
			sourceChecker := func(net.Addr) (bool, error) {
				return sourceCheckResult, sourceCheckErr
			}
			listener = listener.WithSourceChecker(sourceChecker)

			t.Run("source checker return error", func(t *testing.T) {
				sourceCheckErr = errors.New("Source check err")
				sourceCheckResult = false

				conn, err := listener.Accept()
				if err != sourceCheckErr {
					t.Errorf("Unexpected error %s", err)
				}

				if nil != conn {
					t.Errorf("Unexpeced connection %s", conn)
				}
			})

			t.Run("source checker return false", func(t *testing.T) {
				sourceCheckErr = nil
				sourceCheckResult = false

				conn, err := listener.Accept()
				if nil != err {
					t.Errorf("Unexpected error %s", err)
				}

				if rawConn != conn {
					t.Errorf("Unexpected connection %s", conn)
				}
			})

			t.Run("source checker return true", func(t *testing.T) {
				sourceCheckErr = nil
				sourceCheckResult = true

				logger := proxyprotocol.FallbackLogger{Logger: listener.Logger}
				headerParser := NewMockHeaderParser(mockCtrl)
				builder.EXPECT().Build(logger).Return(headerParser)

				conn, err := listener.Accept()
				if nil != err {
					t.Errorf("Unexpected error %s", err)
				}

				expectedConn := proxyprotocol.NewConn(rawConn, logger, headerParser)
				if !reflect.DeepEqual(expectedConn, conn) {
					t.Errorf("Unexpected connection %s", conn)
				}
			})
		})
	})
}
