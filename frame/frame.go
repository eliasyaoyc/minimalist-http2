package frame

import (
	"io"
	"minimalist-http2"
	"net/http"
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
	Header() *HeaderFrame
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
type HeaderFrame struct {
	Length            uint32 // 24 bit
	Type              FrameType
	Flags             Flag
	StreamID          uint32 // R + 31bit
	MaxFrameSize      int32
	MaxHeaderListSize int32
}

func NewFrameHeader(length uint32, types FrameType, flags Flag, streamID uint32) *HeaderFrame {
	return &HeaderFrame{
		Length:   length,
		Type:     types,
		Flags:    flags,
		StreamID: streamID,
	}
}

func (f *HeaderFrame) Write(w io.Writer) error {
	panic("implement me")
}

func (f *HeaderFrame) Read(r io.Writer) error {
	panic("implement me")
}

func (f *HeaderFrame) Header() *HeaderFrame {
	panic("implement me")
}

func (f *HeaderFrame) String() string {
	panic("implement me")
}

// Frame Data
//  +---------------+
// |Pad Length? (8)|
// +---------------+-----------------------------------------------+
// |                            Data (*)                         ...
// +---------------------------------------------------------------+
// |                           Padding (*)                       ...
// +---------------------------------------------------------------+

type DataFrame struct {
	*HeaderFrame
	PadLength uint8
	Data      []byte
	Padding   []byte
}

func NewDataFrame(flags Flag, streamID uint32, data, padding []byte) *DataFrame {
	var padded bool = flags&HEADERS_PADDED == HEADERS_PADDED

	length := len(data)

	if padded {
		length = length + len(padding) + 1
	} else {
		padding = nil
	}

	return &DataFrame{
		HeaderFrame: NewFrameHeader(uint32(length), DataFrameType, flags, streamID),
		PadLength:   uint8(len(padding)),
		Data:        data,
		Padding:     padding,
	}
}

func (f *DataFrame) Write(w io.Writer) error {
	panic("implement me")
}

func (f *DataFrame) Read(r io.Writer) error {
	panic("implement me")
}

func (f *DataFrame) Header() *HeaderFrame {
	panic("implement me")
}

func (f *DataFrame) String() string {
	panic("implement me")
}

// HEADERS
//
// +---------------+
// |Pad Length? (8)|
// +-+-------------+-----------------------------------------------+
// |E|                 Stream Dependency? (31)                     |
// +-+-------------+-----------------------------------------------+
// |  Weight? (8)  |
// +-+-------------+-----------------------------------------------+
// |                   Header Block Fragment (*)                 ...
// +---------------------------------------------------------------+
// |                           Padding (*)                       ...
// +---------------------------------------------------------------+
type HeadersFrame struct {
	*HeaderFrame
	PadLength           uint8
	DependencyTree      *DependencyTree
	HeaderBlockFragment []byte
	Headers             http.Header
	Padding             []byte
}

type DependencyTree struct {
	Exclusive        bool
	StreamDependency uint32
	Weight           uint8
}

func NewHeadersFrame(flags Flag, streamID uint32, dependenctTree *DependencyTree, headerBlockFragment, padding []byte) *HeadersFrame {
	var padded bool = flags&HEADERS_PADDED == HEADERS_PADDED
	var priority bool = flags&HEADERS_PRIORITY == HEADERS_PRIORITY

	length := len(headerBlockFragment)
	if padded {
		length = length + len(padding) + 1
	}
	if priority {
		length = length + 1
	}
	return &HeadersFrame{
		HeaderFrame:         NewFrameHeader(uint32(length), HeadersFrameType, flags, streamID),
		PadLength:           uint8(len(padding)),
		DependencyTree:      dependenctTree,
		HeaderBlockFragment: headerBlockFragment,
		Padding:             padding,
	}
}

func (f *HeadersFrame) Write(w io.Writer) error {
	panic("implement me")
}

func (f *HeadersFrame) Read(r io.Writer) error {
	panic("implement me")
}

func (f *HeadersFrame) Header() *HeaderFrame {
	return f.HeaderFrame
}

func (f *HeadersFrame) String() string {
	panic("implement me")
}

// PRIORITY
//
// +-+-------------------------------------------------------------+
// |E|                  Stream Dependency (31)                     |
// +-+-------------+-----------------------------------------------+
// |   Weight (8)  |
// +-+-------------+

type PriorityFrame struct {
	*HeaderFrame
	Exclusive        bool
	StreamDependency uint32
	Weight           uint8
}

func NewPriorityFrame(streamID uint32, exclusive bool, streamDependency uint32, weight uint8) *PriorityFrame {
	var length uint32 = 5

	return &PriorityFrame{
		HeaderFrame:      NewFrameHeader(length, PriorityFrameType, UNSET, streamID),
		Exclusive:        exclusive,
		StreamDependency: streamDependency,
		Weight:           weight,
	}
}

func (f *PriorityFrame) Write(w io.Writer) error {
	panic("implement me")
}

func (f *PriorityFrame) Read(r io.Writer) error {
	panic("implement me")
}

func (f *PriorityFrame) Header() *HeaderFrame {
	return f.HeaderFrame
}

func (f *PriorityFrame) String() string {
	panic("implement me")
}

// RST_STREAM
//
// +---------------------------------------------------------------+
// |                        Error Code (32)                        |
// +---------------------------------------------------------------+
type RstStreamFrame struct {
	*HeaderFrame
	ErrCode minimalist_http2.ErrCode
}

func NewRstStreamFrame(streamID uint32, errorCode minimalist_http2.ErrCode) *RstStreamFrame {
	var length uint32 = 4

	return &RstStreamFrame{
		HeaderFrame: NewFrameHeader(length, RstStreamFrameType, UNSET, streamID),
		ErrCode:     errorCode,
	}
}

func (f *RstStreamFrame) Write(w io.Writer) error {
	panic("implement me")
}

func (f *RstStreamFrame) Read(r io.Writer) error {
	panic("implement me")
}

func (f *RstStreamFrame) Header() *HeaderFrame {
	return f.HeaderFrame
}

func (f *RstStreamFrame) String() string {
	panic("implement me")
}

// SETTINGS FRAME     section 6.5.1
// +-------------------------------+
// |       Identifier (16)         |
// +-------------------------------+-------------------------------+
// |                        Value (32)                             |
// +---------------------------------------------------------------+
type SettingsID uint16

const (
	SETTINGS_HEADER_TABLE_SIZE      SettingsID = 0x1
	SETTINGS_ENABLE_PUSH                       = 0x2
	SETTINGS_MAX_CONCURRENT_STREAMS            = 0x3
	SETTINGS_INITIAL_WINDOW_SIZE               = 0x4
	SETTINGS_MAX_FRAME_SIZE                    = 0x5
	SETTINGS_MAX_HEADER_LIST_SIZE              = 0x6
)

const (
	DEFAULT_HEADER_TABLE_SIZE      int32 = 4096
	DEFAULT_ENABLE_PUSH                  = 1
	DEFAULT_MAX_CONCURRENT_STREAMS       = 2<<30 - 1
	DEFAULT_INITIAL_WINDOW_SIZE          = 65535
	DEFAULT_MAX_FRAME_SIZE               = 16384
	DEFAULT_MAX_HEADER_LIST_SIZE         = 2<<30 - 1
)

// PUSH_PROMISE  section 6.6
//  +---------------+
// |Pad Length? (8)|
// +-+-------------+-----------------------------------------------+
// |R|                  Promised Stream ID (31)                    |
// +-+-----------------------------+-------------------------------+
// |                   Header Block Fragment (*)                 ...
// +---------------------------------------------------------------+
// |                           Padding (*)                       ...
// +---------------------------------------------------------------+

type PushPromiseFrame struct {
	*HeaderFrame
	PadLength           uint8
	PromisedStreamId    uint32 // R + promisedStreamId
	HeaderBlockFragment []byte
	Padding             []byte
}

func NewPushPromiseFrame(flags Flag, streamId, promisedStreamId uint32, headerBlockFragment, padding []byte) *PushPromiseFrame {
	var padded bool = flags&HEADERS_PADDED == HEADERS_PADDED
	length := 4 + len(headerBlockFragment)

	if padded {
		length = length + len(padding) + 1
	}
	fh := NewFrameHeader(uint32(length), PushPromiseFrameType, flags, streamId)

	return &PushPromiseFrame{
		HeaderFrame:         fh,
		PadLength:           uint8(len(padding)),
		PromisedStreamId:    promisedStreamId,
		HeaderBlockFragment: headerBlockFragment,
		Padding:             padding,
	}
}

func (f *PushPromiseFrame) Write(w io.Writer) error {
	panic("implement me")
}

func (f *PushPromiseFrame) Read(r io.Writer) error {
	panic("implement me")
}

func (f *PushPromiseFrame) Header() *HeaderFrame {
	return f.HeaderFrame
}

func (f *PushPromiseFrame) String() string {
	panic("implement me")
}

// PING
//
// +---------------------------------------------------------------+
// |                                                               |
// |                      Opaque Data (64)                         |
// |                                                               |
// +---------------------------------------------------------------+
type PingFrame struct {
	*HeaderFrame
	OpaqueData []byte
}

func NewPingFrame(flags Flag, streamID uint32, opaqueData []byte) *PingFrame {
	var length uint32 = 8
	return &PingFrame{
		HeaderFrame: NewFrameHeader(length, PingFrameType, flags, streamID),
		OpaqueData:  opaqueData,
	}
}
func (f *PingFrame) Write(w io.Writer) error {
	panic("implement me")
}

func (f *PingFrame) Read(r io.Writer) error {
	panic("implement me")
}

func (f *PingFrame) Header() *HeaderFrame {
	return f.HeaderFrame
}

func (f *PingFrame) String() string {
	panic("implement me")
}

// GOAWAY
//
// +-+-------------------------------------------------------------+
// |R|                  Last-Stream-ID (31)                        |
// +-+-------------------------------------------------------------+
// |                      Error Code (32)                          |
// +---------------------------------------------------------------+
// |                  Additional Debug Data (*)                    |
// +---------------------------------------------------------------+
type GoAwayFrame struct {
	*HeaderFrame
	LastStreamID        uint32
	ErrorCode           minimalist_http2.ErrCode
	AdditionalDebugData []byte
}

func NewGoAwayFrame(streamID uint32, lastStreamID uint32, errorCode minimalist_http2.ErrCode, additionalDebugData []byte) *GoAwayFrame {
	var length = 8 + len(additionalDebugData)

	return &GoAwayFrame{
		HeaderFrame:         NewFrameHeader(uint32(length), GoAwayFrameType, UNSET, streamID),
		LastStreamID:        lastStreamID,
		ErrorCode:           errorCode,
		AdditionalDebugData: additionalDebugData,
	}
}

func (f *GoAwayFrame) Write(w io.Writer) error {
	panic("implement me")
}

func (f *GoAwayFrame) Read(r io.Writer) error {
	panic("implement me")
}

func (f *GoAwayFrame) Header() *HeaderFrame {
	return f.HeaderFrame
}

func (f *GoAwayFrame) String() string {
	panic("implement me")
}

// WINDOW_UPDATE
//
// +-+-------------------------------------------------------------+
// |R|              Window Size Increment (31)                     |
// +-+-------------------------------------------------------------+
type WindowUpdateFrame struct {
	*HeaderFrame
	WindowSizeIncrement uint32
}

func NewWindowUpdateFrame(streamID, incrementSize uint32) *WindowUpdateFrame {
	var length uint32 = 4

	return &WindowUpdateFrame{
		HeaderFrame:         NewFrameHeader(uint32(length), WindowUpdateFrameType, UNSET, streamID),
		WindowSizeIncrement: incrementSize,
	}
}

func (f *WindowUpdateFrame) Write(w io.Writer) error {
	panic("implement me")
}

func (f *WindowUpdateFrame) Read(r io.Writer) error {
	panic("implement me")
}

func (f *WindowUpdateFrame) Header() *HeaderFrame {
	return f.HeaderFrame
}

func (f *WindowUpdateFrame) String() string {
	panic("implement me")
}

// CONTINUATION
//
// +---------------------------------------------------------------+
// |                   Header Block Fragment (*)                 ...
// +---------------------------------------------------------------+
type ContinuationFrame struct {
	*HeaderFrame
	HeaderBlockFragment []byte
}

func NewContinuationFrame(flags Flag, streamID uint32, headerBlockFragment []byte) *ContinuationFrame {
	length := len(headerBlockFragment)
	return &ContinuationFrame{
		HeaderFrame:         NewFrameHeader(uint32(length), ContinuationFrameType, flags, streamID),
		HeaderBlockFragment: headerBlockFragment,
	}
}

func (f *ContinuationFrame) Write(w io.Writer) error {
	panic("implement me")
}

func (f *ContinuationFrame) Read(r io.Writer) error {
	panic("implement me")
}

func (f *ContinuationFrame) Header() *HeaderFrame {
	return f.HeaderFrame
}

func (f *ContinuationFrame) String() string {
	panic("implement me")
}
