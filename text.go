package proxyprotocol

// TextSignature is prefix for proxyprotocol v1
var (
	TextSignature    = []byte("PROXY")
	TextSignatureLen = len(TextSignature)
	TextSeparator    = " "
	TextCR           = byte('\r')
	TextLF           = byte('\n')
	TextCRLF         = []byte{TextCR, TextLF}
)

// TextProtocol list
var (
	TextProtocolIPv4    = "TCP4"
	TextProtocolIPv6    = "TCP6"
	TextProtocolUnknown = "UNKNOWN"
)
