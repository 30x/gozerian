package pipeline

import (
	"time"
	"github.com/Sirupsen/logrus"
	"os"
)

type Key interface{}
type Value interface{}

type Config interface {
	Get(k Key) Value
	Timeout() time.Duration
	Logger() Logger
}

const (
	ConfigTimeout = "timeout"
	ConfigLogLevel = "logLevel"
)

func NewDefaultConfig() Config {
	values := make(map[Key]Value)

	values[ConfigTimeout] = 60000
	values[ConfigLogLevel] = "debug"

	return &config{values}
}

type config struct {
	values map[Key]Value
}

func (self *config) Get(k Key) Value {
	return self.values[k]
}

func (self *config) Timeout() time.Duration {
	timeout := self.values[ConfigTimeout].(int)
	return time.Duration(timeout) * time.Millisecond
}

// todo: probably do this elsewhere?
func (self *config) Logger() Logger {
	level, error := logrus.ParseLevel(self.values[ConfigLogLevel].(string))
	if error != nil {
		panic(error) // todo: handle error
	}
	logrus.SetLevel(level)
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&logrus.TextFormatter{})
	return logrus.New()
}
