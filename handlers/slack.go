package handlers

import (
	"fmt"

	"github.com/nlopes/slack"
	"github.com/rs/zerolog/log"
)

// SlackHandler represents a Slack handler
type SlackHandler struct {
	*Handler
	Channel     string `json:"channel"`
	Token       string `json:"token"`
	BotUsername string `json:"bot_username" bson:"bot_username"`
	BotIconUrl  string `json:"bot_icon_url" bson:"bot_icon_url"`
}

type Slack SlackHandler

// NewSlackHandler creates a SlackHandler instance
func NewSlackHandler(channel, token, botUsername, botIconUrl string) Handler {
	return Handler{
		Type: "slack",
		Handler: &SlackHandler{
			Channel:     channel,
			Token:       token,
			BotUsername: botUsername,
			BotIconUrl:  botIconUrl,
		},
	}
}

// Send sends a message to a Slack channel
func (s *SlackHandler) Send(msg *Message) error {
	text := fmt.Sprintf("%s \n %s", msg.Title, msg.Body)
	var color string

	if msg.Type == MsgTypeNew {
		color = "#DF0101"
	} else if msg.Type == MsgTypeResolve {
		color = "#33FF33"
	}

	api := slack.New(s.Token)
	params := slack.PostMessageParameters{
		Username: s.BotUsername,
		IconURL:  s.BotIconUrl,
	}
	attachment := slack.Attachment{
		Color: color,
		Text:  text,
	}
	channelId, timestamp, err := api.PostMessage(
		s.Channel,
		slack.MsgOptionPostMessageParameters(params),
		slack.MsgOptionAttachments(attachment))
	if err != nil {
		return err
	}

	log.Debug().Msgf("Message successfully sent to channel '%s' at '%s'", channelId, timestamp)
	return nil
}
