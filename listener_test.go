package proxyprotocol_test

import (
	"bufio"
	"errors"
	"net"
	"reflect"
	"testing"
	"time"

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

func TestListener_WithLogger(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	rawListener := NewMockListener(mockCtrl)

	listener := proxyprotocol.NewListener(rawListener)

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

	listener := proxyprotocol.NewListener(rawListener)

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

	listener := proxyprotocol.NewListener(rawListener)

	binaryHeaderParserBuilder := new(proxyprotocol.BinaryHeaderParserBuilder)
	withHeaderParserBuilder := listener.WithHeaderParserBuilder(binaryHeaderParserBuilder)

	if withHeaderParserBuilder.HeaderParserBuilder != binaryHeaderParserBuilder {
		t.Errorf(
			"Unexpected HeaderParserBuilder. Expect %s, got %s.",
			binaryHeaderParserBuilder,
			withHeaderParserBuilder.HeaderParserBuilder,
		)
	}
}

func TestListener_Close(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	rawListener := NewMockListener(mockCtrl)

	listener := proxyprotocol.NewListener(rawListener)

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

	listener := proxyprotocol.NewListener(rawListener)

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

	listener := proxyprotocol.NewListener(rawListener)

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
			conn, err := listener.Accept()

			if nil != err {
				t.Errorf("Expect nil error, but got %s", err)
			}

			logger := proxyprotocol.FallbackLogger{Logger: listener.Logger}
			headerParser := listener.HeaderParserBuilder.Build(logger)
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

				conn, err := listener.Accept()
				if nil != err {
					t.Errorf("Unexpected error %s", err)
				}

				logger := proxyprotocol.FallbackLogger{Logger: listener.Logger}
				headerParser := listener.HeaderParserBuilder.Build(logger)
				expectedConn := proxyprotocol.NewConn(rawConn, logger, headerParser)
				if !reflect.DeepEqual(expectedConn, conn) {
					t.Errorf("Unexpected connection %s", conn)
				}
			})
		})
	})
}

func TestConn_Read(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	rawConn := NewMockConn(mockCtrl)

	rawAddr := &net.TCPAddr{
		IP:   net.IPv4(10, 0, 0, 1),
		Port: 12345,
	}

	rawConn.EXPECT().RemoteAddr().Return(rawAddr).AnyTimes()

	headerParser := NewMockHeaderParser(mockCtrl)
	logger := NewMockLogger(mockCtrl)

	readBuf := bufio.NewReaderSize(rawConn, 1400)

	buf := make([]byte, readBuf.Size())

	t.Run("when header parser return err", func(t *testing.T) {
		parseErr := errors.New("Parse error")
		headerParser.EXPECT().Parse(readBuf).Return(nil, parseErr)

		conn := proxyprotocol.NewConn(rawConn, logger, headerParser)

		n, err := conn.Read(buf)

		if err != parseErr {
			t.Errorf("Unexpected error %s", err)
		}

		if n != 0 {
			t.Errorf("Unexpected read size %d", n)
		}

		if srcAddr := conn.RemoteAddr(); !reflect.DeepEqual(srcAddr, rawAddr) {
			t.Errorf("Unexpected remote addr %s", srcAddr)
		}
	})

	t.Run("when header parser return Header", func(t *testing.T) {
		header := &proxyprotocol.Header{
			SrcAddr: &net.TCPAddr{
				IP: net.IPv4(1, 2, 3, 4),
			},
		}

		headerParser.EXPECT().Parse(readBuf).Return(header, nil).AnyTimes()

		t.Run("when rawConn.Read return err", func(t *testing.T) {
			readErr := errors.New("Read error")
			rawConn.EXPECT().Read(buf).Return(0, readErr)

			conn := proxyprotocol.NewConn(rawConn, logger, headerParser)

			n, err := conn.Read(buf)
			if err != readErr {
				t.Errorf("Unexpected error %s", err)
			}

			if n != 0 {
				t.Errorf("Unexpected read size %d", n)
			}

			if srcAddr := conn.RemoteAddr(); !reflect.DeepEqual(srcAddr, header.SrcAddr) {
				t.Errorf("Unexpected remote addr %s", srcAddr)
			}
		})

		t.Run("when rawConn.Read return read size", func(t *testing.T) {
			readSize := 512
			rawConn.EXPECT().Read(buf).Return(readSize, nil)

			conn := proxyprotocol.NewConn(rawConn, logger, headerParser)

			n, err := conn.Read(buf)

			if err != nil {
				t.Errorf("Unexpected error %s", err)
			}

			if n != readSize {
				t.Errorf("Unexpected read size %d", readSize)
			}

			if srcAddr := conn.RemoteAddr(); !reflect.DeepEqual(srcAddr, header.SrcAddr) {
				t.Errorf("Unexpected remote addr %s", srcAddr)
			}
		})
	})
}

