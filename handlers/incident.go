package handlers

import (
	"time"

	"github.com/alexferl/uberwachen/util"
)

type Incident struct {
	*Check
	ID            string    `json:"id" bson:"_id"`
	Message       *Message  `json:"-" bson:"-"`
	Name          string    `json:"name"`
	CreatedAt     time.Time `json:"created_at" bson:"created_at"`
	LastUpdatedAt time.Time `json:"last_updated_at" bson:"last_updated_at"`
}

func NewIncident(c *Check) *Incident {
	id := util.GenerateShortId()
	c.Attempts = 1
	return &Incident{
		ID:        id,
		CreatedAt: time.Now().UTC(),
		Check:     c,
		Name:      c.Name,
	}
}

func (i *Incident) Update(output string) {
	i.Check.Output = output
	i.Check.Attempts += 1
	i.LastUpdatedAt = time.Now().UTC()
}
