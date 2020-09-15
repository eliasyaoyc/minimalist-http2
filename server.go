package minimalist_http2

import (
	"crypto/tls"
	"github.com/Jxck/color"
	"github.com/Jxck/logger"
	"log"
	"net"
	"net/http"
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

}

func HandlerCallBack(handler http.Handler) CallBack {
	return nil
}
