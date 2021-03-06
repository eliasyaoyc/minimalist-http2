package hpack

type HeaderField struct {
	Name  string
	Value string
}

func NewHeaderField(name, value string) *HeaderField {
	return &HeaderField{name, value}
}

// The size of an entryi the sum of its name's length in octets
// of its value's length in octets and of 32 octets
func (h *HeaderField) Size() uint32 {
	return uint32(len(h.Name) + len(h.Value) + 32)
}
