package minimalist_http2

import (
	"minimalist-http2/frame"
	"minimalist-http2/hpack"
	"net/http"
)

type Stream struct {
	ID           uint32
	State        StreamState
	Window       *Window
	ReadChan     chan frame.Frame
	WriteChan    chan frame.Frame
	Settings     map[frame.SettingsID]int32
	PeerSettings map[frame.SettingsID]int32
	HPackContext *hpack.Context
	CallBack     CallBack
	Bucket       *Bucket
	Closed       bool
}

func NewStream() *Stream {
	return &Stream{}
}

type Bucket struct {
	Headers http.Header
	Body    *Body
}
type CallBack func(stream *Stream)
