package hpack

import "flag"

func init() {
	flag.Parse()
}

type Context struct {
	HT *DynamicTable
	ES *HeaderList
}

func NewContext(SETTING_HEADER_TABLE_SIZE uint32) *Context {
	return &Context{
		HT: NewDynamicTable(SETTING_HEADER_TABLE_SIZE),
		ES: nil,
	}
}
