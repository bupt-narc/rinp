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
	flagPortUsage = "Listening port"

	flagLogLevel      = "log-level"
	flagLogLevelUsage = "Log level"
)

const (
	envStrPort     = "PORT"
	envStrLogLevel = "LOG_LEVEL"
)

const (
	defaultPort     = 32000
	defaultLogLevel = "info"
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
	f.String(flagLogLevel, defaultLogLevel, flagLogLevelUsage)
	f.IntP(flagPort, flagPortShort, defaultPort, flagPortUsage)
}
