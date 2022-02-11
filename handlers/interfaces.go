package handlers

// HandlerSender is a common interface for all handlers
type HandlerSender interface {
	Send(msg *Message) error
}

// Handler represents a notification handler
type Handler struct {
	Name    string        `json:"name"`
	Type    string        `json:"type"`
	Handler HandlerSender `json:"handler,omitempty" bson:"-"`
}
