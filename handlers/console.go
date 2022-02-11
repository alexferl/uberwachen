package handlers

import (
	"fmt"
)

// consoleHandler represents a console handler
type consoleHandler struct {
	*Handler
}

// NewConsoleHandler creates a consoleHandler instance
func NewConsoleHandler() *Handler {
	return &Handler{
		Name:    "console",
		Type:    "console",
		Handler: &consoleHandler{},
	}
}

// Send outputs the message to the console
func (c *consoleHandler) Send(msg *Message) error {
	fmt.Printf("%s: %s\n", msg.Title, msg.Body)
	return nil
}
