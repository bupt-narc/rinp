package proxy

import (
	"fmt"

	"github.com/bupt-narc/rinp/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	flagLogLevel      = "log-level"
	flagLogLevelUsage = "Log level"

	flagPort      = "port"
	flagPortShort = "p"
	flagPortUsage = "UDP listening port"

	flagEnablePProf      = "enable-pprof"
	flagEnablePProfUsage = "Enable performance profiling at :8080/debug/pprof"

	flagName      = "name"
	flagNameShort = "n"
	flagNameUsage = "Name of this proxy"

	flagPublicIP      = "public-ip"
	flagPublicIPUsage = "Public IP address of this proxy"

	flagRedis      = "redis"
	flagRedisShort = "r"
	flagRedisUsage = "Redis address"
)

const (
	envStrLogLevel = "LOG_LEVEL"
	// TODO
)

var (
	defaultLogLevel    = "info"
	defaultEnablePProf = false
	defaultPort        = 5114
	defaultName        = ""
	defaultPublicIP    = ""
	defaultRedis       = "localhost:6379"
)

const (
	cmdLongHelp = `rinp-proxy is the proxy node for RINP (RINP Is Not a Proxy).

All command-line options can be specified as environment variables, which are defined by the command-line option, 
capitalized, with all -’s replaced with _’s.

For example, $LOG_LEVEL can be used in place of --log-level

Options have a priority like this: cli-flags > env > default-values`
)

func NewCommand() *cobra.Command {
	c := &cobra.Command{
		Use:          "proxy",
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
		Short: "show rinp-proxy version and exit",
		Run: func(cmd *cobra.Command, args []string) {
			//nolint:forbidigo // print version
			fmt.Println(version.Version)
		},
	}
	return c
}

func addFlags(f *pflag.FlagSet) {
	f.String(flagLogLevel, defaultLogLevel, flagLogLevelUsage)
	f.IntP(flagPort, flagPortShort, defaultPort, flagPortUsage)
	f.Bool(flagEnablePProf, defaultEnablePProf, flagEnablePProfUsage)
	f.StringP(flagName, flagNameShort, defaultName, flagNameUsage)
	f.String(flagPublicIP, defaultPublicIP, flagPublicIPUsage)
	f.StringP(flagRedis, flagRedisShort, defaultRedis, flagRedisUsage)
}
