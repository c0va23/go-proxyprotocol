package proxyprotocol_test

import (
	"bufio"
	"bytes"
	"reflect"
	"testing"

	proxyprotocol "github.com/c0va23/go-proxyprotocol"
)

type testParserArgs struct {
	headerParser proxyprotocol.HeaderParser
	data         []byte
	header       *proxyprotocol.Header
	err          error
}

func testParser(
	t *testing.T,
	args testParserArgs,
) {
	buf := bufio.NewReader(bytes.NewBuffer(args.data))
	header, err := args.headerParser(buf)

	if !reflect.DeepEqual(args.header, header) {
		t.Errorf("Invalid header. Expected %+v, got %+v", args.header, header)
	}
	if args.err != err {
		t.Errorf("Invalid error. Expected %v, got %v", args.err, err)
	}
}
