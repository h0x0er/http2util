package http2util

import (
	"fmt"
	"net/http"

	"golang.org/x/net/http2"
)

// Frame2HTTPRequest creates http.Request from frame
func FrameToHTTPRequest(frame http2.Frame) (*http.Request, error) {
	f, ok := frame.(*http2.MetaHeadersFrame)
	if !ok {
		return nil, fmt.Errorf("error: only http2.MetaHeadersFrame is supported")
	}
	return processMetaHeadersForRequest(f)
}

// FrameToHTTPResponse creates http.Response from frame
func FrameToHTTPResponse(frame http2.Frame) (*http.Response, error) {
	f, ok := frame.(*http2.MetaHeadersFrame)
	if !ok {
		return nil, fmt.Errorf("error: only http2.MetaHeadersFrame is supported")
	}
	return processMetaHeadersForResponse(f)
}
