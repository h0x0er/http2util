package parsehttp2frame

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"

	"golang.org/x/net/http/httpguts"
	"golang.org/x/net/http2"
)

func processHeaders(f *http2.MetaHeadersFrame) (*http.Request, error) {
	return newWriterAndRequest(f)
}

func newWriterAndRequest(f *http2.MetaHeadersFrame) (*http.Request, error) {

	rp := requestParam{
		method:    f.PseudoValue("method"),
		scheme:    f.PseudoValue("scheme"),
		authority: f.PseudoValue("authority"),
		path:      f.PseudoValue("path"),
	}

	isConnect := rp.method == "CONNECT"
	if isConnect {
		if rp.path != "" || rp.scheme != "" || rp.authority == "" {
			return nil, fmt.Errorf("bad_connect")
		}
	} else if rp.method == "" || rp.path == "" || (rp.scheme != "https" && rp.scheme != "http") {
		return nil, fmt.Errorf("bad_path_method")
	}

	rp.header = make(http.Header)
	for _, hf := range f.RegularFields() {
		rp.header.Add(canonicalHeader(hf.Name), hf.Value)
	}
	if rp.authority == "" {
		rp.authority = rp.header.Get("Host")
	}

	req, err := newWriterAndRequestNoBody(rp)
	if err != nil {
		return nil, err
	}

	// TODO: extract request body

	// bodyOpen := !f.StreamEnded()
	// if bodyOpen {
	// 	if vv, ok := rp.header["Content-Length"]; ok {
	// 		if cl, err := strconv.ParseUint(vv[0], 10, 63); err == nil {
	// 			req.ContentLength = int64(cl)
	// 		} else {
	// 			req.ContentLength = 0
	// 		}
	// 	} else {
	// 		req.ContentLength = -1
	// 	}
	// 	req.Body.(*requestBody).pipe = &pipe{
	// 		b: &dataBuffer{expected: req.ContentLength},
	// 	}
	// }
	return req, nil
}

func newWriterAndRequestNoBody(rp requestParam) (*http.Request, error) {

	var tlsState *tls.ConnectionState // nil if not scheme https
	if rp.scheme == "https" {
		tlsState = &tls.ConnectionState{}
	}

	needsContinue := httpguts.HeaderValuesContainsToken(rp.header["Expect"], "100-continue")
	if needsContinue {
		rp.header.Del("Expect")
	}
	// Merge Cookie headers into one "; "-delimited value.
	if cookies := rp.header["Cookie"]; len(cookies) > 1 {
		rp.header.Set("Cookie", strings.Join(cookies, "; "))
	}

	// Setup Trailers
	var trailer http.Header
	for _, v := range rp.header["Trailer"] {
		for _, key := range strings.Split(v, ",") {
			key = http.CanonicalHeaderKey(textproto.TrimString(key))
			switch key {
			case "Transfer-Encoding", "Trailer", "Content-Length":
				// Bogus. (copy of http1 rules)
				// Ignore.
			default:
				if trailer == nil {
					trailer = make(http.Header)
				}
				trailer[key] = nil
			}
		}
	}
	delete(rp.header, "Trailer")

	var url_ *url.URL
	var requestURI string
	if rp.method == "CONNECT" {
		url_ = &url.URL{Host: rp.authority}
		requestURI = rp.authority // mimic HTTP/1 server behavior
	} else {
		var err error
		url_, err = url.ParseRequestURI(rp.path)
		if err != nil {
			return nil, fmt.Errorf("bad_path")
		}
		requestURI = rp.path
	}

	// TODO: extract request body
	
	// body := &requestBody{
	// 	conn:          sc,
	// 	stream:        st,
	// 	needsContinue: needsContinue,
	// }
	req := &http.Request{
		Method: rp.method,
		URL:    url_,
		// RemoteAddr: sc.remoteAddrStr,
		Header:     rp.header,
		RequestURI: requestURI,
		Proto:      "HTTP/2.0",
		ProtoMajor: 2,
		ProtoMinor: 0,
		TLS:        tlsState,
		Host:       rp.authority,
		// Body:       body,
		Trailer: trailer,
	}

	return req, nil
}

type requestParam struct {
	method                  string
	scheme, authority, path string
	header                  http.Header
}
