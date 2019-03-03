package proxyprotocol_test

import (
	"bufio"
	"bytes"
	"errors"
	"net"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/c0va23/go-proxyprotocol"
)

func TestStubHeaderParser_Parse(t *testing.T) {
	headerParser := new(proxyprotocol.StubHeaderParser)

	bytesBuff := bytes.NewBufferString("")
	readBuf := bufio.NewReader(bytesBuff)
	header, err := headerParser.Parse(readBuf)

	if header != nil {
		t.Errorf("Unexpected header %v", header)
	}

	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
}

func TestFallbackHeaderParserBuilder_Build(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := NewMockLogger(mockCtrl)

	firstHeaderParserBuilder := NewMockHeaderParserBuilder(mockCtrl)
	secondHeaderParserBuilder := NewMockHeaderParserBuilder(mockCtrl)

	firstHeaderParser := NewMockHeaderParser(mockCtrl)
	secondHeaderParser := NewMockHeaderParser(mockCtrl)

	firstHeaderParserBuilder.EXPECT().Build(logger).Return(firstHeaderParser)
	secondHeaderParserBuilder.EXPECT().Build(logger).Return(secondHeaderParser)

	headerParserBuilders := proxyprotocol.NewFallbackHeaderParserBuilder(
		firstHeaderParserBuilder,
		secondHeaderParserBuilder,
	)

	headerParser := headerParserBuilders.Build(logger)

	expectedHeaderParser := proxyprotocol.FallbackHeaderParser{
		Logger: logger,
		HeaderParsers: []proxyprotocol.HeaderParser{
			firstHeaderParser,
			secondHeaderParser,
		},
	}

	if !reflect.DeepEqual(expectedHeaderParser, headerParser) {
		t.Errorf("Unexpected header parser %+v", headerParser)
	}

}

func TestNewFallbackHeaderParser(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := NewMockLogger(mockCtrl)

	headerParser1 := NewMockHeaderParser(mockCtrl)
	headerParser2 := NewMockHeaderParser(mockCtrl)

	fallbackHeaderParser := proxyprotocol.NewFallbackHeaderParser(
		logger,
		headerParser1,
		headerParser2,
	)

	expectedFallbackHeaderParser := proxyprotocol.FallbackHeaderParser{
		Logger: logger,
		HeaderParsers: []proxyprotocol.HeaderParser{
			headerParser1,
			headerParser2,
		},
	}

	if !reflect.DeepEqual(fallbackHeaderParser, expectedFallbackHeaderParser) {
		t.Errorf("Unexpected fallback header parser %v", fallbackHeaderParser)
	}
}

func TestFallbackHeaderParser_Parse(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := NewMockLogger(mockCtrl)
	logger.EXPECT().Printf(gomock.Any(), gomock.Any()).AnyTimes()

	firstHeaderParser := NewMockHeaderParser(mockCtrl)
	secondHeaderParser := NewMockHeaderParser(mockCtrl)

	fallbackHeaderParser := proxyprotocol.NewFallbackHeaderParser(
		logger,
		firstHeaderParser,
		secondHeaderParser,
	)

	reader := bytes.NewBufferString("valid header")
	readBuf := bufio.NewReader(reader)

	parseErr := errors.New("Parse error")

	expectedHeader := &proxyprotocol.Header{
		SrcAddr: &net.TCPAddr{
			IP: net.IPv4(1, 2, 3, 4),
		},
	}

	testHeaderParseResult := func(t *testing.T, header, expectedHeader *proxyprotocol.Header, err, expectedErr error) {
		if header != expectedHeader {
			t.Errorf("Unexpected header %v", header)
		}

		if err != expectedErr {
			t.Errorf("Unexpected error %s", err)
		}
	}

	t.Run("when first header parser return unknown error", func(t *testing.T) {
		firstHeaderParser.EXPECT().Parse(readBuf).Return(nil, parseErr)

		header, err := fallbackHeaderParser.Parse(readBuf)
		testHeaderParseResult(t, header, nil, err, parseErr)
	})

	t.Run("when first header parser return header", func(t *testing.T) {
		firstHeaderParser.EXPECT().Parse(readBuf).Return(expectedHeader, nil)

		header, err := fallbackHeaderParser.Parse(readBuf)
		testHeaderParseResult(t, header, expectedHeader, err, nil)
	})

	t.Run("when first header parser return ErrInvalidSignature error", func(t *testing.T) {
		firstHeaderParser.EXPECT().Parse(readBuf).Return(nil, proxyprotocol.ErrInvalidSignature).AnyTimes()

		t.Run("when second header parser return unknown error", func(t *testing.T) {
			secondHeaderParser.EXPECT().Parse(readBuf).Return(nil, parseErr)

			header, err := fallbackHeaderParser.Parse(readBuf)
			testHeaderParseResult(t, header, nil, err, parseErr)
		})

		t.Run("when second header parser return header", func(t *testing.T) {
			secondHeaderParser.EXPECT().Parse(readBuf).Return(expectedHeader, nil)

			header, err := fallbackHeaderParser.Parse(readBuf)
			testHeaderParseResult(t, header, expectedHeader, err, nil)
		})

		t.Run("when second header parser return ErrInvalidSignature error", func(t *testing.T) {
			secondHeaderParser.EXPECT().Parse(readBuf).Return(nil, proxyprotocol.ErrInvalidSignature)

			header, err := fallbackHeaderParser.Parse(readBuf)
			testHeaderParseResult(t, header, nil, err, proxyprotocol.ErrInvalidHeader)
		})
	})
}
