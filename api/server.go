package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/alexferl/golib/http/router"
	"github.com/alexferl/golib/http/server"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"

	"github.com/alexferl/uberwachen/handlers"
	"github.com/alexferl/uberwachen/storage"
)

type (
	Handler struct {
		Storage storage.Storage
	}
)

// ErrorResponse holds an error message
type ErrorResponse struct {
	Message string `json:"error"`
}

func (h *Handler) root(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"message": "Uberwachen API"})
}

func (h *Handler) GetIncidents(c echo.Context) error {
	incidents := &[]handlers.Incident{}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := h.Storage.GetAll(ctx, incidents)
	if err != nil {
		m := fmt.Sprintf("Error getting incidents: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: m})
	}

	return c.JSON(http.StatusOK, map[string]*[]handlers.Incident{"incidents": incidents})
}

func (h *Handler) GetHandlers(c echo.Context) error {
	hs := viper.Get("handlers").(map[string]handlers.Handler)
	return c.JSON(http.StatusOK, map[string]map[string]handlers.Handler{"handlers": hs})
}

type Message struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func (h *Handler) HandlerSend(c echo.Context) error {
	name := c.Param("name")
	msg := new(Message)

	// TODO: add validation
	if err := c.Bind(msg); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{"bad request"})
	}

	hs := viper.Get("handlers").(map[string]handlers.Handler)

	if _, ok := hs[name]; ok {
		m := &handlers.Message{Title: msg.Title, Body: msg.Body, Type: handlers.MsgTypeNew}
		err := hs[name].Handler.Send(m)
		if err != nil {
			e := fmt.Sprintf("Error sending message: %v", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{e})
		}
	} else {
		e := fmt.Sprintf("Error handler '%s' not found", name)
		return c.JSON(http.StatusNotFound, ErrorResponse{e})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "message sent"})
}

// Start starts the API server
func Start() {
	s := server.New()
	h := &Handler{Storage: viper.Get("storage").(storage.Storage)}
	r := &router.Router{
		Routes: []router.Route{
			{"Root", http.MethodGet, "/", h.root},
			{"Incidents", http.MethodGet, "/incidents", h.GetIncidents},
			{"Handlers", http.MethodGet, "/stats", h.GetHandlers},
			{"HandlerSend", http.MethodPost, "/handlers/:name/send", h.HandlerSend},
		},
	}

	s.Start(r)
}
