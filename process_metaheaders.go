package http2util

// Note: Code is taken from https://cs.opensource.google/go/x/net/+/master:http2

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/net/http/httpguts"
	"golang.org/x/net/http2"
)

func processMetaHeadersForRequest(f *http2.MetaHeadersFrame) (*http.Request, error) {
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

func processMetaHeadersForResponse(f *http2.MetaHeadersFrame) (*http.Response, error) {
	if f.Truncated {

		return nil, fmt.Errorf("headers are truncated")
	}

	status := f.PseudoValue("status")
	if status == "" {
		return nil, errors.New("malformed response from server: missing status pseudo header")
	}
	statusCode, err := strconv.Atoi(status)
	if err != nil {
		return nil, errors.New("malformed response from server: malformed non-numeric status pseudo header")
	}

	regularFields := f.RegularFields()
	strs := make([]string, len(regularFields))
	header := make(http.Header, len(regularFields))
	res := &http.Response{
		Proto:      "HTTP/2.0",
		ProtoMajor: 2,
		Header:     header,
		StatusCode: statusCode,
		Status:     status + " " + http.StatusText(statusCode),
	}
	for _, hf := range regularFields {
		key := canonicalHeader(hf.Name)
		if key == "Trailer" {
			t := res.Trailer
			if t == nil {
				t = make(http.Header)
				res.Trailer = t
			}
			foreachHeaderElement(hf.Value, func(v string) {
				t[canonicalHeader(v)] = nil
			})
		} else {
			vv := header[key]
			if vv == nil && len(strs) > 0 {
				// More than likely this will be a single-element key.
				// Most headers aren't multi-valued.
				// Set the capacity on strs[0] to 1, so any future append
				// won't extend the slice into the other strings.
				vv, strs = strs[:1:1], strs[1:]
				vv[0] = hf.Value
				header[key] = vv
			} else {
				header[key] = append(vv, hf.Value)
			}
		}
	}

	// if statusCode >= 100 && statusCode <= 199 {
	// 	if f.StreamEnded() {
	// 		return nil, errors.New("1xx informational response with END_STREAM flag")
	// 	}
	// 	cs.num1xx++
	// 	const max1xxResponses = 5 // arbitrary bound on number of informational responses, same as net/http
	// 	if cs.num1xx > max1xxResponses {
	// 		return nil, errors.New("http2: too many 1xx informational responses")
	// 	}
	// 	if fn := cs.get1xxTraceFunc(); fn != nil {
	// 		if err := fn(statusCode, textproto.MIMEHeader(header)); err != nil {
	// 			return nil, err
	// 		}
	// 	}
	// 	if statusCode == 100 {
	// 		traceGot100Continue(cs.trace)
	// 		select {
	// 		case cs.on100 <- struct{}{}:
	// 		default:
	// 		}
	// 	}
	// 	cs.pastHeaders = false // do it all again
	// 	return nil, nil
	// }

	// res.ContentLength = -1
	// if clens := res.Header["Content-Length"]; len(clens) == 1 {
	// 	if cl, err := strconv.ParseUint(clens[0], 10, 63); err == nil {
	// 		res.ContentLength = int64(cl)
	// 	} else {
	// 		// TODO: care? unlike http/1, it won't mess up our framing, so it's
	// 		// more safe smuggling-wise to ignore.
	// 	}
	// } else if len(clens) > 1 {
	// 	// TODO: care? unlike http/1, it won't mess up our framing, so it's
	// 	// more safe smuggling-wise to ignore.
	// } else if f.StreamEnded() && !cs.isHead {
	// 	res.ContentLength = 0
	// }

	// if cs.isHead {
	// 	res.Body = noBody
	// 	return res, nil
	// }

	// if f.StreamEnded() {
	// 	if res.ContentLength > 0 {
	// 		res.Body = missingBody{}
	// 	} else {
	// 		res.Body = noBody
	// 	}
	// 	return res, nil
	// }

	// cs.bufPipe.setBuffer(&dataBuffer{expected: res.ContentLength})
	// cs.bytesRemain = res.ContentLength
	// res.Body = transportResponseBody{cs}

	// if cs.requestedGzip && asciiEqualFold(res.Header.Get("Content-Encoding"), "gzip") {
	// 	res.Header.Del("Content-Encoding")
	// 	res.Header.Del("Content-Length")
	// 	res.ContentLength = -1
	// 	res.Body = &gzipReader{body: res.Body}
	// 	res.Uncompressed = true
	// }

	return res, nil

}

// foreachHeaderElement splits v according to the "#rule" construction
// in RFC 7230 section 7 and calls fn for each non-empty element.
func foreachHeaderElement(v string, fn func(string)) {
	v = textproto.TrimString(v)
	if v == "" {
		return
	}
	if !strings.Contains(v, ",") {
		fn(v)
		return
	}
	for _, f := range strings.Split(v, ",") {
		if f = textproto.TrimString(f); f != "" {
			fn(f)
		}
	}
}

func hasStatusHeader(f *http2.MetaHeadersFrame) bool {
	status := f.PseudoValue("status")
	return len(status) > 0
}

// DumpMetaHeaders
func DumpMetaHeaders(f *http2.MetaHeadersFrame) (string, error) {

	// HTTPResponse MetaHeaders
	if hasStatusHeader(f) {
		res, err := processMetaHeadersForResponse(f)
		if err != nil {
			return "", err
		}
		dump, err := httputil.DumpResponse(res, false)
		if err != nil {
			return "", err
		}
		return string(dump), nil
	}

	// HTTPRequest MetaHeaders
	req, err := processMetaHeadersForRequest(f)
	if err != nil {
		return "", err
	}

	dump, err := httputil.DumpRequest(req, false)
	if err != nil {
		return "", err
	}
	return string(dump), nil
}
