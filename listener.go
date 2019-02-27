package proxyprotocol

import (
	"bufio"
	"net"
	"sync"
	"time"
)

const bufferSize = 1400

// SourceChecker check trusted address
type SourceChecker func(net.Addr) (bool, error)

// NewListener create new proxyprocol.Listener from any net.Listener.
func NewListener(listener net.Listener) *Listener {
	return &Listener{
		Listener:            listener,
		HeaderParserBuilder: DefaultFallbackHeaderParserBuilder(),
	}
}

// Listener implement net.Listener
type Listener struct {
	net.Listener
	Logger
	HeaderParserBuilder
	SourceChecker
}

// WithLogger copy Listener and set LoggerFn
func (listener *Listener) WithLogger(logger Logger) *Listener {
	newListener := *listener
	newListener.Logger = logger
	return &newListener
}

// WithHeaderParserBuilder copy Listener and set HeaderParserBuilder.
// Can be used to disable or reorder HeaderParser's.
func (listener *Listener) WithHeaderParserBuilder(
	headerParserBuilder HeaderParserBuilder,
) *Listener {
	newListener := *listener
	newListener.HeaderParserBuilder = headerParserBuilder
	return &newListener
}

// WithSourceChecker copy Listener and set SourceChecker
func (listener *Listener) WithSourceChecker(sourceChecker SourceChecker) *Listener {
	newListener := *listener
	newListener.SourceChecker = sourceChecker
	return &newListener
}

// Accept implement net.Listener.Accept().
// If request have proxyprotocol header, then wrap connection into proxyprotocol.Conn.
// Otherwise return raw net.Conn.
func (listener *Listener) Accept() (net.Conn, error) {
	rawConn, err := listener.Listener.Accept()
	if nil != err {
		return nil, err
	}

	logger := FallbackLogger{Logger: listener.Logger}
	if listener.SourceChecker != nil {
		trusted, err := listener.SourceChecker(rawConn.RemoteAddr())
		if nil != err {
			logger.Printf("Source check error: %s", err)
			return nil, err
		}
		if !trusted {
			return rawConn, nil
		}
	}

	headerParser := listener.HeaderParserBuilder.Build(logger)

	return NewConn(rawConn, logger, headerParser), nil
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
	conn         net.Conn
	logger       Logger
	readBuf      *bufio.Reader
	header       *Header
	headerParser HeaderParser
	once         sync.Once
}

// NewConn create wrapper on net.Conn.
// If proxyprtocol header is local, when header should be nil.
func NewConn(
	conn net.Conn,
	logger Logger,
	headerParser HeaderParser,
) net.Conn {
	readBuf := bufio.NewReaderSize(conn, bufferSize)

	return &Conn{
		conn:         conn,
		readBuf:      readBuf,
		logger:       logger,
		headerParser: headerParser,
	}
}

// Read proxy to conn.Read
func (conn *Conn) Read(buf []byte) (int, error) {
	var err error
	conn.once.Do(func() {
		conn.header, err = conn.headerParser.Parse(conn.readBuf)
	})
	if nil != err {
		return 0, err
	}
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
	conn.once.Do(func() {
		var err error
		conn.header, err = conn.headerParser.Parse(conn.readBuf)
		if nil != err {
			conn.logger.Printf("Header parse error: %s", err)
		}
	})
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
