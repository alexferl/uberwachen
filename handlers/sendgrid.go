package handlers

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// sendGridHandler represents a SendGrid handler
type sendGridHandler struct {
	*Handler
	ApiKey          string `json:"api_key" bson:"api_key"`
	SubjectPrefix   string `json:"subject_prefix" bson:"subject_prefix"`
	From            string `json:"from"`
	FromName        string `json:"from_name" bson:"from_name"`
	To              string `json:"to"`
	ToName          string `json:"to_name" bson:"to_name"`
	NotifyOnResolve bool   `json:"notify_on_resolve" bson:"notify_on_resolve"`
}

// NewSendGridHandler creates a sendGridHandler instance
func NewSendGridHandler(apiKey, subjectPrefix, from, fromName, to, toName string, notifyOnResolve bool) *Handler {
	return &Handler{
		Type: "sendgrid",
		Handler: &sendGridHandler{
			ApiKey:          apiKey,
			SubjectPrefix:   subjectPrefix,
			From:            from,
			FromName:        fromName,
			To:              to,
			ToName:          toName,
			NotifyOnResolve: notifyOnResolve,
		},
	}
}

// Send sends an email via the SendGrid API
func (sg *sendGridHandler) Send(msg *Message) error {
	if sg.NotifyOnResolve {
		subject := fmt.Sprintf("%s %s", sg.SubjectPrefix, msg.Title)

		from := mail.NewEmail(sg.FromName, sg.From)
		to := mail.NewEmail(sg.ToName, sg.To)
		content := mail.NewContent("text/plain", msg.Body)
		m := mail.NewV3MailInit(from, subject, to, content)

		request := sendgrid.GetRequest(sg.ApiKey, "/v3/mail/send", "https://api.sendgrid.com")
		request.Method = "POST"
		request.Body = mail.GetRequestBody(m)
		response, err := sendgrid.API(request)
		if err != nil {
			return err
		} else {
			log.Debug().Msgf("SendGrid returned status code: '%d' body: '%s' headers: '%s'",
				response.StatusCode, response.Body, response.Headers)
		}
	}
	return nil
}
