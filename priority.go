package minimalist_http2

// Priority ensures that limited resources can be directed to the most important stream first.

type Priority struct {
}

func NewPrioritization() *Priority {
	return &Priority{}
}
