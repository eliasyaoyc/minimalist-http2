package hpack

type HeaderList []*HeaderField

func NewHeaderList() *HeaderList {
	return new(HeaderList)
}