func TestConnRemoteAddr(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	rawConn := NewMockConn(mockCtrl)

	rawAddr := &net.TCPAddr{
		IP:   net.IPv4(10, 0, 0, 1),
		Port: 12345,
	}

	rawConn.EXPECT().RemoteAddr().Return(rawAddr).AnyTimes()

	headerParser := NewMockHeaderParser(mockCtrl)
	logger := NewMockLogger(mockCtrl)

	readBuf := bufio.NewReaderSize(rawConn, 1400)

	t.Run("when header parser return error", func(t *testing.T) {
		conn := proxyprotocol.NewConn(rawConn, logger, headerParser)

		parseErr := errors.New("Parse error")

		headerParser.EXPECT().Parse(readBuf).Return(nil, parseErr)
		logger.EXPECT().Printf(gomock.Any(), gomock.Any())

		remoteAddr := conn.RemoteAddr()

		if !reflect.DeepEqual(remoteAddr, rawAddr) {
			t.Errorf("Unexpected remote addr %s", remoteAddr)
		}
	})

	t.Run("when header parser return header", func(t *testing.T) {
		conn := proxyprotocol.NewConn(rawConn, logger, headerParser)

		header := &proxyprotocol.Header{
			SrcAddr: &net.TCPAddr{
				IP: net.IPv4(1, 2, 3, 4),
			},
		}

		headerParser.EXPECT().Parse(readBuf).Return(header, nil)

		remoteAddr := conn.RemoteAddr()

		if !reflect.DeepEqual(remoteAddr, header.SrcAddr) {
			t.Errorf("Unexpected remote adder %s", remoteAddr)
		}

		t.Run("when second call .RemoteAddr()", func(t *testing.T) {
			remoteAddr := conn.RemoteAddr()

			if !reflect.DeepEqual(remoteAddr, header.SrcAddr) {
				t.Errorf("Unexpected remote adder %s", remoteAddr)
			}
		})
	})
}

func TestConn_Close(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	rawConn := NewMockConn(mockCtrl)

	headerParser := NewMockHeaderParser(mockCtrl)
	logger := NewMockLogger(mockCtrl)

	conn := proxyprotocol.NewConn(rawConn, logger, headerParser)

	t.Run("when rawConn.Close() return error", func(t *testing.T) {
		closeErr := errors.New("Close error")
		rawConn.EXPECT().Close().Return(closeErr)

		err := conn.Close()

		if err != closeErr {
			t.Errorf("Unexpected error %s", err)
		}
	})

	t.Run("when rawConn.Close() return nil", func(t *testing.T) {
		rawConn.EXPECT().Close().Return(nil)

		err := conn.Close()

		if err != nil {
			t.Errorf("Unexpected error %s", err)
		}
	})
}

func TestConn_LocalAddr(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	rawConn := NewMockConn(mockCtrl)

	headerParser := NewMockHeaderParser(mockCtrl)
	logger := NewMockLogger(mockCtrl)

	conn := proxyprotocol.NewConn(rawConn, logger, headerParser)

	rawAddr := &net.TCPAddr{
		IP:   net.IPv4(10, 0, 0, 1),
		Port: 12345,
	}
	rawConn.EXPECT().LocalAddr().Return(rawAddr)

	localAddr := conn.LocalAddr()
	if localAddr != rawAddr {
		t.Errorf("Unexpected local addr %s", localAddr)
	}
}

