## processing frame by http2 server

server -> readRawBytes -> convertToFrameStructs -> processFrame()

**Refer**

- https://cs.opensource.google/go/x/net/+/master:http2/server.go
- method: serverConn.processFrame()

## frame processing

- [processFrame(f Frame)](https://cs.opensource.google/go/x/net/+/master:http2/server.go;l=1497)

## rawBytes to Frame

```go
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
```

## Frame of interest

- **MetaHeadersFrame:** Contains actualy http data such as; httpHeaders

## different frame parsers

[Map of Parsers](https://cs.opensource.google/go/x/net/+/master:http2/frame.go;l=127?q=function:parse&ss=go%2Fx%2Fnet:http2%2F)
