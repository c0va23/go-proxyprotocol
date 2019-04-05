package proxyprotocol_test

import (
	"encoding/binary"
	"io"
	"net"
	"testing"

	"github.com/c0va23/go-proxyprotocol"
)

func TestParseV2Header(t *testing.T) {
	logger := proxyprotocol.LoggerFunc(t.Logf)
	binaryHeaderParser := proxyprotocol.NewBinaryHeaderParser(logger)
	t.Run("signature EOF", func(t *testing.T) {
		testParser(t, testParserArgs{
			headerParser: binaryHeaderParser,
			data:         []byte{},
			err:          io.EOF,
		})
	})

	t.Run("Invalid signature", func(t *testing.T) {
		data := make([]byte, proxyprotocol.BinarySignatureLen)
		testParser(t, testParserArgs{
			headerParser: binaryHeaderParser,
			data:         data,
			err:          proxyprotocol.ErrInvalidSignature,
		})
	})

	t.Run("meta EOF", func(t *testing.T) {
		data := append(proxyprotocol.BinarySignature)
		testParser(t, testParserArgs{
			headerParser: binaryHeaderParser,
			data:         data,
			err:          io.EOF,
		})
	})

	t.Run("Invalid version", func(t *testing.T) {
		invalidVersion := byte(0x00)
		data := append(proxyprotocol.BinarySignature, invalidVersion)
		testParser(t, testParserArgs{
			headerParser: binaryHeaderParser,
			data:         data,
			err:          proxyprotocol.ErrUnknownVersion,
		})
	})

	t.Run("Invalid command", func(t *testing.T) {
		invalidCommand := proxyprotocol.BinaryVersion2&proxyprotocol.BinaryVersionMask | proxyprotocol.BinaryCommandMask&0xFF
		t.Logf("Version command bits: %02x", invalidCommand)
		data := append(proxyprotocol.BinarySignature, invalidCommand)
		testParser(t, testParserArgs{
			headerParser: binaryHeaderParser,
			data:         data,
			err:          proxyprotocol.ErrUnknownCommand,
		})
	})

	t.Run("Local command", func(t *testing.T) {
		commandVersion := proxyprotocol.BinaryCommandLocal | proxyprotocol.BinaryVersion2
		protocol := proxyprotocol.BinaryProtocolUnspec
		data := append(proxyprotocol.BinarySignature, commandVersion, protocol, 0, 0)
		testParser(t, testParserArgs{
			headerParser: binaryHeaderParser,
			data:         data,
			header:       nil,
			err:          nil,
			readAll:      true,
		})
	})

	t.Run("Local command with TLV", func(t *testing.T) {
		commandVersion := proxyprotocol.BinaryCommandLocal | proxyprotocol.BinaryVersion2
		protocol := proxyprotocol.BinaryProtocolUnspec

		tlvData := []byte{proxyprotocol.TLVTypeNoop, 0, 0}

		addressLen := make([]byte, 2)
		binary.BigEndian.PutUint16(addressLen, uint16(len(tlvData)))

		t.Logf("Address len: %v", addressLen)

		data := append(proxyprotocol.BinarySignature, commandVersion, protocol)
		data = append(data, addressLen...)
		data = append(data, tlvData...)

		testParser(t, testParserArgs{
			headerParser: binaryHeaderParser,
			data:         data,
			header:       nil,
			err:          nil,
			readAll:      true,
		})
	})

	t.Run("Invalid protocol", func(t *testing.T) {
		commandVersion := proxyprotocol.BinaryCommandProxy | proxyprotocol.BinaryVersion2
		invalidProtocol := byte(0xFF)
		data := append(proxyprotocol.BinarySignature, commandVersion, invalidProtocol, 0, 0)
		testParser(t, testParserArgs{
			headerParser: binaryHeaderParser,
			data:         data,
			err:          proxyprotocol.ErrUnknownProtocol,
			readAll:      true,
		})
	})

	t.Run("Unspec protocol", func(t *testing.T) {
		commandVersion := proxyprotocol.BinaryCommandProxy | proxyprotocol.BinaryVersion2
		data := append(proxyprotocol.BinarySignature, commandVersion, proxyprotocol.BinaryProtocolUnspec, 0, 0)
		testParser(t, testParserArgs{
			headerParser: binaryHeaderParser,
			data:         data,
			header:       nil,
			err:          nil,
			readAll:      true,
		})
	})

	t.Run("TCPv4 protocol", func(t *testing.T) {
		commandVersion := proxyprotocol.BinaryCommandProxy | proxyprotocol.BinaryVersion2
		data := append(proxyprotocol.BinarySignature, commandVersion, proxyprotocol.BinaryProtocolTCPoverIPv4)

		t.Run("Invalid address size", func(t *testing.T) {
			data := append(data, 0, 0)
			testParser(t, testParserArgs{
				headerParser: binaryHeaderParser,
				data:         data,
				err:          proxyprotocol.ErrUnexpectedAddressLen,
				readAll:      true,
			})
		})

		t.Run("Valid address size", func(t *testing.T) {
			addressDataLen := 2 * (net.IPv4len + proxyprotocol.BinaryPortLen)
			data := append(data, 0, byte(addressDataLen))

			t.Run("Invalid address data size", func(t *testing.T) {
				testParser(t, testParserArgs{
					headerParser: binaryHeaderParser,
					data:         data,
					err:          io.EOF,
					readAll:      true,
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
					headerParser: binaryHeaderParser,
					data:         data,
					header:       &expectedHeader,
					readAll:      true,
				})
			})
		})
	})

	t.Run("TCPv6 protocol", func(t *testing.T) {
		commandVersion := proxyprotocol.BinaryCommandProxy | proxyprotocol.BinaryVersion2
		data := append(proxyprotocol.BinarySignature, commandVersion, proxyprotocol.BinaryProtocolTCPoverIPv6)

		t.Run("Invalid address size", func(t *testing.T) {
			data := append(data, 0, 0)
			testParser(t, testParserArgs{
				headerParser: binaryHeaderParser,
				data:         data,
				err:          proxyprotocol.ErrUnexpectedAddressLen,
				readAll:      true,
			})
		})

		t.Run("Valid address size", func(t *testing.T) {
			addressDataLen := 2 * (net.IPv6len + proxyprotocol.BinaryPortLen)
			data := append(data, 0, byte(addressDataLen))

			t.Run("Invalid address data size", func(t *testing.T) {
				testParser(t, testParserArgs{
					headerParser: binaryHeaderParser,
					data:         data,
					err:          io.EOF,
					readAll:      true,
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
					headerParser: binaryHeaderParser,
					data:         data,
					header:       &expectedHeader,
					readAll:      true,
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
					headerParser: binaryHeaderParser,
					data:         data,
					header:       &expectedHeader,
					readAll:      true,
				})
			})
		})
	})
}
