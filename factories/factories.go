package factories

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/alexferl/uberwachen/handlers"
	"github.com/alexferl/uberwachen/storage"
)

// Handler creates a new object with handlers.HandlerSender interface
func Handler(handlerType string, handlerConfig map[string]interface{}) handlers.Handler {
	switch handlerType {
	case "console":
		return handlers.NewConsoleHandler()

	case "sendgrid":
		apiKey := handlerConfig["apiKey"].(string)
		subjectPrefix := handlerConfig["subjectPrefix"].(string)
		from := handlerConfig["from"].(string)
		fromName := handlerConfig["fromName"].(string)
		to := handlerConfig["to"].(string)
		toName := handlerConfig["toName"].(string)
		notifyOnResolve := handlerConfig["notifyOnResolve"].(bool)
		return handlers.NewSendGridHandler(apiKey, subjectPrefix, from, fromName, to, toName, notifyOnResolve)

	case "slack":
		channel := handlerConfig["channel"].(string)
		token := handlerConfig["token"].(string)
		botUsername := handlerConfig["botUsername"].(string)
		botIconUrl := handlerConfig["botIconUrl"].(string)
		return handlers.NewSlackHandler(channel, token, botUsername, botIconUrl)

	default:
		log.Warn().Msgf("Unknown handler type '%s'", handlerType)
		var h handlers.Handler
		return h
	}
}

// Backend creates a new object with handlers.Backend interface
func Backend() (storage.Storage, error) {
	opts := &storage.MongoDBOpts{
		URI:                    viper.GetString("mongodb-uri"),
		DatabaseName:           viper.GetString("mongodb-database-name"),
		Username:               viper.GetString("mongodb-username"),
		Password:               viper.GetString("mongodb-password"),
		ConnectTimeout:         viper.GetDuration("mongodb-connect-timeout"),
		ServerSelectionTimeout: viper.GetDuration("mongodb-server-selection-timeout"),
		SocketTimeout:          viper.GetDuration("mongodb-socket-timeout"),
	}
	return storage.NewMongoDB(opts)
}
