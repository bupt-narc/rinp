package proxy

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bupt-narc/rinp/pkg/version"
	"github.com/pkg/errors"
	"github.com/rueian/rueidis"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	redisProxy   rueidis.Client
	redisClient  rueidis.Client
	redisSidecar rueidis.Client
	opt          *Option
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

	logrus.Infof("rinp-proxy version %s", version.Version)

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
	redisSidecar, err = rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{opt.Redis},
		SelectDB:    2,
	})
	if err != nil {
		return err
	}

	conn, err := NewProxyConn(opt.Port)
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

	go selfHeartbeat(opt.Name, opt.PublicIP, opt.Port)

	conn.Run(ctx)

	return nil
}

func selfHeartbeat(name, ip string, port int) {
	do := func(ctx context.Context) {
		err := redisProxy.Do(ctx, redisProxy.B().Set().Key(name).Value(fmt.Sprintf("%s:%d", ip, port)).ExSeconds(2).Build()).Error()
		if err != nil {
			logrus.Errorf("redis cannot set %s:%s", name, ip)
			return
		}
	}

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)

		defer cancel()
		defer time.Sleep(1 * time.Second)

		do(ctx)
	}
}
