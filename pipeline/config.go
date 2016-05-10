package pipeline

import "time"

type Key interface{}
type Value interface{}

type Config interface {
	Get(k Key) Value
}

func NewDefaultConfig() Config {
	values := make(map[Key]Value)

	values["timeout"] = time.Minute

	return &defaultConfig{values}
}

type defaultConfig struct {
	values map[Key]Value
}

func (self defaultConfig) Get(k Key) Value {
	return self.values[k]
}
