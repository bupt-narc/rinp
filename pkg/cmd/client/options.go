package client

import (
	"fmt"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

type Option struct {
	LogLevel string
	Port     int
}

func NewOption() *Option {
	return &Option{}
}

func (o *Option) WithDefaults() *Option {
	o.Port = defaultPort
	o.LogLevel = defaultLogLevel
	return o
}

func (o *Option) WithEnvVariables() *Option {
	if v, ok := os.LookupEnv(envStrLogLevel); ok && v != "" {
		o.LogLevel = v
	}
	if v, ok := os.LookupEnv(envStrPort); ok && v != "" {
		o.Port, _ = strconv.Atoi(v)
	}

	return o
}

func (o *Option) WithCliFlags(flags *pflag.FlagSet) *Option {
	if v, err := flags.GetString(flagLogLevel); err == nil && flags.Changed(flagLogLevel) {
		o.LogLevel = v
	}
	if v, err := flags.GetInt(flagPort); err == nil && flags.Changed(flagPort) {
		o.Port = v
	}
	return o
}

func (o *Option) Validate() (*Option, error) {
	_, err := logrus.ParseLevel(o.LogLevel)
	if err != nil {
		return nil, err
	}
	if o.Port <= 0 {
		return nil, fmt.Errorf("%s must be greater than 0", flagPort)
	}

	return o, nil
}
