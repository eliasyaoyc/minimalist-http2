package hpack

type DynamicTable struct {
	DYNAMIC_TABLE_SIZE uint32
	HeaderFields       []*HeaderField
}

func NewDynamicTable(SETTINGS_HEADER_TABLE_SIZE uint32) *DynamicTable {
	return &DynamicTable{
		DYNAMIC_TABLE_SIZE: SETTINGS_HEADER_TABLE_SIZE,
		HeaderFields:       make([]*HeaderField, 0),
	}
}
