package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/alexferl/uberwachen/storage"
	"github.com/alexferl/uberwachen/util"
)

const (
	MsgTypeNew     = "new"
	MsgTypeResolve = "resolve"
)

type Event struct {
	*Check
	ID        string    `json:"id" bson:"_id"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

type Message struct {
	Body  string `json:"body"`
	Title string `json:"title"`
	Type  string `json:"type"`
}

func NewEvent(c *Check) *Event {
	id := util.GenerateShortId()
	log.Debug().Msgf("Creating event '%s' for check '%s'", id, c.Name)
	return &Event{
		ID:        id,
		CreatedAt: time.Now().UTC(),
		Check:     c,
	}
}

func (e *Event) Process() {
	db := viper.Get("storage").(storage.Storage)
	log.Debug().Msgf("Processing event '%s' for check '%s'", e.ID, e.Check.Name)

	if e.getHandlers() == nil {
		return
	}

	if e.Check.Status != 0 {
		incident, err := e.getIncident()
		if err != nil {
			log.Error().Msgf("Error getting incident from database: %v", err)
		}

		if incident == nil { // new incident
			incident = NewIncident(e.Check)

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			err := db.Set(ctx, incident)
			if err != nil {
				log.Error().Msgf("Error saving incident to database: %v", err)
			}

			log.Debug().Msgf("Created new incident '%s'", incident.ID)

			msg := &Message{
				Body: fmt.Sprintf("%s", incident.Check.Output),
				Title: fmt.Sprintf("Incident '%s' started - Check '%s' failed after %d attempts",
					incident.ID, incident.Check.Name, incident.Check.Attempts),
				Type: MsgTypeNew,
			}

			e.handle(msg)
		} else { // existing incident
			incident.Update(e.Check.Output)
			log.Debug().Msgf("Existing incident '%s' found", incident.ID)

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			err := db.Update(ctx, incident.Check.Name, incident)
			if err != nil {
				log.Error().Msgf("Error updating incident to database: %v", err)
			}
		}

		if incident.Check.Attempts >= incident.Check.MaxAttempts {
			if e.Check.Renotify && incident.Check.PreviousOutput != e.Check.Output {
				log.Debug().Msgf("Check '%s' failed with a different output: previous: '%v' current: '%v'",
					incident.Check.Name, incident.Check.PreviousOutput, e.Check.Output)

				msg := &Message{
					Body: fmt.Sprintf("%s", incident.Check.Output),
					Title: fmt.Sprintf("Incident '%s' updated - Check '%s' failed with a different output",
						incident.ID, incident.Check.Name),
					Type: MsgTypeNew,
				}
				e.handle(msg)
			}
		}
	} else {
		incident, err := e.getIncident()
		if err != nil {
			log.Error().Msgf("Error getting incident from database: %v", err)
		}

		if incident != nil {
			if incident.Check.Attempts >= incident.Check.MaxAttempts {
				msg := &Message{
					Body:  fmt.Sprintf("%s", e.Check.Output),
					Title: fmt.Sprintf("Incident '%s' resolved - Check '%s' passed", incident.ID, incident.Name),
					Type:  MsgTypeResolve,
				}
				e.handle(msg)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			err := db.Delete(ctx, incident.Name)
			if err != nil {
				log.Error().Msgf("Error deleting incident from database: %v", err)
			}
		}
	}
}

func (e *Event) getIncident() (*Incident, error) {
	db := viper.Get("storage").(storage.Storage)
	incident := &Incident{}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := db.Get(ctx, e.Check.Name, incident)
	if err != nil {
		return nil, err
	}

	if incident.ID == "" {
		return nil, nil
	}

	incident.Check.PreviousOutput = incident.Check.Output

	return incident, nil
}

func (e *Event) handle(msg *Message) {
	for _, handler := range e.Check.Handlers {
		if e.Check.Status == 0 || e.Check.Status == 2 {
			err := handler.Handler.Send(msg)
			if err != nil {
				log.Error().Msgf("Error sending message: %v", err)
			}
		}
	}
}

func (e *Event) getHandlers() []*Handler {
	var handlers []*Handler

	if len(e.Check.Handlers) > 0 {
		for _, handler := range e.Check.Handlers {
			handlers = append(handlers, handler)
		}
		return handlers
	} else {
		log.Error().Msgf("Check '%s' does not have a handler", e.Check.Name)
		return nil
	}
}
