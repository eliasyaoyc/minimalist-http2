package minimalist_http2

import (
	"fmt"
	"github.com/Jxck/color"
	"github.com/Jxck/logger"
	"io"
	"log"
	"minimalist-http2/frame"
	"minimalist-http2/hpack"
	"time"
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
		fr, err := frame.ReadFrame(conn.RW, conn.Settings)
		if err != nil {
			logger.Error("connection.ReadLoop error,err: %v", err)
			h2Error, ok := err.(*H2Error)
			if ok {
				conn.GoAway(0, h2Error)
			}
			break
		}
		if fr != nil {
			logger.Notice("%v %v", color.Green("recv"), util.Indent(fr.String()))
		}

		streamID := fr.Header().StreamID
		types := fr.Header().Type

		if streamID == 0 {
			if types == frame.DataFrameType ||
				types == frame.HeadersFrameType ||
				types == frame.PriorityFrameType ||
				types == frame.RstStreamFrameType ||
				types == frame.PushPromiseFrameType ||
				types == frame.ContinuationFrameType {

				msg := fmt.Sprintf("%s Frame for stream ID 0", types)
				logger.Error("%v", msg)
				conn.GoAway(0, &H2Error{PROTOCOL_ERROR, msg})
				break
			}

			if types == frame.SettingsFrameType {
				settingsFrame, ok := fr.(*frame.SettingsFrame)
				if !ok {
					logger.Error("invalid settings frame %v", fr)
					return
				}
				conn.HandleSettings(settingsFrame)
			}

			if types == frame.WindowUpdateFrameType {
				windowUpdateFrame, ok := fr.(*frame.WindowUpdateFrame)
				if !ok {
					logger.Error("invalid window update frame %v", fr)
					return
				}
				logger.Debug("connection window size increment(%v)", int32(windowUpdateFrame.WindowSizeIncrement))
				conn.Window.UpdatePeer(int32(windowUpdateFrame.WindowSizeIncrement))
			}

			// respond to PING
			if types == frame.PingFrameType {
				if fr.Header().Flags != frame.PING_ACK {
					conn.PingACK([]byte("pong    ")) // 8 byte
				}
				continue
			}

			if types == frame.GoAwayFrameType {
				logger.Debug("stop conn.ReadLoop() by GOAWAY")
				break
			}
		}
		if streamID > 0 {
			if types == frame.SettingsFrameType ||
				types == frame.PingFrameType ||
				types == frame.GoAwayFrameType {
				msg := fmt.Sprintf("%s Frame for stream Id not 0", types)
				logger.Error("%v", msg)
				conn.GoAway(0, &H2Error{PROTOCOL_ERROR, msg})
				break
			}

			if types == frame.DataFrameType {
				length := int32(fr.Header().Length)
				conn.WindowConsume(length)
			}

			stream, ok := conn.Streams[streamID]
			if !ok {
				stream = conn.NewStream(streamID)
				conn.Streams[streamID] = stream

				if streamID > conn.LastStreamID {
					conn.LastStreamID = streamID
				}
			}

			err = stream.ChangeState(fr, RECV)
			if err != nil {
				logger.Error("%v", err)
				h2Error, ok := err.(*H2Error)
				if ok {
					conn.GoAway(0, h2Error)
				}
				break
			}

			if stream.State == CLOSED {
				go func(streamID uint32) {
					<-time.After(1 * time.Second)
					logger.Info("remove stream(%d) from conn.Streams[]", streamID)
					conn.Streams[streamID] = nil
				}(streamID)
			}
			stream.ReadChan <- fr
		}
	}
	logger.Debug("stop the readLoop")
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
