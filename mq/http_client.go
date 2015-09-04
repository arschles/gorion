package mq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/arschles/gorion"
	"golang.org/x/net/context"
)

// Scheme is "http" or "https"
type Scheme string

// String converts a Scheme to a printable string
func (s Scheme) String() string {
	return string(s)
}

const (
	// SchemeHTTP represents http
	SchemeHTTP = "http"
	// SchemeHTTPS represents https
	SchemeHTTPS     = "https"
	applicationJSON = "application/json"
	oauth           = "OAuth"
)

type httpClient struct {
	scheme     Scheme
	host       string
	port       uint16
	transport  *http.Transport
	client     *http.Client
	oauthToken string
}

// NewHTTPClient returns a Client implementation that can talk to the IronMQ v3
// API documented at http://dev.iron.io/mq/3/reference/api/
func NewHTTPClient(scheme Scheme, host string, port uint16) Client {
	transport := &http.Transport{}
	client := &http.Client{Transport: transport}
	return &httpClient{
		scheme:    scheme,
		host:      host,
		port:      port,
		transport: transport,
		client:    client,
	}
}

// headers sets json and oauth headers on r
func (h *httpClient) newReq(method, token, projID, path string, body io.Reader) (*http.Request, error) {
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

// Enqueue posts messages to IronMQ using the API defined at http://dev.iron.io/mq/3/reference/api/#post-messages
func (h *httpClient) Enqueue(ctx context.Context, token, projID, qName string, msgs []NewMessage) (*Enqueued, error) {
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

// Dequeue gets messages from IronMQ using the API defined at http://dev.iron.io/mq/3/reference/api/#reserve-messages
func (h *httpClient) Dequeue(ctx context.Context, token, projID, qName string, num int, timeout Timeout, wait Wait, delete bool) ([]DequeuedMessage, error) {
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

func (h *httpClient) DeleteReserved(ctx context.Context, token, projID, qName string, messageID int, reservationID string) (*Deleted, error) {
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
