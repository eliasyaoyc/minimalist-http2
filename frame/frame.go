package frame

import (
	"fmt"
	"io"
)

const frameHeaderLen = 9

var padZeores = make([]byte, 255) // zero for padding

type FrameType uint8

const (
	DataFrameType         FrameType = 0x0
	HeadersFrameType                = 0x1
	PriorityFrameType               = 0x2
	RstStreamFrameType              = 0x3
	SettingsFrameType               = 0x4
	PushPromiseFrameType            = 0x5
	PingFrameType                   = 0x6
	GoAwayFrameType                 = 0x7
	WindowUpdateFrameType           = 0x8
	ContinuationFrameType           = 0x9
	AltsvcFrameType                 = 0xa
	OriginFrameType                 = 0xc
)

// overwrite
func (frameType FrameType) String() string {
	types := []string{
		"DATA",
		"HEADERS",
		"PRIORITY",
		"RES_STREAM",
		"SETTINGS",
		"PUSH_PROMISE",
		"PING",
		"GOAWAY",
		"WINDOW_UPDATE",
		"CONTINUATION",
		"ALTSVC",
		"ORIGIN",
	}
	return types[int(frameType)]
}

type ErrorCode uint32

const (
	NO_ERROR            ErrorCode = 0x0
	PROTOCOL_ERROR      ErrorCode = 0x1
	INTERNAL_ERROR      ErrorCode = 0x2
	FLOW_CONTROL_ERROR  ErrorCode = 0x3
	SETTINGS_TIMEOUT    ErrorCode = 0x4
	STREAM_CLOSED       ErrorCode = 0x5
	FRAME_SIZE_ERROR    ErrorCode = 0x6
	REFUSED_STREAM      ErrorCode = 0x7
	CANCEL              ErrorCode = 0x8
	COMPRESSION_ERROR   ErrorCode = 0x9
	CONNECT_ERROR       ErrorCode = 0xa
	ENHANCE_YOUR_CALM   ErrorCode = 0xb
	INADEQUATE_SECURITY ErrorCode = 0xc
	HTTP_1_1_REQUIRED   ErrorCode = 0xd
)

func (e ErrorCode) String() string {
	errors := []string{
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
	return errors[int(e)]
}

type H2Error struct {
	ErrorCode           ErrorCode
	AdditionalDebugData string
}

func (e H2Error) String() string {
	return fmt.Sprintf("%v(%v)", e.ErrorCode, e.AdditionalDebugData)
}

func (e H2Error) Error() string {
	return e.ErrorCode.String()
}

// Flags is a bitmask of HTTP/2 flags.
// The meaning of flags varies depending on the frame type.
type Flag uint8

// Has reports whether f contains all (0 or more) flag in v.
func (f Flag) Has(v Flag) bool {
	return (f & v) == v
}

// Frame-specific FrameHeader flag bits.
const (
	UNSET Flag = 0x0
	// Data
	DATA_END_STREAM = 0x1
	DATA_PADDED     = 0x8
	// header
	HEADERS_END_STREAM  = 0x1
	HEADERS_END_HEADERS = 0x4
	HEADERS_PADDED      = 0x8
	HEADERS_PRIORITY    = 0x20

	PING_ACK = 0x1

	CONTINUAION_END_HEADERS  = 0x4
	PUSH_PROMISE_END_HEADERS = 0x4
	PUSH_PROMISE_PADDED      = 0x8
)

func (f Flag) String() string {
	flags := []string{
		"UNSER",
		"DATA_END_STREAM",
		"DATA_PADDED",
		"HEADERS_END_STREAM",
		"HEADERS_END_HEADERS",
		"HEADERS_PADDED",
		"HEADERS_PRIORITY",
		"PING_ACK",
		"CONTINUATION_END_HEADERS",
		"PUSH_PROMISE_END_HEADERS",
		"PUSH_PROMISE_PADDED",
	}
	return flags[int(f)]
}

type Frame interface {
	Write(w io.Writer) error
	Read(r io.Writer) error
	Header() *FrameHeader
	String() string
}

// Frame Header
//
// +-----------------------------------------------+
// |                 Length (24)                   |
// +---------------+---------------+---------------+
// |   Type (8)    |   Flags (8)   |
// +-+-------------+---------------+-------------------------------+
// |R|                 Stream Identifier (31)                      |
// +=+=============================================================+
// |                   Frame Payload (0...)                      ...
// +---------------------------------------------------------------+
type FrameHeader struct {
	Length            uint32 // 24 bit
	Type              FrameType
	Flags             Flag
	StreamID          uint32 // R + 31bit
	MaxFrameSize      int32
	MaxHeaderListSize int32
}