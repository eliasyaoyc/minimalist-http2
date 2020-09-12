package minimalist_http2

import "fmt"

// An ErrCode is an unsigned 32-bit error code as defined in the HTTP/2 specification.
// You can see https://httpwg.org/specs/rfc7540.html#ErrorHandler
type ErrCode uint32

const (
	NO_ERROR                 ErrCode = 0x0
	PROTOCOL_ERROR           ErrCode = 0x1 // protocol error detected
	INTERNAL_ERROR           ErrCode = 0x2 // implementation fault
	FLOW_CONTROL_ERROR       ErrCode = 0x3 // Flow-control limits exceeded
	SETTINGS_TIMEOUT_ERROR   ErrCode = 0x4 // Setting not acknowledged
	STREAM_CLOSED_ERROR      ErrCode = 0x5 // Frame received for closed stream
	FRAME_SIZE_ERROR         ErrCode = 0x6 // Frame size incorrect
	REFUSED_STREAM_ERROR     ErrCode = 0x7 // Stream not processed
	CANCEL_ERROR             ErrCode = 0x8 // Stream cancelled
	COMPRESSION_ERROR        ErrCode = 0x9 // Compression state not updated
	CONNECT_ERROR            ErrCode = 0xa // Tcp connection error for CONNECT method
	ENHANCE_YOUR_CALM_ERROR  ErrCode = 0xb // Processing capacity exceeded
	INDEQUATE_SECURITY_ERROR ErrCode = 0xc // Negotiated TLS parameters not acceptable
	HTTP_1_1_REQUIRED_ERROR  ErrCode = 0xd // Use HTTP/1/1 fpr the request
)

func (e ErrCode) String() string {
	codes := []string{
		"NO_ERROR",
		"PROTOCOL_ERROR",
		"INTERNAL_ERROR",
		"FLOW_CONTROL_ERROR",
		"SETTINGS_TIMEOUT",
		"STREAM_CLOSED",
		"FRAME_SIZE_ERROR",
		"REFUSED_STREAM",
		"CANCEL",
		"COMPRESSION_ERROR",
		"CONNECT_ERROR",
		"ENHANCE_YOUR_CALM",
		"INADEQUATE_SECURITY",
		"HTTP_1_1_REQUIRED",
	}
	return codes[uint32(e)]
}

// ConnectionError is an error that results in the termination of the entire connection
type ConnectionError ErrCode

func (c ConnectionError) Error() string {
	return fmt.Sprintf("connection error: %s", ErrCode(c))
}

// StreamError is an error that only effects one stream within an HTTP/2 connection.
type StreamError struct {
	StreamID uint32
	Code     ErrCode
}

func (s StreamError) Error() string {
	return fmt.Sprintf("stream error: streamID %d; %v", s.StreamID, s.Code)
}

// Section 6.9.1 The Flw Control Window
// If a sender receives a WINDOW_UPDATE that causes a flow control
// window to exceed this maximum it MUST terminate either the stream
// or the connection, as appropriate. For streams,[...]; for the connection,
// a GOAWAY frame with a FLOW_CONTROL_ERROR code.
type goAwayFlowError struct {
}

func (g goAwayFlowError) Error() string {
	return "connection exceeded flow control windwo size"
}
