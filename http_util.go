package gorion

import (
	"net/http"

	"github.com/arschles/gorion/Godeps/_workspace/src/golang.org/x/net/context"
)

// HTTPDo runs the HTTP request in a goroutine and passes the response to f in
// that same goroutine. After the goroutine finishes, returns the result of f.
// If ctx.Done() receives before f returns, calls transport.CancelRequest(req) before
// before returning. Even though HTTPDo executes f in another goroutine, you
// can treat it as a synchronous call to f. Also, since HTTPDo uses client to run
// requests and transport to cancel them, client.Transport should be transport
// in most cases.
//
// Example Usage:
//  type Resp struct { Num int `json:"num"` }
//  var resp *Resp
//  err := HttpDo(ctx, client, transport, req, func(resp *http.Response, err error) error {
//    if err != nil { return err }
//    defer resp.Body.Close()
//
//    if err := json.NewDecoder(resp.Body).Decode(resp); err != nil {
//      return err
//    }
//    return nil
//  })
//  if err != nil { return err }
//  // do something with resp...
//
// This func was stolen/adapted from https://blog.golang.org/context
func HTTPDo(ctx context.Context, client *http.Client, transport *http.Transport, req *http.Request, f func(*http.Response, error) error) error {
	// Run the HTTP request in a goroutine and pass the response to f.
	c := make(chan error, 1)

	go func() {
		c <- f(client.Do(req))
	}()

	select {
	case <-ctx.Done():
		transport.CancelRequest(req)
		<-c // Wait for f to return.
		return ctx.Err()
	case err := <-c:
		return err
	}
}
