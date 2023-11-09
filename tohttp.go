package http2util

import (
	"net/http"

	"golang.org/x/net/http2"
)

// Frame2HTTPRequest
func FrameToHTTPRequest(f *http2.MetaHeadersFrame) (*http.Request, error) {
	return processMetaHeadersForRequest(f)
}

func FrameToHTTPResponse(f *http2.MetaHeadersFrame) (*http.Response, error) {
	return processMetaHeadersForResponse(f)
}
