package hpack

type DynamicTable struct {
	DYNAMIC_TABLE_SIZE uint32
	HeaderFields       []*HeaderField
}
