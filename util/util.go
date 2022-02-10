package util

import (
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/ventu-io/go-shortid"
)

// GenerateShortId generates and returns a short id
func GenerateShortId() string {
	id, err := shortid.Generate()
	if err != nil {
		log.Error().Msg("Error generating short id")
		return ""
	}
	return id
}

// GetCmdPath returns to full path of a check command
func GetCmdPath(cmd string) string {
	commandsDir := viper.GetString("commands-path")
	c, _ := SplitCmd(cmd)
	return commandsDir + "/" + c
}

// SplitCmd splits a command string into a slice of strings
func SplitCmd(cmd string) (string, []string) {
	c := strings.Split(cmd[:], " ")
	return c[0], c[1:]
}

// PathExists check if a specific path exists
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return true, err
}
