package gorion

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/arschles/assert"
	"github.com/arschles/testsrv"
	"golang.org/x/net/context"
)

func TestHTTPDo(t *testing.T) {
	hndl := func(http.ResponseWriter, *http.Request) {}
	srv := testsrv.StartServer(http.HandlerFunc(hndl))
	defer srv.Close()
	transport := &http.Transport{}
	client := &http.Client{Transport: transport}
	req, err := http.NewRequest("GET", srv.URLStr(), strings.NewReader(""))
	assert.NoErr(t, err)
	err = HTTPDo(context.Background(), client, transport, req, func(*http.Response, error) error {
		return nil
	})
	assert.NoErr(t, err)
	recv := srv.AcceptN(1, 100*time.Millisecond)
	assert.Equal(t, 1, len(recv), "number of received requests")
}

func TestHTTPDoAfterCancel(t *testing.T) {
	hndl := func(http.ResponseWriter, *http.Request) {}
	srv := testsrv.StartServer(http.HandlerFunc(hndl))
	defer srv.Close()
	transport := &http.Transport{}
	client := &http.Client{Transport: transport}
	req, err := http.NewRequest("GET", srv.URLStr(), strings.NewReader(""))
	assert.NoErr(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = HTTPDo(ctx, client, transport, req, func(*http.Response, error) error {
		return nil
	})
	assert.Err(t, ErrCancelled, err)
	recv := srv.AcceptN(1, 100*time.Millisecond)
	assert.Equal(t, 0, len(recv), "number of received requests")
}
