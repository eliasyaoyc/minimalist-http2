package hpack

import "flag"

func init() {
	flag.Parse()
}

type Context struct {
	HT *DynamicTable
	ES *HeaderList
}
