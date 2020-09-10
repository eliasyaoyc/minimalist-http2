package minimalist_http2

// Prioritization ensures that limited resources can be directed to the most important stream first.

type Prioritization struct {
}

func NewPrioritization() *Prioritization {
	return &Prioritization{}
}
