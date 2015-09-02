package mq

// NewMessage represents a message to be enqueued in IronMQ
type NewMessage struct {
	Body        string            `json:"body"`
	Delay       uint16            `json:"delay"`
	PushHeaders map[string]string `json:"push_headers"`
}

// DequeuedMessage represents a message that has been dequeued from IronMQ
type DequeuedMessage struct {
	ID            int    `json:"id"`
	Body          string `json:"body"`
	ReservedCount int    `json:"reserved_count"`
	ReservationID string `json:"reservation_id"`
}
