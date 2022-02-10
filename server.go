package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/jasonlvhit/gocron"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/alexferl/uberwachen/api"
	"github.com/alexferl/uberwachen/factories"
	"github.com/alexferl/uberwachen/handlers"
	"github.com/alexferl/uberwachen/loaders"
	"github.com/alexferl/uberwachen/util"
)

func main() {
	c := NewConfig()
	c.BindFlags()

	createFolders()

	log.Info().Msg("Connecting to database")
	loadBackend()

	log.Info().Msg("Adding handlers")
	loadHandlers()

	log.Info().Msg("Adding and scheduling checks")
	s := loadAndScheduleChecks()

	log.Info().Msg("Starting HTTP API")
	go api.Start()

	log.Info().Msg("Starting scheduler")
	<-s.Start()
}

func createFolders() {
	createFolder("checks-path")
	createFolder("commands-path")
	createFolder("handlers-path")
}

func createFolder(keyName string) {
	path := viper.GetString(keyName)
	pathExists, pathErr := util.PathExists(path)

	if pathErr != nil {
		log.Panic().Msgf("Error checking %s: %v", keyName, pathErr)
		os.Exit(1)
	}

	if !pathExists {
		err := os.MkdirAll(path, 0o750)
		if err != nil {
			log.Panic().Msgf("Error creating %s: %v", keyName, err)
			os.Exit(1)
		}
	}
}

func loadBackend() {
	db, err := factories.Backend()
	if err != nil {
		log.Panic().Msgf("Error creating backend: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbErr := db.Init(ctx)
	if dbErr != nil {
		log.Error().Msgf("Error initializing database: %v", dbErr)
	}

	viper.Set("storage", db)
}

func loadHandlers() {
	handlersMap := make(map[string]handlers.Handler)
	handlersMap["console"] = handlers.NewConsoleHandler()

	err := walk(handlersMap)
	if err != nil {
		log.Error().Msgf("Error adding handlers: %v", err)
	}

	viper.Set("handlers", handlersMap)
}

func loadAndScheduleChecks() *gocron.Scheduler {
	fileLoader := loaders.NewFileLoader(viper.GetString("checks-path"))
	flErr := fileLoader.Load()
	if flErr != nil {
		log.Error().Msgf("Error adding checks: %v", flErr)
	}

	s := gocron.NewScheduler()
	for _, c := range fileLoader.(*loaders.FileLoader).Checks {
		cmdPath := util.GetCmdPath(c.Command)
		exists, err := util.PathExists(cmdPath)
		if err != nil {
			log.Error().Msgf("Error checking if command exists: %v", err)
		}

		if exists {
			s.Every(uint64(c.Interval)).Seconds().Do(handlers.RunCheck, c)
			log.Info().Msgf("Scheduling check '%s'", c.Name)
			go handlers.RunCheck(c)
		} else {
			log.Error().Msgf("Command for check '%s' does not exists", c.Name)
		}
	}

	return s
}

func visit(files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) != ".json" {
			return nil
		}

		if err != nil {
			log.Panic().Msg(err.Error())
		}

		*files = append(*files, path)
		return nil
	}
}

func walk(hs map[string]handlers.Handler) error {
	var files []string

	root := viper.GetString("handlers-path")
	abs, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	log.Debug().Msgf("Reading in handlers from '%s'", abs)
	wErr := filepath.Walk(root, visit(&files))
	if wErr != nil {
		return wErr
	}

	if len(files) == 0 {
		log.Info().Msg("No handlers defined")
	}

	for _, file := range files {
		abs, err := filepath.Abs(file)
		if err != nil {
			return err
		}

		log.Debug().Msgf("Reading handler file '%s'", abs)
		f, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}

		m := make(map[string]interface{})
		if err := json.Unmarshal(f, &m); err != nil {
			return err
		}

		keys := make([]string, len(m))
		i := 0
		for s := range m {
			keys[i] = s
			i++
		}

		for _, k := range keys {
			handlerType := m[k].(map[string]interface{})["type"].(string)
			log.Info().Msgf("Adding handler '%s' as type '%s'", k, handlerType)
			newHandler := factories.Handler(handlerType, m[k].(map[string]interface{}))
			hs[k] = newHandler
		}
	}
	return nil
}
