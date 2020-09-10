package minimalist_http2

// Flow control and prioritization ensure that it is possible to efficiently use multiplexed streams.
// Flow control  helps to ensure that only data that can be used by a receiver is transmitted.

type FlowControl struct {
}

func NewFlowControl() *FlowControl {
	return &FlowControl{}
}
