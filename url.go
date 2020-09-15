package minimalist_http2

import (
	neturl "net/url"
)

type URL struct {
	*neturl.URL
	Port string
}
