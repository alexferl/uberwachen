package handlers

import (
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/alexferl/uberwachen/util"
)

// CheckLoad is used to load a check from a file,
// and then it gets converted to a Check
type CheckLoad struct {
	Name         string   `json:"name"`
	Command      string   `json:"command"`
	Interval     int      `json:"interval"`
	MaxAttempts  int      `json:"max_attempts" bson:"max_attempts"`
	HandlerNames []string `json:"handlers" bson:"handlers"`
	Renotify     bool     `json:"renotify"`
}

// Check represents a check
type Check struct {
	*CheckLoad     `bson:",inline"`
	Attempts       int        `json:"attempts"`
	Duration       float64    `json:"duration"`
	ExecutedAt     time.Time  `json:"executed_at" bson:"executed_at"`
	History        []int      `json:"history"`
	IssuedAt       time.Time  `json:"issued_at" bson:"issued_at"`
	PreviousOutput string     `json:"previous_output" bson:"previous_output"`
	Output         string     `json:"output"`
	Status         int        `json:"status"`
	Handlers       []*Handler `json:"-" bson:"-"`
}

// NewCheck creates a new Check
func NewCheck() *Check {
	return &Check{
		CheckLoad: &CheckLoad{},
	}
}

// RunCheck runs a Check and fires an Event with the result for processing
func RunCheck(c *Check) {
	cmdPath := util.GetCmdPath(c.Command)
	_, cmdArgs := util.SplitCmd(c.Command)

	c.IssuedAt = time.Now().UTC()
	code := 0
	output, err := exec.Command(cmdPath, cmdArgs...).CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.Sys().(syscall.WaitStatus).ExitStatus()
		} else {
			log.Error().Msgf("Error running check '%s': %v", c.Command, err)
			return
		}
	}

	c.Status = code
	c.Output = strings.TrimSuffix(string(output), "\n")
	c.ExecutedAt = time.Now().UTC()
	c.Duration = c.ExecutedAt.Sub(c.IssuedAt).Seconds()

	c.History = append([]int{c.Status}, c.History...) // prepend

	if len(c.History) > 10 {
		c.History = c.History[:10] // Don't need more than 10 statuses
	}

	log.Debug().Msgf("Ran check '%s': exit code: '%d' duration: '%.3f' output: '%s'",
		c.Name, c.Status, c.Duration, c.Output)

	event := NewEvent(c)
	event.Process()
}
