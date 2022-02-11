package registries

import (
	"errors"
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/alexferl/uberwachen/handlers"
)

type Handlers struct {
	mu       sync.Mutex
	handlers map[string]*handlers.Handler
}

func NewHandlers() *Handlers {
	m := make(map[string]*handlers.Handler)
	return &Handlers{
		handlers: m,
	}
}

func (h *Handlers) Register(handler *handlers.Handler) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	log.Debug().Msgf("Registering handler '%s'", handler.Name)

	if _, exist := h.handlers[handler.Name]; !exist {
		h.handlers[handler.Name] = handler
	} else {
		return errors.New(fmt.Sprintf("handler with name '%s' already registered", handler.Name))
	}

	return nil
}

func (h *Handlers) Get(name string) (*handlers.Handler, error) {
	if val, exist := h.handlers[name]; exist {
		return val, nil
	} else {
		return nil, errors.New(fmt.Sprintf("no handler with the name '%s' found", name))
	}
}
