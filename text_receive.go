package proxyprotocol

import (
	"bufio"
	"bytes"
	"strings"
)

// ParseTextHeader try parse proxyprotocol header.
func ParseTextHeader(buf *bufio.Reader) (*Header, error) {
	signatureBuf, err := buf.Peek(TextSignatureLen)
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
	default:
		return nil, ErrUnknownProtocol
	}
}
