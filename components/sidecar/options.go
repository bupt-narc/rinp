package sidecar

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/bupt-narc/rinp/pkg/overlay"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

type Option struct {
	LogLevel        string
	Port            int
	ServerVirtualIP net.IP
	ClientCIDRs     []*net.IPNet
	EnablePProf     bool
	PrivateIP       string
	Redis           string
}

func NewOption() *Option {
	return &Option{}
}

func (o *Option) WithDefaults() *Option {
	o.Port = defaultPort
	o.LogLevel = defaultLogLevel
	o.ServerVirtualIP = net.IP(defaultServerVirtualIP)
	o.ClientCIDRs, _ = overlay.StringToCIDRs(defaultClientCIDRs)
	o.EnablePProf = defaultEnablePProf
	o.Redis = defaultRedis
	return o
}

func (o *Option) WithEnvVariables() *Option {
	if v, ok := os.LookupEnv(envStrLogLevel); ok && v != "" {
		o.LogLevel = v
	}
	if v, ok := os.LookupEnv(envStrPort); ok && v != "" {
		o.Port, _ = strconv.Atoi(v)
	}
	// TODO
	return o
}

func (o *Option) WithCliFlags(flags *pflag.FlagSet) *Option {
	if v, err := flags.GetInt(flagPort); err == nil && flags.Changed(flagPort) {
		o.Port = v
	}
	if v, err := flags.GetString(flagLogLevel); err == nil && flags.Changed(flagLogLevel) {
		o.LogLevel = v
	}
	if v, err := flags.GetString(flagServerVirtualIP); err == nil && flags.Changed(flagServerVirtualIP) {
		o.ServerVirtualIP = net.ParseIP(v)
	}
	if v, err := flags.GetStringArray(flagClientCIDRs); err == nil && flags.Changed(flagClientCIDRs) {
		cidrs, err := overlay.StringToCIDRs(v)
		o.ClientCIDRs = cidrs
		if err != nil {
			logrus.Errorln(err)
			o.ClientCIDRs = nil
		}
	}
	if v, err := flags.GetBool(flagEnablePProf); err == nil && flags.Changed(flagEnablePProf) {
		o.EnablePProf = v
	}
	if v, err := flags.GetString(flagPrivateIP); err == nil && flags.Changed(flagPrivateIP) {
		o.PrivateIP = v
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
	// FIXME
	if o.Port <= 0 {
		return nil, fmt.Errorf("%s must be greater than 0", flagPort)
	}
	if o.ServerVirtualIP == nil {
		return nil, fmt.Errorf("%s is not valid", flagServerVirtualIP)
	}
	if o.ClientCIDRs == nil {
		return nil, fmt.Errorf("%s is not valid", flagClientCIDRs)
	}
	if o.PrivateIP == "" {
		return nil, fmt.Errorf("public ip should not be empty")
	}
	if o.Redis == "" {
		return nil, fmt.Errorf("redis should not be empty")
	}
	return o, nil
}
