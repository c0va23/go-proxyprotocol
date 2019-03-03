// Package proxyprotocol impleement HA ProxyProtocol V1 and V2 for receiver
//
// Prorxyprotocol spec http://www.haproxy.org/download/2.0/doc/proxy-protocol.txt
package proxyprotocol

import (
	"bufio"
	"errors"
	"net"
)

// Header struct represent header parsing result
type Header struct {
	SrcAddr net.Addr
	DstAddr net.Addr
}

// HeaderParserBuilder build HeaderParser's
type HeaderParserBuilder interface {
	Build(Logger) HeaderParser
}

// HeaderParser describe interface for header parsers
type HeaderParser interface {
	Parse(readBuf *bufio.Reader) (*Header, error)
}

// Shared HeaderParser errors
var (
	ErrInvalidSignature = errors.New("invalid signature")
	ErrUnknownProtocol  = errors.New("unknown protocol")
)
