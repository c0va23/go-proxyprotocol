package proxyprotocol_test

import (
	"io"
	"net"
	"strconv"

	"github.com/c0va23/go-proxyprotocol"

	"testing"
)

func buildTextHeader(prefix []byte, parts ...string) []byte {
	data := prefix
	for _, part := range parts {
		data = append(data, []byte(proxyprotocol.TextSeparator)...)
		data = append(data, []byte(part)...)
	}
	data = append(data, proxyprotocol.TextCRLF...)
	return data
}

func TestParseTextHeader(t *testing.T) {
	textHeaderParser := proxyprotocol.NewTextHeaderParser(t.Logf)
	t.Run("buffer EOF", func(t *testing.T) {
		data := []byte{}
		testParser(t, testParserArgs{
			headerParser: textHeaderParser,
			data:         data,
			header:       nil,
			err:          io.EOF,
		})
	})

	t.Run("invalid signature", func(t *testing.T) {
		data := []byte("GET / HTTP/1.0")
		testParser(t, testParserArgs{
			headerParser: textHeaderParser,
			data:         data,
			err:          proxyprotocol.ErrInvalidSignature,
		})
	})

	t.Run("without LineFeed (\\n)", func(t *testing.T) {
		data := append(proxyprotocol.TextSignature, ' ')
		testParser(t, testParserArgs{
			headerParser: textHeaderParser,
			data:         data,
			err:          io.EOF,
		})
	})

	t.Run("Invalid protocol", func(t *testing.T) {
		data := append(proxyprotocol.TextSignature, []byte(proxyprotocol.TextSeparator)...)
		data = append(data, []byte("INVALID")...)
		data = append(data, proxyprotocol.TextCRLF...)
		testParser(t, testParserArgs{
			headerParser: textHeaderParser,
			data:         data,
			err:          proxyprotocol.ErrUnknownProtocol,
		})
	})

	t.Run("unknown protocol", func(t *testing.T) {
		data := append(proxyprotocol.TextSignature, []byte(proxyprotocol.TextSeparator)...)
		data = append(data, []byte(proxyprotocol.TextProtocolUnknown)...)
		data = append(data, proxyprotocol.TextCRLF...)
		testParser(t, testParserArgs{
			headerParser: textHeaderParser,
			data:         data,
			header:       nil,
			err:          nil,
		})
	})

	t.Run("IPV4 protocol", func(t *testing.T) {
		data := append(proxyprotocol.TextSignature, []byte(proxyprotocol.TextSeparator)...)
		data = append(data, []byte(proxyprotocol.TextProtocolIPv4)...)

		t.Run("invalid address parts", func(t *testing.T) {
			data := buildTextHeader(data, "192.168.1.2", "192.168.1.3")
			testParser(t, testParserArgs{
				headerParser: textHeaderParser,
				data:         data,
				err:          proxyprotocol.ErrInvalidAddressList,
			})
		})

		t.Run("invalid src IP", func(t *testing.T) {
			data := buildTextHeader(data, "192.168.1", "192.168.1.3", "1080", "12345")
			testParser(t, testParserArgs{
				headerParser: textHeaderParser,
				data:         data,
				err:          proxyprotocol.ErrInvalidIP,
			})
		})

		t.Run("invalid src port", func(t *testing.T) {
			data := buildTextHeader(data, "192.168.1.1", "192.168.1.3", "808080", "12345")
			testParser(t, testParserArgs{
				headerParser: textHeaderParser,
				data:         data,
				err:          proxyprotocol.ErrInvalidPort,
			})
		})

		t.Run("invalid dst IP", func(t *testing.T) {
			data := buildTextHeader(data, "192.168.1.1", "192.168.1", "1080", "12345")
			testParser(t, testParserArgs{
				headerParser: textHeaderParser,
				data:         data,
				err:          proxyprotocol.ErrInvalidIP,
			})
		})

		t.Run("invalid dst port", func(t *testing.T) {
			data := buildTextHeader(data, "192.168.1.1", "192.168.1.3", "1080", "123456")
			testParser(t, testParserArgs{
				headerParser: textHeaderParser,
				data:         data,
				err:          proxyprotocol.ErrInvalidPort,
			})
		})

		t.Run("valid address", func(t *testing.T) {
			srcAddr := net.IPv4(192, 168, 1, 2)
			dstAddr := net.IPv4(10, 0, 0, 2)

			srcPort := 12345
			dstPort := 8080

			expectedHeader := proxyprotocol.Header{
				SrcAddr: &net.TCPAddr{
					IP:   srcAddr,
					Port: srcPort,
				},
				DstAddr: &net.TCPAddr{
					IP:   dstAddr,
					Port: dstPort,
				},
			}

			data := buildTextHeader(data, srcAddr.String(), dstAddr.String(), strconv.Itoa(srcPort), strconv.Itoa(dstPort))
			testParser(t, testParserArgs{
				headerParser: textHeaderParser,
				data:         data,
				header:       &expectedHeader,
			})
		})
	})

	t.Run("IPV6 protocol", func(t *testing.T) {
		data := append(proxyprotocol.TextSignature, []byte(proxyprotocol.TextSeparator)...)
		data = append(data, []byte(proxyprotocol.TextProtocolIPv6)...)

		t.Run("invalid address parts", func(t *testing.T) {
			data := buildTextHeader(data, "::1", "::2")
			testParser(t, testParserArgs{
				headerParser: textHeaderParser,
				data:         data,
				err:          proxyprotocol.ErrInvalidAddressList,
			})
		})

		t.Run("invalid src IP", func(t *testing.T) {
			data := buildTextHeader(data, "::ZZ", "::2", "1080", "12345")
			testParser(t, testParserArgs{
				headerParser: textHeaderParser,
				data:         data,
				err:          proxyprotocol.ErrInvalidIP,
			})
		})

		t.Run("invalid src port", func(t *testing.T) {
			data := buildTextHeader(data, "::1", "::2", "808080", "12345")
			testParser(t, testParserArgs{
				headerParser: textHeaderParser,
				data:         data,
				err:          proxyprotocol.ErrInvalidPort,
			})
		})

		t.Run("invalid dst IP", func(t *testing.T) {
			data := buildTextHeader(data, "::1", "::ZZ", "1080", "12345")
			testParser(t, testParserArgs{
				headerParser: textHeaderParser,
				data:         data,
				err:          proxyprotocol.ErrInvalidIP,
			})
		})

		t.Run("invalid dst port", func(t *testing.T) {
			data := buildTextHeader(data, "::1", "::2", "1080", "123456")
			testParser(t, testParserArgs{
				headerParser: textHeaderParser,
				data:         data,
				err:          proxyprotocol.ErrInvalidPort,
			})
		})

		t.Run("valid address", func(t *testing.T) {
			srcAddr := net.ParseIP("::1")
			dstAddr := net.ParseIP("::2")

			srcPort := 12345
			dstPort := 8080

			expectedHeader := proxyprotocol.Header{
				SrcAddr: &net.TCPAddr{
					IP:   srcAddr,
					Port: srcPort,
				},
				DstAddr: &net.TCPAddr{
					IP:   dstAddr,
					Port: dstPort,
				},
			}

			data := buildTextHeader(data, srcAddr.String(), dstAddr.String(), strconv.Itoa(srcPort), strconv.Itoa(dstPort))
			testParser(t, testParserArgs{
				headerParser: textHeaderParser,
				data:         data,
				header:       &expectedHeader,
			})
		})
	})
}
