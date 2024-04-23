package client

import (
	"fmt"
	"net"
	"os"

	"github.com/bupt-narc/rinp/pkg/overlay"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

type Option struct {
	LogLevel         string
	ProxyAddress     string
	ClientVirtualIP  net.IP
	ServerCIDRs      []*net.IPNet
	EnablePProf      bool
	AuthBaseURL      string
	SchedulerAddress string
}

func NewOption() *Option {
	return &Option{}
}

func (o *Option) WithDefaults() *Option {
	o.LogLevel = defaultLogLevel
	o.ProxyAddress = defaultProxyAddress
	o.ClientVirtualIP = net.ParseIP(defaultClientVirtualIP)
	o.ServerCIDRs, _ = overlay.StringToCIDRs(defaultServerCIDRs)
	o.EnablePProf = defaultEnablePProf
	o.AuthBaseURL = defaultAuthBaseURL
	o.SchedulerAddress = defaultSchedulerAddress
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
	if v, err := flags.GetString(flagProxyAddress); err == nil && flags.Changed(flagProxyAddress) {
		o.ProxyAddress = v
	}
	if v, err := flags.GetString(flagClientVirtualIP); err == nil && flags.Changed(flagClientVirtualIP) {
		o.ClientVirtualIP = net.ParseIP(v)
	}
	if v, err := flags.GetStringArray(flagServerCIDRs); err == nil && flags.Changed(flagServerCIDRs) {
		cidrs, err := overlay.StringToCIDRs(v)
		o.ServerCIDRs = cidrs
		if err != nil {
			logrus.Errorln(err)
			o.ServerCIDRs = nil
		}
	}
	if v, err := flags.GetString(flagSchedulerAddress); err == nil && flags.Changed(flagSchedulerAddress) {
		o.SchedulerAddress = v
	}
	return o
}

func (o *Option) WithPreFlags(flags *pflag.FlagSet) *Option {
	if v, err := flags.GetBool(flagEnablePProf); err == nil && flags.Changed(flagEnablePProf) {
		o.EnablePProf = v
	}
	if v, err := flags.GetString(flagAuthBaseURL); err == nil && flags.Changed(flagAuthBaseURL) {
		o.AuthBaseURL = v
	}
	return o
}

func (o *Option) WithNetwork() *Option {
	// TODO input email and password
	err := setInfoByDefault(o)
	if err != nil {
		logrus.Errorln(err)
		return o
	}
	return o
}

func (o *Option) Validate() (*Option, error) {
	_, err := logrus.ParseLevel(o.LogLevel)
	if err != nil {
		return nil, err
	}
	if o.ProxyAddress == "" {
		return nil, fmt.Errorf("%s is not valid, %v", flagProxyAddress, o.ProxyAddress)
	}
	if o.ClientVirtualIP == nil {
		return nil, fmt.Errorf("%s is not valid, %v", flagClientVirtualIP, o.ClientVirtualIP)
	}
	if o.ServerCIDRs == nil {
		return nil, fmt.Errorf("%s is not valid, %v", flagClientVirtualIP, o.ServerCIDRs)
	}
	return o, nil
}
