package proxy

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

type Option struct {
	LogLevel    string
	Port        int
	EnablePProf bool
	Name        string
	PublicIP    string
	Redis       string
}

func NewOption() *Option {
	return &Option{}
}

func (o *Option) WithDefaults() *Option {
	o.LogLevel = defaultLogLevel
	o.Port = defaultPort
	o.Redis = defaultRedis
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
	if v, err := flags.GetInt(flagPort); err == nil && flags.Changed(flagPort) {
		o.Port = v
	}
	if v, err := flags.GetBool(flagEnablePProf); err == nil && flags.Changed(flagEnablePProf) {
		o.EnablePProf = v
	}
	if v, err := flags.GetString(flagName); err == nil && flags.Changed(flagName) {
		o.Name = v
	}
	if v, err := flags.GetString(flagPublicIP); err == nil && flags.Changed(flagPublicIP) {
		o.PublicIP = v
	}
	if v, err := flags.GetString(flagRedis); err == nil && flags.Changed(flagRedis) {
		o.Redis = v
	}
	return o
}

func (o *Option) Validate() (*Option, error) {
	_, err := logrus.ParseLevel(o.LogLevel)
	if err != nil {
		return nil, err
	}
	if o.Port < 0 || o.Port > 65535 {
		return nil, fmt.Errorf("invalid port number %d", o.Port)
	}
	if o.Name == "" {
		return nil, fmt.Errorf("name should not be empty")
	}
	if o.PublicIP == "" {
		return nil, fmt.Errorf("public ip should not be empty")
	}
	if o.Redis == "" {
		return nil, fmt.Errorf("redis should not be empty")
	}
	return o, nil
}
