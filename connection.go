package minimalist_http2

import (
	"fmt"
	"github.com/Jxck/color"
	"github.com/Jxck/logger"
	"io"
	"log"
	"minimalist-http2/frame"
	"minimalist-http2/hpack"
)

// A transport-layer connection between tow endpoints
func init() {
	log.SetFlags(log.Lshortfile)
}

type Connection struct {
	RW           io.ReadWriter
	HPackContext *hpack.Context
	LastStreamID uint32
	Window       *Window
	Settings     map[frame.SettingsID]int32
	PeerSettings map[frame.SettingsID]int32
	Streams      map[uint32]*Stream
	WriteChan    chan frame.Frame
	CallBack     func(stream *Stream)
}

func NewConnection(rw io.ReadWriter) *Connection {
	return &Connection{
		RW:           rw,
		HPackContext: hpack.NewContext(uint32(frame.DEFAULT_HEADER_TABLE_SIZE)),
		Window:       NewDefaultWindow(),
		Settings:     DefaultSettings,
		PeerSettings: DefaultSettings,
		Streams:      make(map[uint32]*Stream),
		WriteChan:    make(chan frame.Frame),
	}
}

func (conn *Connection) NewStream(streamID uint32) *Stream {
	logger.Debug("adding new stream (id=%d) total (%d)", streamID, len(conn.Streams))

	return NewStream(
		streamID,
		conn.WriteChan,
		conn.Settings,
		conn.PeerSettings,
		conn.HPackContext,
		conn.CallBack)
}

func (conn *Connection) HandleSettings(settingsFrame *frame.SettingsFrame) {
	if settingsFrame.Flags == frame.PING_ACK {
		logger.Trace("receive SETTINGS ack")
		return
	}

	if settingsFrame.Flags != frame.UNSET {
		logger.Error("unknown flag of SETTINGS Frame %v", settingsFrame.Flags)
		return
	}

	// received SETTINGS frame
	settings := settingsFrame.Settings

	defaultSettings := map[frame.SettingsID]int32{
		frame.SETTINGS_HEADER_TABLE_SIZE:      frame.DEFAULT_HEADER_TABLE_SIZE,
		frame.SETTINGS_ENABLE_PUSH:            frame.DEFAULT_ENABLE_PUSH,
		frame.SETTINGS_MAX_CONCURRENT_STREAMS: frame.DEFAULT_MAX_CONCURRENT_STREAMS,
		frame.SETTINGS_INITIAL_WINDOW_SIZE:    frame.DEFAULT_INITIAL_WINDOW_SIZE,
		frame.SETTINGS_MAX_FRAME_SIZE:         frame.DEFAULT_MAX_FRAME_SIZE,
		frame.SETTINGS_MAX_HEADER_LIST_SIZE:   frame.DEFAULT_MAX_HEADER_LIST_SIZE,
	}

	for k, v := range settings {
		defaultSettings[k] = v
	}

	logger.Trace("merged settings==================")

	for k, v := range defaultSettings {
		logger.Trace("%v:%v", k, v)
	}

	// save settings to conn
	conn.Settings = settings

	initialWindowSize, ok := settings[frame.SETTINGS_INITIAL_WINDOW_SIZE]
	if ok {
		if initialWindowSize > 2147483647 { // validate < 2^31-1
			logger.Error("FLOW_CONTROL_ERROR (%s)", "SETTINGS_INITIAL_WINDOW_SIZE too large")
			return
		}
		conn.PeerSettings[frame.SETTINGS_INITIAL_WINDOW_SIZE] = initialWindowSize

		for _, stream := range conn.Streams {
			log.Println("apply settings to stream", stream)
			stream.Window.UpdateInitialSize(initialWindowSize)
			stream.PeerSettings[frame.SETTINGS_INITIAL_WINDOW_SIZE] = initialWindowSize
		}
	}

	// send ack
	ack := frame.NewSettingsFrame(frame.PING_ACK, 0, NilSettings)
	conn.WriteChan <- ack
}

func (conn *Connection) ReadLoop() {
	logger.Debug("stop the readLoop")
	for {
		frame, err := frame.ReadFrame(conn.RW, conn.Settings)
		if err != nil {
			logger.Error("connection.ReadLoop error,err: %v", err)
			h2Error, ok := err.(*H2Error)
			if ok {
				conn.GoAway(0, h2Error)
			}
			break
		}
		if frame != nil {
			logger.Notice("%v %v", color.Green("recv"), util.Indent(frame.String()))
		}

		streamID := frame.Header().StreamID
		types := frame.Header().Type

		if streamID == 0 {

		}
	}
}

func (conn *Connection) WriteLoop() error {
	logger.Debug("start connection.WriteLoop")
	for frame := range conn.WriteChan {
		logger.Notice("%v %v", color.Red("send"), util.Indent(frame.String()))

		err := frame.Write(conn.RW)
		if err != nil {
			logger.Error("connection frame.Write error, err: %v", err)
			return err
		}
	}
	return nil
}

func (conn *Connection) PingACK(opaqueData []byte) {
	logger.Debug("Ping ACK with opaque(%v)", opaqueData)
	pingACK := frame.NewPingFrame(frame.PING_ACK, 0, opaqueData)
	conn.WriteChan <- pingACK
}

func (conn *Connection) GoAway(streamID uint32, h2Error *H2Error) {
	logger.Debug("connection close with GO_AWAY(%v)", h2Error)
	errorCode := h2Error.ErrCode
	additionalDebugData := []byte(h2Error.AdditionalDebugData)
	goaway := frame.NewGoAwayFrame(streamID, conn.LastStreamID, errorCode, additionalDebugData)
	conn.WriteChan <- goaway
}

func (conn *Connection) WindowConsume(length int32) {
	logger.Debug("connection window update %d byte", length)

	update := conn.Window.Consume(length)

	if update > 0 {
		conn.WriteChan <- frame.NewWindowUpdateFrame(0, uint32(update))
		conn.Window.Update(update)
	}
}

func (conn *Connection) WriteMagic() error {
	_, err := conn.RW.Write([]byte(CONNECTION_PREFACE))
	if err != nil {
		return err
	}
	logger.Info("%v %q", color.Red("send"), CONNECTION_PREFACE)
	return nil
}
func (conn *Connection) ReadMagic() error {
	magic := make([]byte, len(CONNECTION_PREFACE))
	_, err := conn.RW.Read(magic)
	if err != nil {
		return err
	}
	if string(magic) != CONNECTION_PREFACE {
		logger.Info("Invalid Magic String: %q", string(magic))
		return fmt.Errorf("Invalid Magic String")
	}
	return nil
}

func (conn *Connection) Close() {
	logger.Info("close all connection.frame")
	for i, stream := range conn.Streams {
		if stream != nil {
			logger.Debug("close stream(%d)", i)
			stream.Close()
		}
	}
	close(conn.WriteChan)
}
