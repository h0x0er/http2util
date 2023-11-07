package parsehttp2frame

import (
	"net/http"

	"golang.org/x/net/http2"
)
// Frame2HTTPRequest
func Frame2HTTPRequest(f *http2.MetaHeadersFrame) (*http.Request, error) {
	return processMetaHeadersForRequest(f)
}

func Frame2HTTPResponse(f *http2.MetaHeadersFrame) (*http.Response, error) {
	return processMetaHeadersForResponse(f)
}
