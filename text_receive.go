package proxyprotocol

import (
	"bufio"
	"bytes"
	"errors"
	"net"
	"strconv"
	"strings"
)

// Text protocol errors
var (
	ErrInvalidAddressList = errors.New("Invalid address list")
	ErrInvalidIP          = errors.New("Invalid IP")
	ErrInvalidPort        = errors.New("Invalid port")
)

// ParseTextHeader try parse proxyprotocol header.
func ParseTextHeader(buf *bufio.Reader) (*Header, error) {
	signatureBuf, err := buf.Peek(textSignatureLen)
	if nil != err {
		return nil, err
	}

	if !bytes.Equal(signatureBuf, TextSignature) {
		return nil, ErrInvalidSignature
	}

	headerLine, err := buf.ReadString(TextLF)
	if nil != err {
		return nil, err
	}

	// Strip CR char on line end
	if headerLine[len(headerLine)-2] == TextCR {
		headerLine = headerLine[:len(headerLine)-2]
	}

	headerParts := strings.Split(headerLine, TextSeparator)

	protocol := headerParts[1]

	switch protocol {
	case TextProtocolUnknown:
		return nil, nil
	case TextProtocolIPv4:
		addressParts := headerParts[2:]
		if textAddressPartsLen != len(addressParts) {
			return nil, ErrInvalidAddressList
		}

		srcIPStr := addressParts[0]
		srcIP := net.ParseIP(srcIPStr)
		if nil == srcIP {
			return nil, ErrInvalidIP
		}

		dstIPStr := addressParts[1]
		dstIP := net.ParseIP(dstIPStr)
		if nil == dstIP {
			return nil, ErrInvalidIP
		}

		srcPortSrt := addressParts[2]
		srcPort, err := strconv.ParseUint(srcPortSrt, 10, textPortBitSize)
		if nil != err {
			return nil, ErrInvalidPort
		}

		dstPortSrt := addressParts[3]
		dstPort, err := strconv.ParseUint(dstPortSrt, 10, textPortBitSize)
		if nil != err {
			return nil, ErrInvalidPort
		}

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
	default:
		return nil, ErrUnknownProtocol
	}
}
