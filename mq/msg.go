package mq

type NewMessage struct {
	Body        string            `json:"body"`
	Delay       uint16            `json:"delay"`
	PushHeaders map[string]string `json:"push_headers"`
}

type DequeuedMessage struct {
	ID            int    `json:"id"`
	Body          string `json:"body"`
	ReservedCount int    `json:"reserved_count"`
	ReservationID string `json:"reservation_id"`
}
