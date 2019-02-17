package proxyprotocol

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"net"
)

// Errors
var (
	ErrUnknownVersion       = errors.New("Unknown version")
	ErrUnknownCommand       = errors.New("Unknown command")
	ErrUnexpectedAddressLen = errors.New("Unexpected address length")
)

// ParseV2Header from incoming connection
func ParseV2Header(buf *bufio.Reader) (*Header, error) {
	magicBuf, err := buf.Peek(BinarySignatueLen)
	if nil != err {
		return nil, err
	}

	if !bytes.Equal(magicBuf, BinarySignatue) {
		return nil, ErrInvalidSignature
	}

	buf.Discard(BinarySignatueLen)

	versionCommandByte, err := buf.ReadByte()
	if nil != err {
		return nil, err
	}

	if versionCommandByte&BinaryVersionMask != BinaryVersion2 {
		return nil, ErrUnknownVersion
	}

	switch versionCommandByte & BinaryCommandMask {
	case BinaryCommandProxy:
		protocol, err := buf.ReadByte()
		if nil != err {
			return nil, err
		}

		addressSizeBuf := make([]byte, 2)
		_, err = buf.Read(addressSizeBuf)
		if nil != err {
			return nil, err
		}
		addressesLen := int(binary.BigEndian.Uint16(addressSizeBuf))

		addressesBuf := make([]byte, addressesLen)
		_, err = buf.Read(addressesBuf)
		if nil != err {
			return nil, err
		}

		var header *Header
		switch protocol & BinaryAFMask {
		case BinaryProtocolUnspec:
			return nil, nil
		case BinaryAFInet:
			if addressesLen != BinaryAddressLenIPv4 {
				return nil, ErrUnexpectedAddressLen
			}

			srcIP := make(net.IP, net.IPv4len)
			copy(srcIP, addressesBuf[0:net.IPv4len])

			dstIP := make(net.IP, net.IPv4len)
			copy(dstIP, addressesBuf[net.IPv4len:2*net.IPv4len])

			srcPort := binary.BigEndian.Uint16(addressesBuf[2*net.IPv4len : 2*net.IPv4len+binaryPortSize])
			dstPort := binary.BigEndian.Uint16(addressesBuf[2*net.IPv4len+binaryPortSize : 2*net.IPv4len+2*binaryPortSize])

			header = &Header{
				SrcAddr: &net.TCPAddr{
					IP:   srcIP,
					Port: int(srcPort),
				},
				DstAddr: &net.TCPAddr{
					IP:   dstIP,
					Port: int(dstPort),
				},
			}
			return header, nil
		case BinaryAFInet6:
			if addressesLen != BinaryAddressLenIPv6 {
				return nil, ErrUnexpectedAddressLen
			}

			srcIP := make(net.IP, net.IPv6len)
			copy(srcIP, addressesBuf[0:net.IPv6len])

			dstIP := make(net.IP, net.IPv6len)
			copy(dstIP, addressesBuf[net.IPv6len:2*net.IPv6len])

			srcPort := binary.BigEndian.Uint16(addressesBuf[2*net.IPv6len : 2*net.IPv6len+binaryPortSize])
			dstPort := binary.BigEndian.Uint16(addressesBuf[2*net.IPv6len+binaryPortSize : 2*net.IPv6len+2*binaryPortSize])

			header = &Header{
				SrcAddr: &net.TCPAddr{
					IP:   srcIP,
					Port: int(srcPort),
				},
				DstAddr: &net.TCPAddr{
					IP:   dstIP,
					Port: int(dstPort),
				},
			}
			return header, nil
		default:
			return nil, ErrUnknownProtocol
		}
	case BinaryCommandLocal:
		return nil, nil
	default:
		return nil, ErrUnknownCommand
	}
}
