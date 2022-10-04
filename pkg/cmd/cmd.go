package cmd

import (
	"fmt"

	"github.com/bupt-narc/mtda/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	flagPort      = "port"
	flagPortShort = "p"
)

const (
	defaultPort = 8080
)

const (
	cmdLongHelp = `this is looooong help`
)

func NewCommand() *cobra.Command {
	c := &cobra.Command{
		Use:  "mtda",
		Long: cmdLongHelp,
		RunE: runCli,
	}
	addFlags(c.Flags())
	c.AddCommand(NewVersionCommand())
	return c
}

func NewVersionCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "version",
		Short: "show kube-trigger version and exit",
		Run: func(cmd *cobra.Command, args []string) {
			//nolint:forbidigo // print version
			fmt.Println(version.Version)
		},
	}
	return c
}

func addFlags(f *pflag.FlagSet) {
	f.IntP(flagPort, flagPortShort, defaultPort, "Listen port")
}
