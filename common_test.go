package proxyprotocol_test

import (
	"bufio"
	"bytes"
	"reflect"
	"testing"

	proxyprotocol "github.com/c0va23/go-proxyprotocol"
)

func testParser(
	t *testing.T,
	headerParser proxyprotocol.HeaderParser,
	expectedHeader *proxyprotocol.Header,
	expextedError error,
	data []byte,
) {
	buf := bufio.NewReader(bytes.NewBuffer(data))
	header, err := headerParser(buf)

	if !reflect.DeepEqual(expectedHeader, header) {
		t.Errorf("Invalid header. Expected %+v, got %+v", expectedHeader, header)
	}
	if expextedError != err {
		t.Errorf("Invalid error. Expected %v, got %v", expextedError, err)
	}
}
