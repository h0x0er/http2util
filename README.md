## http2util: Dump rawBytes of http2 frames to string, http.Request or http.Response

## Usage

```go
import "github.com/h0x0er/http2util"

rawBytes := []byte{} // http2 frame bytes

// Creating frame out of rawBytes
frame, _ := http2util.BytesToFrame(rawBytes)

// Usage 1: Dumping frame to string
frameString, _ := http2util.Dump(frame)
fmt.Println(frameString)

// Usage2: Creating http.Request from frame
req, _ = http2util.FrameToHTTPRequest(frame)

// Usage3: Creating http.Response from frame
res, _ = http2util.FrameToHTTPReponse(frame)

```

## Limitations

- `FrameToHTTPRequest()` & `FrameToHTTPResponse()`: Currently doesn't supports extraction of `httpBody` from frame.


## Contribution

Feel free to open an issue or send a PR for improvement
