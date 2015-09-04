package mq

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/arschles/gorion"
	"golang.org/x/net/context"
)

// Scheme is "http" or "https"
type Scheme string

// SchemeFromString returns a Scheme representation of s. If s is not a supported scheme, returns ErrInvalidScheme
func SchemeFromString(s string) (Scheme, error) {
	switch s {
	case SchemeHTTP:
		return SchemeHTTP, nil
	case SchemeHTTPS:
		return SchemeHTTPS, nil
	default:
		return Scheme(""), ErrInvalidScheme
	}
}

// String converts a Scheme to a printable string
func (s Scheme) String() string {
	return string(s)
}

var (
	// ErrInvalidScheme is returned from any func that converts something to a Scheme when the value is an invalid scheme
	ErrInvalidScheme = errors.New("invalid scheme")
)

const (
	// SchemeHTTP represents http
	SchemeHTTP = "http"
	// SchemeHTTPS represents https
	SchemeHTTPS     = "https"
	applicationJSON = "application/json"
	oauth           = "OAuth"
)

// HTTPClient is a Client implementation that talks to an arbitrary IronMQ v3 API (http://dev.iron.io/mq/3/reference/api/).
// Use NewHTTPClient to create a new one of these.
type HTTPClient struct {
	scheme     Scheme
	host       string
	port       uint16
	transport  *http.Transport
	client     *http.Client
	oauthToken string
}

// NewHTTPClient returns a new HTTPClient that talks to the IronMQ v3 API at {scheme}://{host}:{port}
func NewHTTPClient(scheme Scheme, host string, port uint16) *HTTPClient {
	transport := &http.Transport{}
	client := &http.Client{Transport: transport}
	return &HTTPClient{
		scheme:    scheme,
		host:      host,
		port:      port,
		transport: transport,
		client:    client,
	}
}

// headers sets json and oauth headers on r
func (h *HTTPClient) newReq(method, token, projID, path string, body io.Reader) (*http.Request, error) {
	urlStr := fmt.Sprintf("%s://%s:%d/3/projects/%s/%s", h.scheme, h.host, h.port, projID, path)
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "OAuth "+token)
	return req, nil
}

type enqueueReq struct {
	Messages []NewMessage `json:"messages"`
}

// Enqueue is the Client implementation for the v3 API http://dev.iron.io/mq/3/reference/api/#post-messages
func (h *HTTPClient) Enqueue(ctx context.Context, token, projID, qName string, msgs []NewMessage) (*Enqueued, error) {
	reqBody := &bytes.Buffer{}
	if err := json.NewEncoder(reqBody).Encode(enqueueReq{Messages: msgs}); err != nil {
		return nil, err
	}

	req, err := h.newReq("POST", token, projID, fmt.Sprintf("queues/%s/messages", qName), reqBody)
	if err != nil {
		return nil, err
	}
	ret := new(Enqueued)
	doFunc := func(resp *http.Response, err error) error {
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(ret); err != nil {
			return err
		}
		return nil
	}
	if err := gorion.HTTPDo(ctx, h.client, h.transport, req, doFunc); err != nil {
		return nil, err
	}
	return ret, nil
}

type dequeueReq struct {
	Num     int  `json:"n"`
	Timeout int  `json:"timeout"`
	Wait    int  `json:"wait"`
	Delete  bool `json:"delete"`
}

type dequeueResp struct {
	Messages []DequeuedMessage `json:"messages"`
}

// Dequeue is the client implementation for the v3 API (http://dev.iron.io/mq/3/reference/api/#reserve-messages)
func (h *HTTPClient) Dequeue(ctx context.Context, token, projID, qName string, num int, timeout Timeout, wait Wait, delete bool) ([]DequeuedMessage, error) {
	if !timeoutInRange(timeout) {
		return nil, ErrTimeoutOutOfRange
	}
	if !waitInRange(wait) {
		return nil, ErrWaitOutOfRange
	}

	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(dequeueReq{Num: num, Timeout: int(timeout), Wait: int(wait), Delete: delete}); err != nil {
		return nil, err
	}
	req, err := h.newReq("POST", token, projID, fmt.Sprintf("queues/%s/reservations", qName), body)
	if err != nil {
		return nil, err
	}
	ret := new(dequeueResp)
	doFunc := func(resp *http.Response, err error) error {
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(ret); err != nil {
			return err
		}
		return nil
	}
	if err := gorion.HTTPDo(ctx, h.client, h.transport, req, doFunc); err != nil {
		return nil, err
	}
	return ret.Messages, nil
}

type deleteReservedReq struct {
	ReservationID string `json:"reservation_id"`
}

// DeleteReserved is the client implementation for the IronMQ v3 API (http://dev.iron.io/mq/3/reference/api/#delete-message)
func (h *HTTPClient) DeleteReserved(ctx context.Context, token, projID, qName string, messageID int, reservationID string) (*Deleted, error) {
	body := &bytes.Buffer{}
	if err := json.NewEncoder(body).Encode(deleteReservedReq{ReservationID: reservationID}); err != nil {
		return nil, err
	}
	req, err := h.newReq("DELETE", token, projID, fmt.Sprintf("queues/%s/messages/%d", qName, messageID), body)
	if err != nil {
		return nil, err
	}
	ret := new(Deleted)
	doFunc := func(resp *http.Response, err error) error {
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(ret); err != nil {
			return err
		}
		return nil
	}
	if err := gorion.HTTPDo(ctx, h.client, h.transport, req, doFunc); err != nil {
		return nil, err
	}
	return ret, nil
}
