package mq

import (
	"fmt"

	"github.com/arschles/gorion/Godeps/_workspace/src/golang.org/x/net/context"
)

const (
	token  = "test-token"
	qName  = "test-queue"
	projID = "test-proj"
)

func qOperations(cl Client) error {
	newMsgs := []NewMessage{{Body: "123", Delay: 0, PushHeaders: make(map[string]string)}}
	ctx := context.Background()
	enqRes, err := cl.Enqueue(ctx, token, projID, qName, newMsgs)
	if err != nil {
		return fmt.Errorf("got error on enqueue [%s]", err)
	}
	if len(enqRes.IDs) != 1 {
		return fmt.Errorf("Enqueue returned [%d] message IDs, expected 1", len(enqRes.IDs))
	}
	// dequeue & reserve
	dqMsgs, err := cl.Dequeue(ctx, token, projID, qName, 1, Timeout(30), Wait(1), false)
	if err != nil {
		return fmt.Errorf("got error on dequeue [%s]", err)
	}
	if len(newMsgs) != len(dqMsgs) {
		return fmt.Errorf("number of dequeued messages [%d] != number of enqueued messages [%d]", len(dqMsgs), len(newMsgs))
	}
	for i, dqMsg := range dqMsgs {
		msg := newMsgs[i]
		if msg.Body != dqMsg.Body {
			return fmt.Errorf("message # [%d] body [%s] isn't dequeued message body [%s]", i, msg.Body, dqMsg.Body)
		}
		if dqMsg.ReservedCount != 1 {
			return fmt.Errorf("dequeued message was reserved [%d] times", dqMsg.ReservedCount)
		}
		deletedRes, err := cl.DeleteReserved(ctx, token, projID, qName, dqMsg.ID, dqMsg.ReservationID)
		if err != nil {
			return fmt.Errorf("DeleteReserved returned error [%s]", err)
		}
		if len(deletedRes.Msg) <= 0 {
			return fmt.Errorf("DeleteReserved returned an empty message")
		}
	}
	return nil
}
