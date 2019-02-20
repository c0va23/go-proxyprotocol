package proxyprotocol

import (
	"bufio"
	"net"
	"time"
)

const bufferSize = 1400

// LoggerFn type of logger function
type LoggerFn func(string, ...interface{})

// SourceChecker check trusted address
type SourceChecker func(net.Addr) (bool, error)

// NewListener create new proxyprocol.Listener from any net.Listener.
func NewListener(listener net.Listener) *Listener {
	return &Listener{
		Listener: listener,
		HeaderParsers: []HeaderParser{
			ParseTextHeader,
			ParseBinaryHeader,
		},
	}
}

// Listener implement net.Listener
type Listener struct {
	Listener      net.Listener
	LoggerFn      LoggerFn
	HeaderParsers []HeaderParser
	SourceCheck   SourceChecker
}

func (listener *Listener) log(str string, args ...interface{}) {
	if nil != listener.LoggerFn {
		listener.LoggerFn(str, args...)
	}
}

// WithLogger copy Listener and set LoggerFn
func (listener *Listener) WithLogger(loggerFn LoggerFn) *Listener {
	newListener := *listener
	newListener.LoggerFn = loggerFn
	return &newListener
}

// WithHeaderParsers copy Listener and set HeaderParser.
// Can be used to disable or reorder HeaderParser's.
func (listener *Listener) WithHeaderParsers(headerParser ...HeaderParser) *Listener {
	newListener := *listener
	newListener.HeaderParsers = headerParser
	return &newListener
}

// WithSourceChecker copy Listener and set SourceChecker
func (listener *Listener) WithSourceChecker(sourceChecker SourceChecker) *Listener {
	newListener := *listener
	newListener.SourceCheck = sourceChecker
	return &newListener
}

func (listener *Listener) parserHeader(readBuf *bufio.Reader) (*Header, error) {
	for _, headerParser := range listener.HeaderParsers {
		header, err := headerParser(readBuf, listener.log)
		switch err {
		case nil:
			listener.log("Use header remote addr")
			return header, nil
		case ErrInvalidSignature:
			continue
		default:
			listener.log("Parse header error: %s", err)
			return nil, err
		}
	}
	listener.log("Use raw remote addr")
	return nil, nil
}

// Accept implement net.Listener.Accept().
// If request have proxyprotocol header, then wrap connection into proxyprotocol.Conn.
// Otherwise return raw net.Conn.
func (listener *Listener) Accept() (net.Conn, error) {
	rawConn, err := listener.Listener.Accept()
	if nil != err {
		return nil, err
	}

	if listener.SourceCheck != nil {
		trusted, err := listener.SourceCheck(rawConn.RemoteAddr())
		if nil != err {
			listener.log("Source check error: %s", err)
			return nil, err
		}
		if !trusted {
			return rawConn, nil
		}
	}

	readBuf := bufio.NewReaderSize(rawConn, bufferSize)

	header, err := listener.parserHeader(readBuf)
	if nil != err {
		return nil, err
	}

	return NewConn(rawConn, readBuf, header), nil
}

// Close is proxy to listener.Close()
func (listener *Listener) Close() error {
	return listener.Listener.Close()
}

// Addr is proxy to listener.Addr()
func (listener Listener) Addr() net.Addr {
	return listener.Listener.Addr()
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
