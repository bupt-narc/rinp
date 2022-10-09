package client

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/elazarl/goproxy"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func runCli(cmd *cobra.Command, args []string) error {
	opt, err := NewOption().
		WithDefaults().
		WithEnvVariables().
		WithCliFlags(cmd.Flags()).
		Validate()
	if err != nil {
		return errors.Wrap(err, "error when paring flags")
	}

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true

	logrus.Infof("listening on port %d", opt.Port)

	return http.ListenAndServe(fmt.Sprintf(":%d", opt.Port), proxy)
}
