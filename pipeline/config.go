package pipeline

import (
	"time"

	"github.com/spf13/viper"
)

const (
	ConfigTimeout  = "timeout"
	ConfigLogLevel = "logLevel"
)

var config Config

func init() {
	v := viper.New()
	config = v

	v.SetDefault(ConfigTimeout, "1m") // 1 minute
	v.SetDefault(ConfigLogLevel, "debug")
}

// Config is the configuration for system
type Config interface {
	Get(key string) interface{}
	GetBool(key string) bool
	GetFloat64(key string) float64
	GetInt(key string) int
	GetString(key string) string
	GetStringMap(key string) map[string]interface{}
	GetStringMapString(key string) map[string]string
	GetStringSlice(key string) []string
	GetTime(key string) time.Time
	GetDuration(key string) time.Duration
	IsSet(key string) bool

	UnmarshalKey(key string, rawVal interface{}) error
}

// GetConfig retrieves the configuration
func GetConfig() Config {
	return config
}
