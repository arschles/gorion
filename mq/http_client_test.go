package mq

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/arschles/gorion/Godeps/_workspace/src/github.com/arschles/assert"
	"github.com/arschles/gorion/Godeps/_workspace/src/github.com/arschles/testsrv"
	"github.com/arschles/gorion/Godeps/_workspace/src/github.com/gorilla/mux"
	"github.com/arschles/gorion/Godeps/_workspace/src/golang.org/x/net/context"
)

var (
	bgCtx = context.Background()
)

type qServer struct {
	mem Client
}

func (q *qServer) enqueueHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		qName, ok := mux.Vars(r)["queue_name"]
		if !ok {
			http.Error(w, "missing queue name", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		req := new(enqueueReq)
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			http.Error(w, fmt.Sprintf("invalid json [%s]", err), http.StatusBadRequest)
			return
		}

		res, err := q.mem.Enqueue(bgCtx, token, projID, qName, req.Messages)
		if err != nil {
			http.Error(w, fmt.Sprintf("error enqueueing [%s]", err), http.StatusInternalServerError)
			return
		}
		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, fmt.Sprintf("error encoding enqueue response json [%s]", err), http.StatusInternalServerError)
			return
		}
	})
}

func (q *qServer) dequeueHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		qName, ok := mux.Vars(r)["queue_name"]
		if !ok {
			http.Error(w, "missing queue name", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		req := new(dequeueReq)
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			http.Error(w, fmt.Sprintf("invalid json [%s]", err), http.StatusBadRequest)
			return
		}

		msgs, err := q.mem.Dequeue(bgCtx, token, projID, qName, req.Num, Timeout(req.Timeout), Wait(req.Wait), false)
		if err != nil {
			http.Error(w, fmt.Sprintf("dequeue error [%s]", err), http.StatusInternalServerError)
			return
		}
		if err := json.NewEncoder(w).Encode(dequeueResp{Messages: msgs}); err != nil {
			http.Error(w, fmt.Sprintf("error encoding dequeue response json [%s]", err), http.StatusInternalServerError)
			return
		}
	})
}

func (q *qServer) deleteReservedHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		qName, ok := mux.Vars(r)["queue_name"]
		if !ok {
			http.Error(w, "missing queue name", http.StatusBadRequest)
			return
		}
		msgIDStr, ok := mux.Vars(r)["message_id"]
		if !ok {
			http.Error(w, "missing message ID", http.StatusBadRequest)
			return
		}
		msgID, err := strconv.Atoi(msgIDStr)
		if err != nil {
			http.Error(w, "message ID must be an int", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		req := new(deleteReservedReq)
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			http.Error(w, fmt.Sprintf("invalid json [%s]", err), http.StatusBadRequest)
			return
		}
		ret, err := q.mem.DeleteReserved(bgCtx, token, projID, qName, msgID, req.ReservationID)
		if err != nil {
			http.Error(w, fmt.Sprintf("error deleting reserved msg [%s]", err), http.StatusInternalServerError)
			return
		}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			http.Error(w, fmt.Sprintf("error encoding response json [%s]", err), http.StatusInternalServerError)
			return
		}
	})
}

func makeQHandler() http.Handler {
	srv := &qServer{mem: NewMemClient()}
	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, fmt.Sprintf(`{"msg":"path %s not found"`, r.URL), http.StatusNotFound)
	})
	r.Handle("/3/projects/{project_id}/queues/{queue_name}/messages", srv.enqueueHandler()).Methods("POST")
	r.Handle("/3/projects/{project_id}/queues/{queue_name}/reservations", srv.dequeueHandler()).Methods("POST")
	r.Handle("/3/projects/{project_id}/queues/{queue_name}/messages/{message_id}", srv.deleteReservedHandler()).Methods("DELETE")
	return r
}

func TestHTTPQueueOperations(t *testing.T) {
	srv := testsrv.StartServer(makeQHandler())
	defer srv.Close()
	urlStrSplit := strings.Split(strings.TrimPrefix(srv.URLStr(), "http://"), ":")
	assert.Equal(t, 2, len(urlStrSplit), "number of elements in the URL string")
	host := urlStrSplit[0]
	port, err := strconv.Atoi(urlStrSplit[1])
	assert.NoErr(t, err)
	if port > 65535 {
		t.Fatalf("port [%d] not a uint16", port)
	}
	cl := NewHTTPClient(SchemeHTTP, host, uint16(port))
	assert.NoErr(t, qOperations(cl))
}
