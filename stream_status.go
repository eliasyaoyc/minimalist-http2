package minimalist_http2

import (
	"fmt"
	"github.com/Jxck/color"
	"github.com/Jxck/logger"
	"log"
	xframe "minimalist-http2/frame"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

// section 5.1. Stream States
type StreamState int

const (
	IDLE StreamState = iota
	RESERVED_LOCAL
	RESERVED_REMOTE
	OPEN
	HALF_CLOSED_LOCAL
	HALF_CLOSED_REMOTE
	CLOSED
)

func (s StreamState) String() string {
	states := []string{
		"IDLE",
		"RESERVED_LOCAL",
		"RESERVED_REMOTE",
		"OPEN",
		"HALF_CLOSED_LOCAL",
		"HALF_CLOSED_REMOTE",
		"CLOSED",
	}
	return states[int(s)]
}

type Context int

const (
	RECV Context = iota
	SEND
)

func (c Context) String() string {
	contexts := []string{
		"RECV",
		"SEND",
	}
	return contexts[int(c)]
}

//  Stream States
//                           +--------+
//                   send PP |        | recv PP
//                  ,--------|  idle  |--------.
//                 /         |        |         \
//                v          +--------+          v
//         +----------+          |           +----------+
//         |          |          | send H /  |          |
//  ,------| reserved |          | recv H    | reserved |------.
//  |      | (local)  |          |           | (remote) |      |
//  |      +----------+          v           +----------+      |
//  |          |             +--------+             |          |
//  |          |     recv ES |        | send ES     |          |
//  |   send H |     ,-------|  open  |-------.     | recv H   |
//  |          |    /        |        |        \    |          |
//  |          v   v         +--------+         v   v          |
//  |      +----------+          |           +----------+      |
//  |      |   half   |          |           |   half   |      |
//  |      |  closed  |          | send R /  |  closed  |      |
//  |      | (remote) |          | recv R    | (local)  |      |
//  |      +----------+          |           +----------+      |
//  |           |                |                 |           |
//  |           | send ES /      |       recv ES / |           |
//  |           | send R /       v        send R / |           |
//  |           | recv R     +--------+   recv R   |           |
//  | send R /  `----------->|        |<-----------'  send R / |
//  | recv R                 | closed |               recv R   |
//  `----------------------->|        |<----------------------'
//                           +--------+
//
//     send:   endpoint sends this frame
//     recv:   endpoint receives this frame
//
//     H:  HEADERS frame (with implied CONTINUATIONs)
//     PP: PUSH_PROMISE frame (with implied CONTINUATIONs)
//     ES: END_STREAM flag
//     R:  RST_STREAM frame
func (stream *Stream) ChangeState(frame xframe.Frame, context Context) (err error) {
	header := frame.Header()
	frameType := header.Type
	flags := header.Flags
	state := stream.State

	logger.Trace("change state(%v) with %v frame type(%v)", state, context, frameType)

	if frameType == xframe.SettingsFrameType ||
		frameType == xframe.GoAwayFrameType {
		return nil
	}

	switch stream.State {
	case IDLE:
		// H (Headers frame)
		if frameType == xframe.HeadersFrameType {
			stream.changeState(OPEN)

			// END_STREAM flag
			if flags&xframe.HEADERS_END_STREAM == xframe.HEADERS_END_STREAM {
				if context == RECV {
					stream.changeState(HALF_CLOSED_REMOTE)
				} else {
					stream.changeState(HALF_CLOSED_LOCAL)
				}
			}
			return
		}

		// PUSH_PROMISE frame
		if frameType == xframe.PushPromiseFrameType {
			if context == RECV {
				stream.changeState(RESERVED_REMOTE)
			} else {
				stream.changeState(RESERVED_LOCAL)
			}
			return
		}

		// Priority
		if frameType == xframe.PriorityFrameType {
			return
		}
	case RESERVED_LOCAL:
		// H (Headers frame)
		if frameType == xframe.HeadersFrameType && context == SEND {
			stream.changeState(HALF_CLOSED_REMOTE)
			return
		}
		// R (RST stream frame)
		if frameType == xframe.RstStreamFrameType {
			stream.changeState(CLOSED)
			return
		}
	case RESERVED_REMOTE:
		// H (Headers frame)
		if frameType == xframe.HeadersFrameType && context == RECV {
			stream.changeState(HALF_CLOSED_LOCAL)
			return
		}

		// R (RST stream frame)
		if frameType == xframe.RstStreamFrameType {
			stream.changeState(CLOSED)
			return
		}
	case OPEN:
		// ES (END_STREAM)
		if frameType == xframe.HEADERS_END_STREAM {
			if context == SEND {
				stream.changeState(HALF_CLOSED_LOCAL)
			} else {
				stream.changeState(HALF_CLOSED_REMOTE)
			}
			return
		}

		// R (RST stream frame)
		if frameType == xframe.RstStreamFrameType {
			stream.changeState(CLOSED)
			return
		}
		// every type of frame accepted
		return
	case HALF_CLOSED_LOCAL:
		if context == SEND {
			if frameType == xframe.WindowUpdateFrameType || frameType == xframe.PriorityFrameType {
				return
			}

			// R (RST  stream frame)
			if frameType == xframe.RstStreamFrameType {
				stream.changeState(CLOSED)
				return
			}
		}

		if context == RECV {
			// ES (END_STREAM)
			if frameType == xframe.HEADERS_END_STREAM {
				stream.changeState(CLOSED)
				return
			}

			// R (RST stream frame)
			if frameType == xframe.RstStreamFrameType {
				stream.changeState(CLOSED)
				return
			}

			// recv any type of frames are valid
			return
		}
	case HALF_CLOSED_REMOTE:
		if context == RECV {
			if frameType == xframe.WindowUpdateFrameType || frameType == xframe.PriorityFrameType {
				return
			}

			// R (RST stream frame)
			if frameType == xframe.RstStreamFrameType {
				stream.changeState(CLOSED)
				return
			}

			msg := fmt.Sprintf("invalid frame type %v at %v state", frameType, state)
			logger.Error(color.Red(msg))
			return &xframe.H2Error{
				ErrorCode:           xframe.STREAM_CLOSED,
				AdditionalDebugData: msg,
			}
		}

		if context == SEND {
			// ES (end stream)
			if flags&xframe.HEADERS_END_STREAM == xframe.HEADERS_END_STREAM {
				stream.changeState(CLOSED)
				return
			}

			// R (RST stream)
			if frameType == xframe.RstStreamFrameType {
				stream.changeState(CLOSED)
				return
			}
			// send frames of any types
			return
		}
	case CLOSED:
		if context == SEND {
			if frameType == xframe.PriorityFrameType {
				return
			}
		}
		if context == RECV {
			if frameType == xframe.WindowUpdateFrameType ||
				frameType == xframe.PriorityFrameType ||
				frameType == xframe.RstStreamFrameType {

				// valid frame
				return
			}
			msg := fmt.Sprintf("invalid frame type %v at %v state", frameType, state)
			logger.Error(color.Red(msg))
			return &xframe.H2Error{
				ErrorCode:           xframe.STREAM_CLOSED,
				AdditionalDebugData: msg,
			}
		}

	}

	msg := fmt.Sprintf("invalid frame type %v at %v state", frameType, state)
	logger.Error(color.Red(msg))
	return &xframe.H2Error{
		ErrorCode:           xframe.PROTOCOL_ERROR,
		AdditionalDebugData: msg,
	}
}

func (stream *Stream) changeState(state StreamState) {
	logger.Info("change stream (%d) state (%s -> %s)", stream.ID, stream.State, color.Pink(state.String()))
	stream.State = state
}
