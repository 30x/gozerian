package pipeline

import (
	"os"
	"time"

	"github.com/Sirupsen/logrus"
)

// A Key for a Config
type Key interface{}

// A Value for a Config
type Value interface{}

// Config is the configuration for system
type Config interface {
	Get(k Key) Value
	Timeout() time.Duration
	Log() Logger
}

var conf Config

// GetConfig retrieves the configuration
func GetConfig() Config {
	if conf == nil {
		conf = NewDefaultConfig()
	}
	return conf
}

// Well-known Config keys
const (
	ConfigTimeout  = "timeout"
	ConfigLogLevel = "logLevel"
)

// NewDefaultConfig creates a config
func NewDefaultConfig() Config {
	values := make(map[Key]Value)

	values[ConfigTimeout] = 60000
	values[ConfigLogLevel] = "debug"

	return &config{values}
}

type config struct {
	values map[Key]Value
}

func (c *config) Get(k Key) Value {
	return c.values[k]
}

func (c *config) Timeout() time.Duration {
	timeout := c.values[ConfigTimeout].(int)
	return time.Duration(timeout) * time.Millisecond
}

// todo: probably do this elsewhere?
func (c *config) Log() Logger {
	level, error := logrus.ParseLevel(c.values[ConfigLogLevel].(string))
	if error != nil {
		panic(error) // todo: handle error
	}
	logrus.SetLevel(level)
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&logrus.TextFormatter{})
	return logrus.New()
}
