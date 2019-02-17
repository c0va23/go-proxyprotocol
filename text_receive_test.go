package proxyprotocol_test

import (
	"io"

	"github.com/c0va23/go-proxyprotocol"

	"testing"
)

func TestParseTextHeader(t *testing.T) {
	t.Run("buffer EOF", func(t *testing.T) {
		data := []byte{}
		testParser(t, proxyprotocol.ParseTextHeader, nil, io.EOF, data)
	})

	t.Run("invalid signature", func(t *testing.T) {
		data := []byte("GET / HTTP/1.0")
		testParser(t, proxyprotocol.ParseTextHeader, nil, proxyprotocol.ErrInvalidSignature, data)
	})

	t.Run("without LineFeed (\\n)", func(t *testing.T) {
		data := append(proxyprotocol.TextSignature, ' ')
		testParser(t, proxyprotocol.ParseTextHeader, nil, io.EOF, data)
	})

	t.Run("Invalid protocol", func(t *testing.T) {
		data := append(proxyprotocol.TextSignature, []byte(proxyprotocol.TextSeparator)...)
		data = append(data, []byte("INVALID")...)
		data = append(data, proxyprotocol.TextCRLF...)
		testParser(t, proxyprotocol.ParseTextHeader, nil, proxyprotocol.ErrUnknownProtocol, data)
	})

	t.Run("unknown protocol", func(t *testing.T) {
		data := append(proxyprotocol.TextSignature, []byte(proxyprotocol.TextSeparator)...)
		data = append(data, []byte(proxyprotocol.TextProtocolUnknown)...)
		data = append(data, proxyprotocol.TextCRLF...)
		testParser(t, proxyprotocol.ParseTextHeader, nil, nil, data)
	})
}
