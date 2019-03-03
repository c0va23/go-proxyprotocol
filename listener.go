package proxyprotocol

import (
	"net"
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
