// Package proxyprotocol impleement HA ProxyProtocol V1 and V2 for receiver
//
// Prorxyprotocol spec http://www.haproxy.org/download/2.0/doc/proxy-protocol.txt
package proxyprotocol

import "net"

// Header struct represent header parsing result
type Header struct {
	SrcAddr net.Addr
	DstAddr net.Addr
}
