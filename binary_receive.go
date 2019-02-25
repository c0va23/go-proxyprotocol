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

// Meta buffer byte positoin
const (
	versionCommandPos  = 0
	protocolPos        = 1
	addressLenStartPos = 2
	addressLenEndPos   = 4
)

// BinaryHeaderParser parse proxyprotocol header from Reader
type BinaryHeaderParser struct {
	logf LoggerFn
}

// NewBinaryHeaderParser create new instance BinaryHeaderParser
func NewBinaryHeaderParser(logf LoggerFn) BinaryHeaderParser {
	return BinaryHeaderParser{
		logf: logf,
	}
}

// Parse buffer
func (parser BinaryHeaderParser) Parse(buf *bufio.Reader) (*Header, error) {
	magicBuf, err := buf.Peek(BinarySignatueLen)
	if nil != err {
		parser.logf("Read magif prefix error: %s", err)
		return nil, err
	}

	if !bytes.Equal(magicBuf, BinarySignatue) {
		return nil, ErrInvalidSignature
	}

	buf.Discard(BinarySignatueLen)

	metaBuf := make([]byte, addressLenEndPos)
	if _, err = buf.Read(metaBuf); nil != err {
		parser.logf("Read meta error: %s", err)
		return nil, err
	}

	versionCommandByte := metaBuf[versionCommandPos]

	if versionCommandByte&BinaryVersionMask != BinaryVersion2 {
		return nil, ErrUnknownVersion
	}

	addressSizeBuf := metaBuf[addressLenStartPos:addressLenEndPos]
	addressesLen := int(binary.BigEndian.Uint16(addressSizeBuf))
	parser.logf("Addresses len: %d", addressesLen)

	addressesBuf := make([]byte, addressesLen)
	addressReaded, err := buf.Read(addressesBuf)
	if nil != err {
		parser.logf("Read address error: %s", err)
		return nil, err
	}
	parser.logf("Address readed: %d", addressReaded)

	switch versionCommandByte & BinaryCommandMask {
	case BinaryCommandProxy:
		protocol := metaBuf[protocolPos]

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

// BinaryHeaderParserBuilder build BinaryHeaderParser
type BinaryHeaderParserBuilder struct{}

// Build BinaryHeaderParser
func (builder *BinaryHeaderParserBuilder) Build(logf LoggerFn) HeaderParser {
	return NewBinaryHeaderParser(logf)
}
