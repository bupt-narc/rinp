package scheduler

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/bupt-narc/rinp/pkg/version"
	"github.com/pkg/errors"
	"github.com/rueian/rueidis"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	redisProxy  rueidis.Client
	redisClient rueidis.Client
	opt         *Option
)

func runCli(cmd *cobra.Command, args []string) error {
	var err error
	opt, err = NewOption().
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

	logrus.Infof("rinp-scheduler version %s", version.Version)

	// Performance profiling
	if opt.EnablePProf {
		logrus.Info("performance profiling enabled, http server listing at :8080/debug/pprof")
		go http.ListenAndServe(":8080", nil)
	}

	redisProxy, err = rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{opt.Redis},
		SelectDB:    1,
	})
	if err != nil {
		return err
	}
	redisClient, err = rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{opt.Redis},
		SelectDB:    0,
	})
	if err != nil {
		return err
	}

	conn, err := NewSchedulerConn(opt.Port)
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

	conn.Run(ctx)

	return nil
}
