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
)

const (
	envStrLogLevel = "LOG_LEVEL"
	// TODO
)

var (
	defaultLogLevel        = "info"
	defaultProxyAddress    = ""
	defaultClientVirtualIP = ""
	defualtServerCIDRs     = []string{}
	defaultEnablePProf     = false
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
	f.StringArrayP(flagServerCIDRs, flagServerCIDRsShort, defualtServerCIDRs, flagServerCIDRsUsage)
	f.Bool(flagEnablePProf, defaultEnablePProf, flagEnablePProfUsage)
}
