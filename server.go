package minimalist_http2

import (
	"crypto/tls"
	"fmt"
	"github.com/Jxck/color"
	"github.com/Jxck/logger"
	"log"
	"minimalist-http2/frame"
	"minimalist-http2/hpack"
	"net"
	"net/http"
	neturl "net/url"
	"strconv"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

var TLSNextProto = map[string]func(server *http.Server, conn *tls.Conn, handler http.Handler){
	VERSION: TSLNextProtoHandler,
}

var TSLNextProtoHandler = func(server *http.Server, conn *tls.Conn, handler http.Handler) {
	logger.Notice(color.Yellow("New Connection from %s"), conn.RemoteAddr())
	HandleTLSConnection(conn, handler)
	return
}

func HandleTLSConnection(conn net.Conn, handler http.Handler) {
	logger.Info("Handle TLS Connection")

	Conn := NewConnection(conn)

	Conn.CallBack = HandlerCallBack(handler)

	err := Conn.ReadMagic()
	if err != nil {
		logger.Error("%v", err)
		return
	}

	go Conn.WriteLoop()

	settingsFrame := frame.NewSettingsFrame(frame.UNSET, 0, DefaultSettings)
	Conn.WriteChan <- settingsFrame

	Conn.ReadLoop()

	Conn.Close()

	logger.Info("return TLSNextProto will close connection")

	return
}

func HandlerCallBack(handler http.Handler) CallBack {
	return func(stream *Stream) {
		header := stream.Bucket.Headers
		body := stream.Bucket.Body

		authority := header.Get(":authority")
		method := header.Get(":method")
		path := header.Get(":path")
		scheme := header.Get(":scheme")

		header.Del(":authority")
		header.Del(":method")
		header.Del(":path")
		header.Del(":scheme")

		rawurl := fmt.Sprintf("%s://%s%s", scheme, authority, path)
		url, err := neturl.ParseRequestURI(rawurl)
		if err != nil {
			logger.Fatal("%v", err)
		}

		req := &http.Request{
			Method:           method,
			URL:              url,
			Proto:            "HTTP/1.1",
			ProtoMajor:       1,
			ProtoMinor:       1,
			Header:           header,
			Body:             body,
			ContentLength:    int64(body.Buffer.Len()),
			TransferEncoding: []string{},
			Close:            false,
			Host:             authority,
		}

		logger.Info("\n%s", color.Lime(util.RequestString(req)))

		// Handle HTTP using handler
		res := NewResponseWriter()
		handler.ServeHTTP(res, req)
		responseHeader := res.Header()
		responseHeader.Add(":status", strconv.Itoa(res.status))

		logger.Info("\n%s", color.Aqua(res.String()))

		// send response header ad HEADERS frame
		headerList := hpack.ToHeaderList(responseHeader)
		headerBlockFragment := stream.HPackContext.Encode(*headerList)
		logger.Debug("%v", headerList)

		headersFrame := frame.NewHeadersFrame(frame.HEADERS_END_HEADERS, stream.ID, nil, headerBlockFragment, nil)
		headersFrame.Headers = responseHeader

		stream.Write(headersFrame)

		// Send response body as DATA Frame
		// each DataFrame has data in window size
		data := res.body.Bytes()
		maxFrameSize := stream.PeerSettings[frame.SETTINGS_MAX_FRAME_SIZE]
		rest := int32(len(data))
		frameSize := rest

		for {
			logger.Debug("rest data size(%v), current peer(%v) window(%v)", rest, stream.ID, stream.Window)

			if rest == 0 {
				break
			}

			frameSize = stream.Window.Consumable(rest)

			if frameSize <= 0 {
				continue
			}

			if frameSize > maxFrameSize {
				frameSize = maxFrameSize
			}

			logger.Debug("send %v/%v data", frameSize, rest)

			dataToSend := make([]byte, frameSize)
			copy(dataToSend, data[:frameSize])
			dataFrame := frame.NewDataFrame(frame.UNSET, stream.ID, dataToSend, nil)
			stream.Write(dataFrame)

			rest -= frameSize
			copy(data, data[frameSize:])
			data = data[:rest]

			stream.Window.ConsumePeer(frameSize)
		}

		// End Stream in empty DATA Frame
		endDataFrame := frame.NewDataFrame(frame.DATA_END_STREAM, stream.ID, nil, nil)
		stream.Write(endDataFrame)
	}
}
