package sidecar

import (
	"fmt"

	"github.com/bupt-narc/rinp/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	flagPort      = "port"
	flagPortShort = "p"
	flagPortUsage = "UDP listening port"

	flagLogLevel      = "log-level"
	flagLogLevelUsage = "Log level"

	flagServerVirtualIP      = "server-virtual-ip"
	flagServerVirtualIPShort = "s"
	flagServerVirtualIPUsage = "The virtual IP of this sidecar"

	flagClientCIDRs      = "client-virtual-cidrs"
	flagClientCIDRsShort = "c"
	flagClientCIDRsUsage = "The CIDRs of clients' virtual IP addresses"

	flagEnablePProf      = "enable-pprof"
	flagEnablePProfUsage = "Enable performance profiling at :8080/debug/pprof"

	flagPublicIP      = "public-ip"
	flagPublicIPUsage = "Public IP address of this sidecar"

	flagRedis      = "redis"
	flagRedisShort = "r"
	flagRedisUsage = "Redis address"
)

const (
	envStrPort     = "PORT"
	envStrLogLevel = "LOG_LEVEL"
	// TODO
)

var (
	defaultPort            = 32000
	defaultLogLevel        = "info"
	defaultServerVirtualIP = ""
	defaultClientCIDRs     = []string{}
	defaultEnablePProf     = false
	defaultPublicIP        = ""
	defaultRedis           = "localhost:6379"
)

const (
	cmdLongHelp = `rinp-sidecar is the service sidecar for RINP (RINP Is Not a Proxy).

All command-line options can be specified as environment variables, which are defined by the command-line option, 
capitalized, with all -’s replaced with _’s.

For example, $LOG_LEVEL can be used in place of --log-level

Options have a priority like this: cli-flags > env > default-values`
)

func NewCommand() *cobra.Command {
	c := &cobra.Command{
		Use:          "sidecar",
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
		Short: "show rinp-sidecar version and exit",
		Run: func(cmd *cobra.Command, args []string) {
			//nolint:forbidigo // print version
			fmt.Println(version.Version)
		},
	}
	return c
}

func addFlags(f *pflag.FlagSet) {
	f.IntP(flagPort, flagPortShort, defaultPort, flagPortUsage)
	f.String(flagLogLevel, defaultLogLevel, flagLogLevelUsage)
	f.StringP(flagServerVirtualIP, flagServerVirtualIPShort, defaultServerVirtualIP, flagServerVirtualIPUsage)
	f.StringArrayP(flagClientCIDRs, flagClientCIDRsShort, defaultClientCIDRs, flagClientCIDRsUsage)
	f.Bool(flagEnablePProf, defaultEnablePProf, flagEnablePProfUsage)
	f.String(flagPublicIP, defaultPublicIP, flagPublicIPUsage)
	f.StringP(flagRedis, flagRedisShort, defaultRedis, flagRedisUsage)
}
