package client

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

type Option struct {
	LogLevel      string
	ServerAddress string
	ClientAddress string // change name
}

func NewOption() *Option {
	return &Option{}
}

func (o *Option) WithDefaults() *Option {
	o.LogLevel = defaultLogLevel
	return o
}

func (o *Option) WithEnvVariables() *Option {
	if v, ok := os.LookupEnv(envStrLogLevel); ok && v != "" {
		o.LogLevel = v
	}
	return o
}

func (o *Option) WithCliFlags(flags *pflag.FlagSet) *Option {
	if v, err := flags.GetString(flagLogLevel); err == nil && flags.Changed(flagLogLevel) {
		o.LogLevel = v
	}
	if v, err := flags.GetString(flagServerAddress); err == nil && flags.Changed(flagServerAddress) {
		o.ServerAddress = v
	}
	if v, err := flags.GetString(flagClientAddress); err == nil && flags.Changed(flagClientAddress) {
		o.ClientAddress = v
	}
	return o
}

func (o *Option) Validate() (*Option, error) {
	_, err := logrus.ParseLevel(o.LogLevel)
	if err != nil {
		return nil, err
	}
	if o.ServerAddress == "" {
		return nil, fmt.Errorf("%s must not be empty", flagServerAddress)
	}
	if o.ClientAddress == "" {
		return nil, fmt.Errorf("%s must not be empty", flagClientAddress)
	}
	return o, nil
}
