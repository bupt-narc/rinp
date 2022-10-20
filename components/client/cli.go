package client

import (
	"bufio"
	"context"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/bupt-narc/rinp/pkg/overlay"
	"github.com/bupt-narc/rinp/pkg/version"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func runCli(cmd *cobra.Command, args []string) error {
	opt, err := NewOption().
		WithDefaults().
		WithNetwork().
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

	logrus.Infof("rinp-client version %s", version.Version)

	// Performance profiling
	if opt.EnablePProf {
		logrus.Info("performance profiling enabled, http server listing at :8080/debug/pprof")
		go http.ListenAndServe(":8080", nil)
	}

	conn, err := overlay.NewClientConn(
		"tunclient0",
		opt.ClientVirtualIP,
		opt.ServerCIDRs,
		opt.ProxyAddress,
	)
	if err != nil {
		return errors.Wrap(err, "cannot create connection to server")
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
	go switchProxyAddress(ctx, conn, "11.22.33.55:5525") // TODO add scheduler address to option
	conn.Run(ctx)

	return nil
}

// recieve command packet from scheduler to change proxy address
func switchProxyAddress(ctx context.Context, clientConn *overlay.ClientConn, schedulerAddress string) {
	ch := make(chan struct{})
	go func() {
		// construct a tcp connection to scheduler
		conn, err := net.Dial("tcp", schedulerAddress)
		if err != nil {
			logrus.Errorf("cannot connect to scheduler: %v", err)
			close(ch)
			return
		}
		defer conn.Close()

		for {
			// read command packet from scheduler
			message, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				logrus.Errorf("cannot read command packet: %v", err)
				break
			}
			addr := message[:len(message)-1]

			// change proxy address
			err = clientConn.SetProxyAddr(addr)
			if err != nil {
				logrus.Errorf("cannot change proxy address: %v", err)
				break
			}

			// TODO how does this ack packet work for the system?
			conn.Write([]byte(addr + "\n"))
		}
		close(ch)
	}()

	select {
	case <-ch:
		logrus.Infof("stopped reading")
	case <-ctx.Done():
	}
}
