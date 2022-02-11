package loaders

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"

	"github.com/alexferl/uberwachen/handlers"
	"github.com/alexferl/uberwachen/registries"
)

type FileLoader struct {
	Path   string
	Checks []*handlers.Check
}

func NewFileLoader(path string) Loader {
	return Loader(&FileLoader{
		Path: path,
	})
}

func (fl *FileLoader) Load(registry *registries.Handlers) error {
	abs, err := filepath.Abs(fl.Path)
	if err != nil {
		return err
	}

	log.Debug().Msgf("Reading in checks from '%s'", abs)

	_, pErr := fl.pathExists()
	if pErr != nil {
		return pErr
	}

	files, err := fl.walk()
	if err != nil {
		return err
	}

	if len(files) == 0 {
		log.Info().Msg("No checks defined")
	}

	for _, file := range files {
		abs, err := filepath.Abs(file)
		if err != nil {
			return err
		}

		log.Debug().Msgf("Reading check file '%s'", abs)

		f, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}

		var m map[string]interface{}
		if err := json.Unmarshal(f, &m); err != nil {
			return err
		}

		err = fl.parseChecks(m, registry)
		if err != nil {
			return err
		}
	}

	return nil
}

func checkInSlice(check *handlers.Check, slice []*handlers.Check) bool {
	for _, c := range slice {
		if c.Name == check.Name {
			return true
		}
	}
	return false
}

func (fl *FileLoader) parseChecks(m map[string]interface{}, registry *registries.Handlers) error {
	keys := make([]string, len(m))

	i := 0
	for s := range m {
		keys[i] = s
		i++
	}

	for _, key := range keys {
		if _, ok := m[key]; ok {
			b, err := json.Marshal(m[key])
			if err != nil {
				return err
			}

			var ch handlers.CheckLoad
			if err := json.Unmarshal(b, &ch); err != nil {
				return err
			}

			c := handlers.NewCheck()
			c.Name = key

			if ch.Command == "" {
				msg := fmt.Sprintf("Command is required")
				return errors.New(msg)
			} else {
				c.Command = ch.Command
			}

			if ch.Interval == 0 {
				msg := fmt.Sprintf("Interval is required")
				return errors.New(msg)
			} else {
				c.Interval = ch.Interval
			}

			if ch.MaxAttempts == 0 {
				c.MaxAttempts = 1
			} else {
				c.MaxAttempts = ch.MaxAttempts
			}

			c.Renotify = ch.Renotify
			c.HandlerNames = ch.HandlerNames

			for _, handler := range ch.HandlerNames {
				h, _ := registry.Get(handler)
				if h.Handler != nil {
					c.Handlers = append(c.Handlers, h)
				}
			}

			if !checkInSlice(c, fl.Checks) {
				fl.Checks = append(fl.Checks, c)
			}

		}
	}
	return nil
}

// isDirectory check if the path is a directory
func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), nil
}

func (fl *FileLoader) walk() ([]string, error) {
	var fileList []string
	err := filepath.Walk(fl.Path, func(path string, f os.FileInfo, err error) error {
		dir, dirErr := isDirectory(path)
		if err != nil {
			return dirErr
		}

		if !dir {
			ext := filepath.Ext(path)
			if ext == ".json" {
				fileList = append(fileList, path)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return fileList, nil
}

// pathExists check if a specific path exists
func (fl *FileLoader) pathExists() (bool, error) {
	_, err := os.Stat(fl.Path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return true, err
}
