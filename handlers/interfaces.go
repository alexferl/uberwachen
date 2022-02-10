package handlers

// HandlerSender is a common interface for all handlers
type HandlerSender interface {
	Send(msg *Message) error
}

// Handler represents a notification handler
type Handler struct {
	Type    string        `json:"type"`
	Handler HandlerSender `json:"handler,omitempty" bson:"-"`
}
