package proxyprotocol_test

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"reflect"

	"github.com/c0va23/go-proxyprotocol"

	"testing"
)

func testResult(
	t *testing.T,
	expectedHeader *proxyprotocol.Header,
	expextedError error,
	data []byte,
) {
	buf := bufio.NewReader(bytes.NewBuffer(data))
	header, err := proxyprotocol.ParseV2Header(buf)

	if !reflect.DeepEqual(expectedHeader, header) {
		t.Errorf("Invalid header. Expected %v, got %v", expectedHeader, header)
	}
	if expextedError != err {
		t.Errorf("Invalid error: %v", err)
	}
}

func TestParseV2Header(t *testing.T) {
	t.Run("buffer EOF", func(t *testing.T) {
		testResult(t, nil, io.EOF, []byte{})
	})

	t.Run("Invalid signature", func(t *testing.T) {
		data := make([]byte, proxyprotocol.BinarySignatueLen)
		testResult(t, nil, proxyprotocol.ErrInvalidSignature, data)
	})

	t.Run("Invalid version", func(t *testing.T) {
		invalidVersoin := byte(0x00)
		data := append(proxyprotocol.BinarySignatue, invalidVersoin)
		testResult(t, nil, proxyprotocol.ErrUnknownVersion, data)
	})

	t.Run("Invalid command", func(t *testing.T) {
		invalidCommand := proxyprotocol.BinaryVersion2&proxyprotocol.BinaryVersionMask | proxyprotocol.BinaryCommandMask&0xFF
		t.Logf("Version command bits: %02x", invalidCommand)
		data := append(proxyprotocol.BinarySignatue, invalidCommand)
		testResult(t, nil, proxyprotocol.ErrUnknownCommand, data)
	})

	t.Run("Invalid protocol", func(t *testing.T) {
		commandVerison := proxyprotocol.BinaryCommandProxy | proxyprotocol.BinaryVersion2
		invalidProtocol := byte(0xFF)
		data := append(proxyprotocol.BinarySignatue, commandVerison, invalidProtocol, 0, 0)
		testResult(t, nil, proxyprotocol.ErrUnknownProtocol, data)
	})

	t.Run("Unspec protocol", func(t *testing.T) {
		commandVerison := proxyprotocol.BinaryCommandProxy | proxyprotocol.BinaryVersion2
		data := append(proxyprotocol.BinarySignatue, commandVerison, proxyprotocol.BinaryProtocolUnspec, 0, 0)
		testResult(t, nil, nil, data)
	})

	t.Run("IPv4 protocol", func(t *testing.T) {
		commandVerison := proxyprotocol.BinaryCommandProxy | proxyprotocol.BinaryVersion2
		data := append(proxyprotocol.BinarySignatue, commandVerison, proxyprotocol.BinaryProtocolTCPoverIPv4)

		t.Run("Invalid address size", func(t *testing.T) {
			data := append(data, 0, 0)
			testResult(t, nil, proxyprotocol.ErrUnexpectedAddressLen, data)
		})

		t.Run("Valid address size", func(t *testing.T) {
			data := append(data, 0, 12)

			t.Run("Invalid address data size", func(t *testing.T) {
				testResult(t, nil, io.EOF, data)
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

				expectedHader := proxyprotocol.Header{
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
				testResult(t, &expectedHader, nil, data)
			})
		})
	})
}
