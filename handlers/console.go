package handlers

import (
	"fmt"
)

// ConsoleHandler represents a console handler
type ConsoleHandler struct {
	*Handler
}

// NewConsoleHandler creates a ConsoleHandler instance
func NewConsoleHandler() Handler {
	return Handler{
		Type:    "console",
		Handler: &ConsoleHandler{},
	}
}

// Send outputs the message to the console
func (c *ConsoleHandler) Send(msg *Message) error {
	fmt.Printf("%s: %s\n", msg.Title, msg.Body)
	return nil
}
