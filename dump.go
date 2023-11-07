package http2util

// Note: Code is taken from https://cs.opensource.google/go/x/net/+/master:http2

import (
	"fmt"

	"golang.org/x/net/http2"
)

func Dump(f http2.Frame) (string, error) {
	switch f := f.(type) {
	case *http2.SettingsFrame:
		return fmt.Sprintf("%#v", f), nil
	case *http2.MetaHeadersFrame:
		return DumpMetaHeaders(f)
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
