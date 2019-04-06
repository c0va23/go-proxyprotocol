package proxyprotocol_test

import (
	"bufio"
	"errors"
	"net"
	"reflect"
	"testing"

	"github.com/c0va23/go-proxyprotocol"
	"github.com/golang/mock/gomock"
)

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
		parseErr := errors.New("parse error")
		headerParser.EXPECT().Parse(readBuf).Return(nil, parseErr)
		logger.EXPECT().Printf(gomock.Any(), gomock.Any()).AnyTimes()

		trustedAddr := true
		conn := proxyprotocol.NewConn(rawConn, logger, headerParser, trustedAddr)

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

		t.Run("when call conn.Read second time", func(t *testing.T) {
			_, err := conn.Read(buf)

			if err != parseErr {
				t.Errorf("Unexpected error %s", err)
			}
		})
	})

	t.Run("when header parser return Header", func(t *testing.T) {
		header := &proxyprotocol.Header{
			SrcAddr: &net.TCPAddr{
				IP: net.IPv4(1, 2, 3, 4),
			},
		}

		headerParser.EXPECT().Parse(readBuf).Return(header, nil).AnyTimes()

		t.Run("when rawConn.Read return err", func(t *testing.T) {
			readErr := errors.New("read error")
			rawConn.EXPECT().Read(buf).Return(0, readErr)

			trustedAddr := true
			conn := proxyprotocol.NewConn(rawConn, logger, headerParser, trustedAddr)

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

			trustedAddr := true
			conn := proxyprotocol.NewConn(rawConn, logger, headerParser, trustedAddr)

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

func TestConn_RemoteAddr(t *testing.T) {
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
		trustedAddr := true
		conn := proxyprotocol.NewConn(rawConn, logger, headerParser, trustedAddr)

		parseErr := errors.New("parse error")

		headerParser.EXPECT().Parse(readBuf).Return(nil, parseErr)
		logger.EXPECT().Printf(gomock.Any(), gomock.Any()).AnyTimes()

		remoteAddr := conn.RemoteAddr()

		if !reflect.DeepEqual(remoteAddr, rawAddr) {
			t.Errorf("Unexpected remote addr %s", remoteAddr)
		}
	})

	t.Run("when header parser return header", func(t *testing.T) {
		header := &proxyprotocol.Header{
			SrcAddr: &net.TCPAddr{
				IP: net.IPv4(1, 2, 3, 4),
			},
		}

		headerParser.EXPECT().Parse(readBuf).Return(header, nil).AnyTimes()

		t.Run("when trusted addr", func(t *testing.T) {
			trustedAddr := true
			conn := proxyprotocol.NewConn(rawConn, logger, headerParser, trustedAddr)

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

		t.Run("when not trusted addr", func(t *testing.T) {
			trustedAddr := false
			conn := proxyprotocol.NewConn(rawConn, logger, headerParser, trustedAddr)

			remoteAddr := conn.RemoteAddr()

			if !reflect.DeepEqual(remoteAddr, rawAddr) {
				t.Errorf("Unexpected remote adder %s", remoteAddr)
			}
		})
	})
}

func TestConn_LocalAddr(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	rawConn := NewMockConn(mockCtrl)

	headerParser := NewMockHeaderParser(mockCtrl)
	logger := NewMockLogger(mockCtrl)

	logger.EXPECT().Printf(gomock.Any(), gomock.Any()).AnyTimes()

	rawAddr := &net.TCPAddr{
		IP:   net.IPv4(10, 0, 0, 1),
		Port: 12345,
	}
	rawConn.EXPECT().LocalAddr().Return(rawAddr).AnyTimes()

	readBuf := bufio.NewReaderSize(rawConn, 1400)

	t.Run("when header parser return nil header", func(t *testing.T) {
		headerParser.EXPECT().Parse(readBuf).Return(nil, nil)

		trustedAddr := true
		conn := proxyprotocol.NewConn(rawConn, logger, headerParser, trustedAddr)

		localAddr := conn.LocalAddr()
		if localAddr != rawAddr {
			t.Errorf("Unexpected local addr %s", localAddr)
		}
	})

	t.Run("when header parser return not nil header", func(t *testing.T) {
		trustedAddr := true
		conn := proxyprotocol.NewConn(rawConn, logger, headerParser, trustedAddr)

		header := &proxyprotocol.Header{
			DstAddr: &net.TCPAddr{
				IP:   net.IPv4(1, 2, 3, 4),
				Port: 12345,
			},
		}
		headerParser.EXPECT().Parse(readBuf).Return(header, nil).AnyTimes()

		localAddr := conn.LocalAddr()
		if localAddr != header.DstAddr {
			t.Errorf("Unexpected local addr %s", localAddr)
		}

		t.Run("when second call LocalAddr()", func(t *testing.T) {
			t.Run("when trusted addr", func(t *testing.T) {
				trustedAddr := true
				conn := proxyprotocol.NewConn(rawConn, logger, headerParser, trustedAddr)

				localAddr := conn.LocalAddr()
				if localAddr != header.DstAddr {
					t.Errorf("Unexpected local addr %s", localAddr)
				}
			})

			t.Run("when trusted addr", func(t *testing.T) {
				trustedAddr := false
				conn := proxyprotocol.NewConn(rawConn, logger, headerParser, trustedAddr)

				localAddr := conn.LocalAddr()
				if localAddr != rawAddr {
					t.Errorf("Unexpected local addr %s", localAddr)
				}
			})
		})
	})
}