func TestConn_Write(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	rawConn := NewMockConn(mockCtrl)

	headerParser := NewMockHeaderParser(mockCtrl)
	logger := NewMockLogger(mockCtrl)

	conn := proxyprotocol.NewConn(rawConn, logger, headerParser)

	buf := []byte{1, 2, 3, 4, 5}

	t.Run("when rawConn.Write return error", func(t *testing.T) {
		writeErr := errors.New("Write error")
		rawConn.EXPECT().Write(buf).Return(0, writeErr)

		n, err := conn.Write(buf)
		if err != writeErr {
			t.Errorf("Unexpected error %s", err)
		}

		if n != 0 {
			t.Errorf("Unexpected write size")
		}
	})

	t.Run("when rawConn.Write return write size", func(t *testing.T) {
		writeSize := len(buf)
		rawConn.EXPECT().Write(buf).Return(writeSize, nil)

		n, err := conn.Write(buf)

		if err != nil {
			t.Errorf("Unexpected error %s", err)
		}

		if n != writeSize {
			t.Errorf("Unexpected write size")
		}
	})
}

func TestConn_SetDeadline(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	rawConn := NewMockConn(mockCtrl)

	headerParser := NewMockHeaderParser(mockCtrl)
	logger := NewMockLogger(mockCtrl)

	conn := proxyprotocol.NewConn(rawConn, logger, headerParser)

	deadLine := time.Now().Add(time.Second * 15)

	t.Run("when rawConn.SetDeadline return error", func(t *testing.T) {
		deadLineErr := errors.New("DeadLine error")
		rawConn.EXPECT().SetDeadline(deadLine).Return(deadLineErr)

		err := conn.SetDeadline(deadLine)

		if err != deadLineErr {
			t.Errorf("Unexpected error %s", err)
		}
	})

	t.Run("when rawConn.SetDeadline return nil", func(t *testing.T) {
		rawConn.EXPECT().SetDeadline(deadLine).Return(nil)

		err := conn.SetDeadline(deadLine)

		if err != nil {
			t.Errorf("Unexpected error %s", err)
		}
	})
}

func TestConn_SetReadDeadline(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	rawConn := NewMockConn(mockCtrl)

	headerParser := NewMockHeaderParser(mockCtrl)
	logger := NewMockLogger(mockCtrl)

	conn := proxyprotocol.NewConn(rawConn, logger, headerParser)

	deadLine := time.Now().Add(time.Second * 15)

	t.Run("when rawConn.SetReadDeadline return error", func(t *testing.T) {
		deadLineErr := errors.New("ReadDeadLine error")
		rawConn.EXPECT().SetReadDeadline(deadLine).Return(deadLineErr)

		err := conn.SetReadDeadline(deadLine)

		if err != deadLineErr {
			t.Errorf("Unexpected error %s", err)
		}
	})

	t.Run("when rawConn.SetReadDeadline return nil", func(t *testing.T) {
		rawConn.EXPECT().SetReadDeadline(deadLine).Return(nil)

		err := conn.SetReadDeadline(deadLine)

		if err != nil {
			t.Errorf("Unexpected error %s", err)
		}
	})
}

func TestConn_SetWriteDeadline(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	rawConn := NewMockConn(mockCtrl)

	headerParser := NewMockHeaderParser(mockCtrl)
	logger := NewMockLogger(mockCtrl)

	conn := proxyprotocol.NewConn(rawConn, logger, headerParser)

	deadLine := time.Now().Add(time.Second * 15)

	t.Run("when rawConn.SetWriteDeadline return error", func(t *testing.T) {
		deadLineErr := errors.New("WriteDeadLine error")
		rawConn.EXPECT().SetWriteDeadline(deadLine).Return(deadLineErr)

		err := conn.SetWriteDeadline(deadLine)

		if err != deadLineErr {
			t.Errorf("Unexpected error %s", err)
		}
	})

	t.Run("when rawConn.SetWriteDeadline return nil", func(t *testing.T) {
		rawConn.EXPECT().SetWriteDeadline(deadLine).Return(nil)

		err := conn.SetWriteDeadline(deadLine)

		if err != nil {
			t.Errorf("Unexpected error %s", err)
		}
	})
}
