## Parsehttp2frame: Convert rawBytes of http2Frames to string representation



## Usage
```go
import "github.com/h0x0er/parsehttp2frame"

rawBytes := []byte{} // http2 frame bytes

frame, _ := parsehttp2frame.BytesToHTTP2Frame()
frameString, _ := parsehttp2frame.Frame2String(frame)

fmt.Println(frameString)

```


## Use-cases

[] HeadersFrame To HttpRequest 

[] HeadersFrame to HttpResponse




