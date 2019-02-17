package proxyprotocol

import (
	"bufio"
	"net"
	"time"
)

const bufferSize = 1400

// LoggerFn type of logger function
type LoggerFn func(string, ...interface{})

// NewListener create new proxyprocol.Listener from any net.Listener.
// If loggerFn not nil, then enabled debug mode.
func NewListener(listener net.Listener, loggerFn LoggerFn) net.Listener {
	return &Listener{
		listener: listener,
		loggerFn: loggerFn,
		headerParsers: []HeaderParser{
			ParseTextHeader,
			ParseV2Header,
		},
	}
}

// Listener implement net.Listener
type Listener struct {
	listener      net.Listener
	loggerFn      LoggerFn
	headerParsers []HeaderParser
}

func (listener *Listener) log(str string, args ...interface{}) {
	if nil != listener.loggerFn {
		listener.loggerFn(str, args...)
	}
}

// Accept implement net.Listener.Accept().
// If request have proxyprotocol header, then wrap connection into proxyprotocol.Conn.
// Otherwise return raw net.Conn.
func (listener *Listener) Accept() (net.Conn, error) {
	rawConn, err := listener.listener.Accept()
	if nil != err {
		return nil, err
	}

	readBuf := bufio.NewReaderSize(rawConn, bufferSize)

	var header *Header

parserLoop:
	for _, headerParser := range listener.headerParsers {
		header, err = headerParser(readBuf)
		switch err {
		case nil:
			break parserLoop
		case ErrInvalidSignature:
			continue
		default:
			return nil, err
		}
	}

	if header == nil {
		listener.log("Use raw remote addr")
	} else {
		listener.log("Use header remote addr")
	}

	return NewConn(rawConn, readBuf, header), nil
}

// Close is proxy to listener.Close()
func (listener *Listener) Close() error {
	return listener.listener.Close()
}

// Addr is proxy to listener.Addr()
func (listener Listener) Addr() net.Addr {
	return listener.listener.Addr()
}

// Conn is wrapper on net.Conn with overrided RemoteAddr()
type Conn struct {
	conn    net.Conn
	readBuf *bufio.Reader
	header  *Header
}

// NewConn create wrapper on net.Conn.
// If proxyprtocol header is local, when header should be nil.
func NewConn(conn net.Conn, readBuf *bufio.Reader, header *Header) net.Conn {
	return &Conn{
		conn:    conn,
		readBuf: readBuf,
		header:  header,
	}
}

// Read proxy to conn.Read
func (conn *Conn) Read(buf []byte) (int, error) {
	return conn.readBuf.Read(buf)
}

// Write proxy to conn.Write
func (conn *Conn) Write(buf []byte) (int, error) {
	return conn.conn.Write(buf)
}

// Close proxy to conn.Close
func (conn *Conn) Close() error {
	return conn.conn.Close()
}

// LocalAddr proxy to conn.LocalAddr
func (conn *Conn) LocalAddr() net.Addr {
	return conn.conn.LocalAddr()
}

// RemoteAddr return addr of remote client.
// If proxyprtocol not local, then return src from header.
func (conn *Conn) RemoteAddr() net.Addr {
	if nil != conn.header {
		return conn.header.SrcAddr
	}
	return conn.conn.RemoteAddr()
}

// SetDeadline proxy to conn.SetDeadline
func (conn *Conn) SetDeadline(t time.Time) error {
	return conn.conn.SetDeadline(t)
}

// SetReadDeadline proxy to conn.SetReadDeadline
func (conn *Conn) SetReadDeadline(t time.Time) error {
	return conn.conn.SetReadDeadline(t)
}

// SetWriteDeadline  proxy to conn.SetWriteDeadline
func (conn *Conn) SetWriteDeadline(t time.Time) error {
	return conn.conn.SetWriteDeadline(t)
}
