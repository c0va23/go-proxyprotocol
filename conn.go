package proxyprotocol

import (
	"bufio"
	"net"
	"sync"
	"time"
)

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
