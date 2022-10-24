package sidecar

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bupt-narc/rinp/pkg/overlay"
	"github.com/bupt-narc/rinp/pkg/version"
	"github.com/pkg/errors"
	"github.com/rueian/rueidis"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	redisSidecar rueidis.Client
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

	// Set log level. No need to check error, we validated it previously.
	level, _ := logrus.ParseLevel(opt.LogLevel)
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "15:04:05.000",
		FullTimestamp:   true,
	})

	logrus.Infof("rinp-sidecar version %s", version.Version)

	// Performance profiling
	if opt.EnablePProf {
		logrus.Info("performance profiling enabled, http server listing at :8080/debug/pprof")
		go http.ListenAndServe(":8080", nil)
	}

	redisSidecar, err = rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{opt.Redis},
		SelectDB:    2,
	})
	if err != nil {
		return err
	}

	conn, err := overlay.NewServerConn(
		"tunsidecar0",
		opt.ServerVirtualIP,
		opt.Port,
		opt.ClientCIDRs,
	)
	if err != nil {
		return errors.Wrap(err, "cannot create connection")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		// Listen to termination signals.
		sigterm := make(chan os.Signal, 1)
		signal.Notify(sigterm, syscall.SIGTERM)
		signal.Notify(sigterm, syscall.SIGINT)
		<-sigterm
		cancel()
	}()

	go selfHeartbeat(opt.ServerVirtualIP.String(), opt.PublicIP, opt.Port)

	conn.Run(ctx)

	return nil
}

func selfHeartbeat(vip, publicIP string, port int) {
	do := func(ctx context.Context) {
		err := redisSidecar.Do(ctx, redisSidecar.B().Set().Key(vip).Value(fmt.Sprintf("%s:%d", publicIP, port)).ExSeconds(10).Build()).Error()
		if err != nil {
			logrus.Errorf("redis cannot set %s:%s", vip, publicIP)
			return
		}
	}

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		defer cancel()
		defer time.Sleep(5 * time.Second)

		do(ctx)
	}
}
