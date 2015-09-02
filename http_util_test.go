package gorion

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/arschles/gorion/Godeps/_workspace/src/github.com/arschles/assert"
	"github.com/arschles/gorion/Godeps/_workspace/src/github.com/arschles/testsrv"
	"github.com/arschles/gorion/Godeps/_workspace/src/golang.org/x/net/context"
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
	recv := srv.AcceptN(1, 100*time.Second)
	assert.Equal(t, 1, len(recv), "number of received requests")
}
