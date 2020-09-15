package minimalist_http2

import (
	"github.com/Jxck/logger"
	"log"
	"minimalist-http2/frame"
	"minimalist-http2/hpack"
	"net/http"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

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

func NewStream(id uint32, writeChan chan frame.Frame, settings, peerSettings map[frame.SettingsID]int32, hpackContext *hpack.Context, callback CallBack) *Stream {
	stream := &Stream{
		ID:           id,
		State:        IDLE,
		Window:       NewWindow(settings[frame.SETTINGS_INITIAL_WINDOW_SIZE], peerSettings[frame.SETTINGS_INITIAL_WINDOW_SIZE]),
		ReadChan:     make(chan frame.Frame),
		WriteChan:    writeChan,
		Settings:     settings,
		PeerSettings: peerSettings,
		HPackContext: hpackContext,
		CallBack:     callback,
		Bucket:       NewBucket(),
		Closed:       false,
	}
	go stream.ReadLoop()
	return stream
}

type Bucket struct {
	Headers http.Header
	Body    *Body
}

func NewBucket() *Bucket {
	return &Bucket{
		Headers: make(http.Header),
		Body:    new(Body),
	}
}

type CallBack func(stream *Stream)

func (stream *Stream) Read(f frame.Frame) {
	logger.Debug("stream (%d) recv (%v)", stream.ID, f.Header().Type)

}

func (stream *Stream) ReadLoop() {
	logger.Debug("start stream (%d) ReadLoop()", stream.ID)
	for f := range stream.ReadChan {
		stream.Read(f)
	}
	logger.Debug("stop Stream (%d) ReadLoop()", stream.ID)
}

func (stream *Stream) Write(frame frame.Frame) {
	logger.Trace("stream.Write (%v)", frame)
	if stream.Closed {
		return
	}
	stream.ChangeState(frame, SEND)
	stream.WriteChan <- frame
}

func (stream *Stream) Close() {
	logger.Debug("stream(%d) Close()", stream.ID)
	stream.Closed = true
	logger.Info("close stream(%v).ReadChan", stream.ID)
	close(stream.ReadChan)
}
