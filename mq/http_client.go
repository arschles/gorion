package mq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/arschles/gorion"
	"github.com/arschles/gorion/Godeps/_workspace/src/golang.org/x/net/context"
)

type Scheme string

func (s Scheme) String() string {
	return string(s)
}

const (
	SchemeHTTP  = "http"
	SchemeHTTPS = "https"
)

type httpClient struct {
	endpt     string
	transport *http.Transport
	client    *http.Client
}

func NewHTTPClient(scheme Scheme, host string, port uint16) Client {
	transport := &http.Transport{}
	client := &http.Client{Transport: transport}
	return &httpClient{
		transport: transport,
		client:    client,
		endpt:     fmt.Sprintf("%s://%s:%d", scheme, host, port),
	}
}

type enqueueReq struct {
	Messages []NewMessage `json:"messages"`
}

func (h *httpClient) Enqueue(ctx context.Context, queueName string, msgs []NewMessage) (*Enqueued, error) {
	reqBody := &bytes.Buffer{}
	if err := json.NewEncoder(reqBody).Encode(enqueueReq{Messages: msgs}); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", h.urlStr("queues/%s/messages", queueName), reqBody)
	if err != nil {
		return nil, err
	}
	var ret *Enqueued
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

func (h *httpClient) Dequeue(ctx context.Context, qName string, num int, timeout Timeout, wait Wait, delete bool) ([]DequeuedMessage, error) {
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
	req, err := http.NewRequest("POST", h.urlStr("queues/%s/reservations", qName), body)
	if err != nil {
		return nil, err
	}
	var ret *dequeueResp
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

// urlStr returns the url string resulting from appending path to h.endpt.
// pass path without a leading slash
func (h *httpClient) urlStr(pathFmt string, fmtVars ...interface{}) string {
	return h.endpt + "/" + fmt.Sprintf(pathFmt, fmtVars...)
}
