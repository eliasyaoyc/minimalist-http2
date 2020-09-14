package minimalist_http2

import "minimalist-http2/frame"

type Window struct {
	initialSize     int32
	currentSize     int32
	threshold       int32
	peerInitialSize int32
	peerCurrentSize int32
	peerThreshold   int32
}

func NewDefaultWindow() *Window {
	return &Window{
		initialSize:     frame.DEFAULT_INITIAL_WINDOW_SIZE,
		currentSize:     frame.DEFAULT_INITIAL_WINDOW_SIZE,
		threshold:       frame.DEFAULT_INITIAL_WINDOW_SIZE/2 + 1,
		peerInitialSize: frame.DEFAULT_INITIAL_WINDOW_SIZE,
		peerCurrentSize: frame.DEFAULT_INITIAL_WINDOW_SIZE,
		peerThreshold:   frame.DEFAULT_INITIAL_WINDOW_SIZE/2 + 1,
	}
}

func NewWindow(initialWindow, peerInitialWindow int32) *Window {
	return &Window{
		initialSize:     initialWindow,
		currentSize:     initialWindow,
		threshold:       initialWindow/2 + 1,
		peerInitialSize: peerInitialWindow,
		peerCurrentSize: peerInitialWindow,
		peerThreshold:   peerInitialWindow/2 + 1,
	}
}
