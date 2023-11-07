package parsehttp2frame

import (
	"fmt"
	"net/http/httputil"

	"golang.org/x/net/http2"
)

func Frame2String(f http2.Frame) (string, error) {
	switch ftype := f.(type) {
	case *http2.SettingsFrame:
		return fmt.Sprintf("%#v", f), nil
	case *http2.MetaHeadersFrame:
		req, err := processHeaders(f)
		if err != nil {
			return "", err
		}
		b, err := httputil.DumpRequest(req, false)
		return string(b), err
	case *http2.WindowUpdateFrame:
		return fmt.Sprintf("%#v", f), nil
	case *http2.PingFrame:
		return fmt.Sprintf("%#v", f), nil
	case *http2.DataFrame:
		return fmt.Sprintf("%#v", f), nil
	case *http2.RSTStreamFrame:
		return fmt.Sprintf("%#v", f), nil
	case *http2.PriorityFrame:
		return fmt.Sprintf("%#v", f), nil
	case *http2.GoAwayFrame:
		return fmt.Sprintf("%#v", f), nil
	case *http2.PushPromiseFrame:
		return fmt.Sprintf("%#v", f), nil
	default:
		return "", fmt.Errorf("[Frame2String] unable to handle frame")
	}
}
