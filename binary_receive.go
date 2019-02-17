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

// ParseBinaryHeader from incoming connection
func ParseBinaryHeader(buf *bufio.Reader) (*Header, error) {
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

		switch protocol & BinaryAFMask {
		case BinaryProtocolUnspec:
			return nil, nil
		case BinaryAFInet:
			return parseAddressData(addressesBuf, net.IPv4len)
		case BinaryAFInet6:
			return parseAddressData(addressesBuf, net.IPv6len)
		default:
			return nil, ErrUnknownProtocol
		}
	case BinaryCommandLocal:
		return nil, nil
	default:
		return nil, ErrUnknownCommand
	}
}

func parseAddressData(addressesBuf []byte, IPLen int) (*Header, error) {
	expectedBufSize := 2 * (IPLen + BinaryPortLen)
	if len(addressesBuf) < expectedBufSize {
		return nil, ErrUnexpectedAddressLen
	}

	srcIP := make(net.IP, IPLen)
	copy(srcIP, addressesBuf[:IPLen])
	addressesBuf = addressesBuf[IPLen:]

	dstIP := make(net.IP, IPLen)
	copy(dstIP, addressesBuf[:IPLen])
	addressesBuf = addressesBuf[IPLen:]

	srcPort := binary.BigEndian.Uint16(addressesBuf[:BinaryPortLen])
	addressesBuf = addressesBuf[BinaryPortLen:]

	dstPort := binary.BigEndian.Uint16(addressesBuf[:BinaryPortLen])
	addressesBuf = addressesBuf[BinaryPortLen:]

	return &Header{
		SrcAddr: &net.TCPAddr{
			IP:   srcIP,
			Port: int(srcPort),
		},
		DstAddr: &net.TCPAddr{
			IP:   dstIP,
			Port: int(dstPort),
		},
	}, nil

}
