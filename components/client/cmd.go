package client

import (
	"fmt"

	"github.com/bupt-narc/rinp/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	flagLogLevel      = "log-level"
	flagLogLevelUsage = "Log level"

	flagProxyAddress      = "proxy-address"
	flagProxyAddressShort = "p"
	flagProxyAddressUsage = "The address of UDP proxy"

	flagClientVirtualIP      = "client-virtual-ip"
	flagClientVirtualIPShort = "c"
	flagClientVirtualIPUsage = "The virtual IP of this client"

	flagServerCIDRs      = "server-virtual-cidrs"
	flagServerCIDRsShort = "s"
	flagServerCIDRsUsage = "The CIDRs of servers' virtual IP addresses"

	flagEnablePProf      = "enable-pprof"
	flagEnablePProfUsage = "Enable performance profiling at :8080/debug/pprof"

	flagAuthBaseURL      = "auth-base-url"
	flagAuthBaseURLUsage = "The base URL of auth service"

	flagSchedulerAddress      = "scheduler"
	flagSchedulerAddressUsage = "The address of scheduler"
)

const (
	envStrLogLevel = "LOG_LEVEL"
	// TODO
)

var (
	defaultLogLevel         = "info"
	defaultProxyAddress     = "proxy1:5114"
	defaultClientVirtualIP  = "7.1.2.3"
	defaultServerCIDRs      = []string{"11.22.33.0/24"}
	defaultEnablePProf      = false
	defaultAuthBaseURL      = "http://auth:8090"
	defaultSchedulerAddress = "http://11.22.33.55"
)

const (
	cmdLongHelp = `rinp-client is the user client for RINP (RINP Is Not a Proxy).

All command-line options can be specified as environment variables, which are defined by the command-line option, 
capitalized, with all -’s replaced with _’s.

For example, $LOG_LEVEL can be used in place of --log-level

Options have a priority like this: cli-flags > env > default-values`
)

func NewCommand() *cobra.Command {
	c := &cobra.Command{
		Use:          "client",
		Long:         cmdLongHelp,
		SilenceUsage: true,
		RunE:         runCli,
	}
	addFlags(c.Flags())
	c.AddCommand(NewVersionCommand())
	return c
}

func NewVersionCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "version",
		Short: "show rinp-client version and exit",
		Run: func(cmd *cobra.Command, args []string) {
			//nolint:forbidigo // print version
			fmt.Println(version.Version)
		},
	}
	return c
}

func addFlags(f *pflag.FlagSet) {
	f.String(flagLogLevel, defaultLogLevel, flagLogLevelUsage)
	f.StringP(flagProxyAddress, flagProxyAddressShort, defaultProxyAddress, flagProxyAddressUsage)
	f.StringP(flagClientVirtualIP, flagClientVirtualIPShort, defaultClientVirtualIP, flagClientVirtualIPUsage)
	f.StringArrayP(flagServerCIDRs, flagServerCIDRsShort, defaultServerCIDRs, flagServerCIDRsUsage)
	f.Bool(flagEnablePProf, defaultEnablePProf, flagEnablePProfUsage)
	f.String(flagAuthBaseURL, defaultAuthBaseURL, flagAuthBaseURLUsage)
	f.String(flagSchedulerAddress, defaultSchedulerAddress, flagSchedulerAddressUsage)
}
