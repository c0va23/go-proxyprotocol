package proxyprotocol_test

import (
	"errors"
	"net"
	"reflect"
	"testing"

	"github.com/c0va23/go-proxyprotocol"
	"github.com/golang/mock/gomock"
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

func TestListener_Accept(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	rawListener := NewMockListener(mockCtrl)

	builder := NewMockHeaderParserBuilder(mockCtrl)
	listener := proxyprotocol.NewListener(rawListener, builder)

	t.Run("when raw listener accept return error", func(t *testing.T) {
		acceptErr := errors.New("accept error")
		rawListener.EXPECT().Accept().Return(nil, acceptErr)

		conn, err := listener.Accept()

		if conn != nil {
			t.Errorf("Expect nil conn, but got %s", conn)
		}

		if err != acceptErr {
			t.Errorf("Unexpected error. Expect %s, got %s", acceptErr, err)
		}
	})

	t.Run("when raw listener accept return connection", func(t *testing.T) {
		rawConn := NewMockConn(mockCtrl)
		rawListener.EXPECT().Accept().Return(rawConn, nil).AnyTimes()

		remoteAddr := net.TCPAddr{
			IP:   net.IPv4(1, 2, 3, 4),
			Port: 12345,
		}
		rawConn.EXPECT().RemoteAddr().Return(&remoteAddr).AnyTimes()

		logger := proxyprotocol.FallbackLogger{Logger: listener.Logger}
		headerParser := NewMockHeaderParser(mockCtrl)

		t.Run("listener not have SourceChecker", func(t *testing.T) {
			builder.EXPECT().Build(logger).Return(headerParser)

			conn, err := listener.Accept()
			if err != nil {
				t.Errorf("Expect nil error, but got %s", err)
			}

			trustedAddr := true
			expectedConn := proxyprotocol.NewConn(rawConn, logger, headerParser, trustedAddr)
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
				sourceCheckErr = errors.New("source check err")
				sourceCheckResult = false

				conn, err := listener.Accept()
				if err != sourceCheckErr {
					t.Errorf("Unexpected error %s", err)
				}

				if conn != nil {
					t.Errorf("Unexpeced connection %s", conn)
				}
			})

			t.Run("source checker return false", func(t *testing.T) {
				sourceCheckErr = nil
				sourceCheckResult = false

				builder.EXPECT().Build(logger).Return(headerParser)

				conn, err := listener.Accept()
				if err != nil {
					t.Errorf("Unexpected error %s", err)
				}

				trustedAddr := false
				expectedConn := proxyprotocol.NewConn(rawConn, logger, headerParser, trustedAddr)
				if !reflect.DeepEqual(expectedConn, conn) {
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
				if err != nil {
					t.Errorf("Unexpected error %s", err)
				}

				trustedAddr := true
				expectedConn := proxyprotocol.NewConn(rawConn, logger, headerParser, trustedAddr)
				if !reflect.DeepEqual(expectedConn, conn) {
					t.Errorf("Unexpected connection %s", conn)
				}
			})
		})
	})
}
