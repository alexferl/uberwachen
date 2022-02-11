package main

import (
	"time"

	xconfig "github.com/alexferl/golib/config"
	xhttp "github.com/alexferl/golib/http/config"
	xlog "github.com/alexferl/golib/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config holds all configuration for our program
type Config struct {
	Config           *xconfig.Config
	Http             *xhttp.Config
	Logging          *xlog.Config
	ChecksPath       string
	CommandsPath     string
	HandlersPath     string
	RunChecksOnStart bool
	MongoDB          *MongoDB
}

// MongoDB holds all the configuration for the MongoDB storage
type MongoDB struct {
	URI                    string
	DatabaseName           string
	ReplicaSet             string
	Username               string
	Password               string
	ConnectTimeout         time.Duration
	ServerSelectionTimeout time.Duration
	SocketTimeout          time.Duration
}

// NewConfig creates a Config instance
func NewConfig() *Config {
	return &Config{
		Config:           xconfig.New(),
		Http:             xhttp.DefaultConfig,
		Logging:          xlog.DefaultConfig,
		ChecksPath:       "./examples/checks",
		CommandsPath:     "./examples/commands",
		HandlersPath:     "./examples/handlers",
		RunChecksOnStart: false,
		MongoDB: &MongoDB{
			URI:                    "mongodb://localhost:27017",
			DatabaseName:           "uberwachen",
			Username:               "",
			ReplicaSet:             "",
			Password:               "",
			ConnectTimeout:         5 * time.Second,
			ServerSelectionTimeout: 5 * time.Second,
			SocketTimeout:          30 * time.Second,
		},
	}
}

// addFlags adds all the flags from the command line
func (c *Config) addFlags(fs *pflag.FlagSet) {
	fs.StringVar(&c.ChecksPath, "checks-path", c.ChecksPath,
		"Path to the checks definition files")
	fs.StringVar(&c.CommandsPath, "commands-path", c.CommandsPath,
		"Path to the commands files")
	fs.StringVar(&c.HandlersPath, "handlers-path", c.HandlersPath,
		"Path to the handlers definition files")
	fs.BoolVar(&c.RunChecksOnStart, "run-checks-on-start", c.RunChecksOnStart,
		"Run checks when they're first registered and then on their normal schedule")

	// MongoDB
	fs.StringVar(&c.MongoDB.URI, "mongodb-uri", c.MongoDB.URI, "MongoDB URI")
	fs.StringVar(&c.MongoDB.DatabaseName, "mongodb-database-name", c.MongoDB.DatabaseName,
		"MongoDB database name")
	fs.StringVar(&c.MongoDB.ReplicaSet, "mongodb-replica-set", c.MongoDB.ReplicaSet, "MongoDB replica set")
	fs.StringVar(&c.MongoDB.Username, "mongodb-username", c.MongoDB.Username, "MongoDB username")
	fs.StringVar(&c.MongoDB.Password, "mongodb-password", c.MongoDB.Password, "MongoDB password")
	fs.DurationVar(&c.MongoDB.ConnectTimeout, "mongodb-connect-timeout", c.MongoDB.ConnectTimeout,
		"MongoDB connect timeout")
	fs.DurationVar(&c.MongoDB.ServerSelectionTimeout, "mongodb-server-selection-timeout",
		c.MongoDB.ServerSelectionTimeout, "MongoDB server selection timeout")
	fs.DurationVar(&c.MongoDB.SocketTimeout, "mongodb-socket-timeout", c.MongoDB.SocketTimeout,
		"MongoDB socket timeout")
}

func (c *Config) BindFlags() {
	c.addFlags(pflag.CommandLine)
	c.Logging.BindFlags(pflag.CommandLine)
	c.Http.BindFlags(pflag.CommandLine)

	err := c.Config.BindFlags()
	if err != nil {
		panic(err)
	}

	err = xlog.New(&xlog.Config{
		LogLevel:  viper.GetString("log-level"),
		LogOutput: viper.GetString("log-output"),
		LogWriter: viper.GetString("log-writer"),
	})
	if err != nil {
		panic(err)
	}
}
