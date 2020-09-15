package minimalist_http2

import (
	"fmt"
	"github.com/Jxck/color"
	"github.com/Jxck/logger"
	"minimalist-http2/frame"
)

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

func (window *Window) UpdateInitialSize(newInitialWindowSize int32) {
	curInitialWindowSize := window.initialSize
	curWindowSize := window.peerCurrentSize
	newWindowSize := newInitialWindowSize - (window.initialSize - curWindowSize)

	window.peerCurrentSize = newWindowSize
	window.initialSize = newInitialWindowSize
	logger.Trace(color.Brown(`update initial window size
	"New WindowSize(%v)" = "New InitialWindowSize(%v)" - ("Current InitialWindow ize(%v)" - "Current WindowSize(%v)")`),
		newWindowSize, newInitialWindowSize, curInitialWindowSize, curWindowSize)
}

func (window *Window) Update(windowSizeIncrement int32) {
	cur := window.currentSize
	window.currentSize = cur + windowSizeIncrement
	logger.Trace(color.Brown("increment current window size (%v) + increment (%v) = (%v)"), cur, windowSizeIncrement, window.currentSize)
}

func (window *Window) UpdatePeer(windowSizeIncrement int32) {
	cur := window.peerCurrentSize
	window.peerCurrentSize = cur + windowSizeIncrement
	logger.Trace(color.Brown("increment peer window size (%v) + increment (%v) = (%v)"), cur, windowSizeIncrement, window.peerCurrentSize)

}

func (window *Window) Consume(length int32) (update int32) {
	window.currentSize -= length
	if window.currentSize < window.threshold {
		update = window.initialSize - window.currentSize
	}
	return update
}

func (window *Window) ConsumePeer(length int32) {
	current := window.peerCurrentSize
	window.peerCurrentSize = current - length
	logger.Trace("consume peer window size (%v) - (%v) = (%v)", current, length, window.peerCurrentSize)
}

func (window *Window) Consumable(length int32) int32 {
	if window.peerCurrentSize < length {
		return window.peerCurrentSize
	} else {
		return length
	}
}

func (window *Window) String() string {
	return fmt.Sprintf(color.Yellow("window: curr(%d) - peer(%d)"), window.currentSize, window.peerCurrentSize)
}
