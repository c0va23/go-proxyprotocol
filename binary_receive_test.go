package proxyprotocol_test

import (
	"encoding/binary"
	"io"
	"net"

	"github.com/c0va23/go-proxyprotocol"

	"testing"
)

func TestParseV2Header(t *testing.T) {
	t.Run("signature EOF", func(t *testing.T) {
		testParser(t, testParserArgs{
			headerParser: proxyprotocol.ParseBinaryHeader,
			data:         []byte{},
			err:          io.EOF,
		})
	})

	t.Run("Invalid signature", func(t *testing.T) {
		data := make([]byte, proxyprotocol.BinarySignatueLen)
		testParser(t, testParserArgs{
			headerParser: proxyprotocol.ParseBinaryHeader,
			data:         data,
			err:          proxyprotocol.ErrInvalidSignature,
		})
	})

	t.Run("meta EOF", func(t *testing.T) {
		data := append(proxyprotocol.BinarySignatue)
		testParser(t, testParserArgs{
			headerParser: proxyprotocol.ParseBinaryHeader,
			data:         data,
			err:          io.EOF,
		})
	})

	t.Run("Invalid version", func(t *testing.T) {
		invalidVersoin := byte(0x00)
		data := append(proxyprotocol.BinarySignatue, invalidVersoin)
		testParser(t, testParserArgs{
			headerParser: proxyprotocol.ParseBinaryHeader,
			data:         data,
			err:          proxyprotocol.ErrUnknownVersion,
		})
	})

	t.Run("Invalid command", func(t *testing.T) {
		invalidCommand := proxyprotocol.BinaryVersion2&proxyprotocol.BinaryVersionMask | proxyprotocol.BinaryCommandMask&0xFF
		t.Logf("Version command bits: %02x", invalidCommand)
		data := append(proxyprotocol.BinarySignatue, invalidCommand)
		testParser(t, testParserArgs{
			headerParser: proxyprotocol.ParseBinaryHeader,
			data:         data,
			err:          proxyprotocol.ErrUnknownCommand,
		})
	})

	t.Run("Local command", func(t *testing.T) {
		commandVerison := proxyprotocol.BinaryCommandLocal | proxyprotocol.BinaryVersion2
		protocol := proxyprotocol.BinaryProtocolUnspec
		data := append(proxyprotocol.BinarySignatue, commandVerison, protocol, 0, 0)
		testParser(t, testParserArgs{
			headerParser: proxyprotocol.ParseBinaryHeader,
			data:         data,
			header:       nil,
			err:          nil,
		})
	})

	t.Run("Invalid protocol", func(t *testing.T) {
		commandVerison := proxyprotocol.BinaryCommandProxy | proxyprotocol.BinaryVersion2
		invalidProtocol := byte(0xFF)
		data := append(proxyprotocol.BinarySignatue, commandVerison, invalidProtocol, 0, 0)
		testParser(t, testParserArgs{
			headerParser: proxyprotocol.ParseBinaryHeader,
			data:         data,
			err:          proxyprotocol.ErrUnknownProtocol,
		})
	})

	t.Run("Unspec protocol", func(t *testing.T) {
		commandVerison := proxyprotocol.BinaryCommandProxy | proxyprotocol.BinaryVersion2
		data := append(proxyprotocol.BinarySignatue, commandVerison, proxyprotocol.BinaryProtocolUnspec, 0, 0)
		testParser(t, testParserArgs{
			headerParser: proxyprotocol.ParseBinaryHeader,
			data:         data,
			header:       nil,
			err:          nil,
		})
	})

	t.Run("TCPv4 protocol", func(t *testing.T) {
		commandVerison := proxyprotocol.BinaryCommandProxy | proxyprotocol.BinaryVersion2
		data := append(proxyprotocol.BinarySignatue, commandVerison, proxyprotocol.BinaryProtocolTCPoverIPv4)

		t.Run("Invalid address size", func(t *testing.T) {
			data := append(data, 0, 0)
			testParser(t, testParserArgs{
				headerParser: proxyprotocol.ParseBinaryHeader,
				data:         data,
				err:          proxyprotocol.ErrUnexpectedAddressLen,
			})
		})

		t.Run("Valid address size", func(t *testing.T) {
			addressDataLen := 2 * (net.IPv4len + proxyprotocol.BinaryPortLen)
			data := append(data, 0, byte(addressDataLen))

			t.Run("Invalid address data size", func(t *testing.T) {
				testParser(t, testParserArgs{
					headerParser: proxyprotocol.ParseBinaryHeader,
					data:         data,
					err:          io.EOF,
				})
			})

			t.Run("Valid address data size", func(t *testing.T) {
				srcAddr := net.IP{192, 168, 1, 2}
				dstAddr := net.IP{10, 0, 0, 2}

				srcPort := 12345
				dstPort := 8080

				srcPortBuf := make([]byte, 2)
				dstPortBuf := make([]byte, 2)

				binary.BigEndian.PutUint16(srcPortBuf, uint16(srcPort))
				binary.BigEndian.PutUint16(dstPortBuf, uint16(dstPort))

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
				data := append(data, srcAddr...)
				data = append(data, dstAddr...)
				data = append(data, srcPortBuf...)
				data = append(data, dstPortBuf...)
				testParser(t, testParserArgs{
					headerParser: proxyprotocol.ParseBinaryHeader,
					data:         data,
					header:       &expectedHeader,
				})
			})
		})
	})

	t.Run("TCPv6 protocol", func(t *testing.T) {
		commandVerison := proxyprotocol.BinaryCommandProxy | proxyprotocol.BinaryVersion2
		data := append(proxyprotocol.BinarySignatue, commandVerison, proxyprotocol.BinaryProtocolTCPoverIPv6)

		t.Run("Invalid address size", func(t *testing.T) {
			data := append(data, 0, 0)
			testParser(t, testParserArgs{
				headerParser: proxyprotocol.ParseBinaryHeader,
				data:         data,
				err:          proxyprotocol.ErrUnexpectedAddressLen,
			})
		})

		t.Run("Valid address size", func(t *testing.T) {
			addressDataLen := 2 * (net.IPv6len + proxyprotocol.BinaryPortLen)
			data := append(data, 0, byte(addressDataLen))

			t.Run("Invalid address data size", func(t *testing.T) {
				testParser(t, testParserArgs{
					headerParser: proxyprotocol.ParseBinaryHeader,
					data:         data,
					err:          io.EOF,
				})
			})

			t.Run("Valid address data size", func(t *testing.T) {
				srcAddr := net.ParseIP("::1")
				dstAddr := net.ParseIP("::2")

				srcPort := 12345
				dstPort := 8080

				srcPortBuf := make([]byte, 2)
				dstPortBuf := make([]byte, 2)

				binary.BigEndian.PutUint16(srcPortBuf, uint16(srcPort))
				binary.BigEndian.PutUint16(dstPortBuf, uint16(dstPort))

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
				data := append(data, srcAddr...)
				data = append(data, dstAddr...)
				data = append(data, srcPortBuf...)
				data = append(data, dstPortBuf...)
				testParser(t, testParserArgs{
					headerParser: proxyprotocol.ParseBinaryHeader,
					data:         data,
					header:       &expectedHeader,
				})
			})
		})

		t.Run("address data with TLV", func(t *testing.T) {
			tlvTypeLen := 1
			tlvLengthLen := 2
			tlvLen := tlvTypeLen + tlvLengthLen
			dataLen := 2 * (net.IPv6len + proxyprotocol.BinaryPortLen)
			data := append(data, 0, byte(dataLen+tlvLen))

			t.Run("address data with TLV data", func(t *testing.T) {
				srcAddr := net.ParseIP("::1")
				dstAddr := net.ParseIP("::2")

				srcPort := 12345
				dstPort := 8080

				srcPortBuf := make([]byte, 2)
				dstPortBuf := make([]byte, 2)

				binary.BigEndian.PutUint16(srcPortBuf, uint16(srcPort))
				binary.BigEndian.PutUint16(dstPortBuf, uint16(dstPort))

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
				data := append(data, srcAddr...)
				data = append(data, dstAddr...)
				data = append(data, srcPortBuf...)
				data = append(data, dstPortBuf...)
				data = append(data, proxyprotocol.TLVTypeNoop, 0, 0)
				testParser(t, testParserArgs{
					headerParser: proxyprotocol.ParseBinaryHeader,
					data:         data,
					header:       &expectedHeader,
				})
			})
		})
	})
}
