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

	flagServerAddress      = "server-address"
	flagServerAddressShort = "s"

	flagClientAddress      = "client-address"
	flagClientAddressShort = "c"
)

const (
	envStrPort     = "PORT"
	envStrLogLevel = "LOG_LEVEL"
)

const (
	defaultPort     = 8080
	defaultLogLevel = "trace"
)

const (
	cmdLongHelp = `rinp-client is the client for RINP (RINP Is Not a Proxy).

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
	f.StringP(flagServerAddress, flagServerAddressShort, "", "")
	f.StringP(flagClientAddress, flagClientAddressShort, "", "")
}
