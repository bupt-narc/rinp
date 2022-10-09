package client

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/pflag"
)

type Option struct {
	Port int
}

func NewOption() *Option {
	return &Option{}
}

func (o *Option) WithDefaults() *Option {
	o.Port = defaultPort

	return o
}

func (o *Option) WithEnvVariables() *Option {
	if v, ok := os.LookupEnv(envStrPort); ok && v != "" {
		o.Port, _ = strconv.Atoi(v)
	}

	return o
}

func (o *Option) WithCliFlags(flags *pflag.FlagSet) *Option {
	if v, err := flags.GetInt(flagPort); err == nil && flags.Changed(flagPort) {
		o.Port = v
	}

	return o
}

func (o *Option) Validate() (*Option, error) {
	if o.Port <= 0 {
		return nil, fmt.Errorf("%s must be greater than 0", flagPort)
	}

	return o, nil
}
