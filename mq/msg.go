package mq

// NewMessage represents a message to be enqueued in IronMQ
type NewMessage struct {
	// The body of the message
	Body string `json:"body"`
	// The delay, in seconds until the message is available on the queue. Max is 604,800 (7 days)
	Delay uint32 `json:"delay"`
	// The push headers of the message. When creating a new message, ensure that this is non-nil
	PushHeaders map[string]string `json:"push_headers"`
}

// DequeuedMessage represents a message that has been dequeued from IronMQ.
type DequeuedMessage struct {
	ID            int    `json:"id"`
	Body          string `json:"body"`
	ReservedCount int    `json:"reserved_count"`
	ReservationID string `json:"reservation_id"`
}
