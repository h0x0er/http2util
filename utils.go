package parsehttp2frame

import (
	"bufio"
	"bytes"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
)

func BytesToHTTP2Frame(b []byte) (http2.Frame, error) {

	rd := bytes.NewReader(b)
	buf := bufio.NewReader(rd)

	fr := http2.NewFramer(nil, buf)
	fr.ReadMetaHeaders = hpack.NewDecoder(0, nil)
	f, err := fr.ReadFrame()
	if err != nil {
		return nil, err
	}
	return f, nil
}

func GetFrameType(f http2.Frame) http2.FrameType {
	return f.Header().Type
}

func hasStatusHeader(f *http2.MetaHeadersFrame) bool {
	status := f.PseudoValue("status")
	return len(status) > 0
}
