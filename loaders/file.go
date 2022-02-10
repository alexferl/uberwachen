package loaders

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/alexferl/uberwachen/handlers"
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

func (fl *FileLoader) Load() error {
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

		err = fl.parseChecks(m)
		if err != nil {
			return err
		}
	}

	return nil
}

func checkInSlice(check *handlers.Check, list []*handlers.Check) bool {
	for _, c := range list {
		if c.Name == check.Name {
			return true
		}
	}
	return false
}

func (fl *FileLoader) parseChecks(m map[string]interface{}) error {
	keys := make([]string, len(m))

	i := 0
	for s := range m {
		keys[i] = s
		i++
	}

	for _, key := range keys {
		if check, ok := m[key].(map[string]interface{}); ok {
			c := &handlers.Check{}

			c.Name = key

			if m[key].(map[string]interface{})["command"] == nil {
				msg := fmt.Sprintf("Command is required")
				return errors.New(msg)
			} else {
				c.Command = m[key].(map[string]interface{})["command"].(string)
			}

			if m[key].(map[string]interface{})["interval"] == nil {
				msg := fmt.Sprintf("Interval is required")
				return errors.New(msg)
			} else {
				c.Interval = int(m[key].(map[string]interface{})["interval"].(float64))
			}

			if m[key].(map[string]interface{})["max_attempts"] == nil {
				c.MaxAttempts = 1
			} else {
				c.MaxAttempts = int(m[key].(map[string]interface{})["max_attempts"].(float64))
			}

			if m[key].(map[string]interface{})["renotify"] == nil {
				c.Renotify = false
			} else {
				c.Renotify = m[key].(map[string]interface{})["renotify"].(bool)
			}

			if m[key].(map[string]interface{})["handlers"] != nil {
				for _, handler := range m[key].(map[string]interface{})["handlers"].([]interface{}) {
					h := fl.addHandler(handler.(string))
					if h.Handler != nil {
						c.Handlers = append(c.Handlers, h)
					}
				}
			}

			if !checkInSlice(c, fl.Checks) {
				fl.Checks = append(fl.Checks, c)
			}

		} else {
			msg := fmt.Sprintf("check not a map[string]interface{}: %v\n", check)
			err := errors.New(msg)
			return err
		}
	}
	return nil
}

func (fl *FileLoader) addHandler(handler string) handlers.Handler {
	v := viper.Get("handlers").(map[string]handlers.Handler)
	return v[handler]
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