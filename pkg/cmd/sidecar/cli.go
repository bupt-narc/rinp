package sidecar

import (
	"context"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/bupt-narc/rinp/pkg/overlay"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	packetLog = logrus.WithField("client", "packet")
	tunLog    = logrus.WithField("client", "tun")
	udpLog    = logrus.WithField("client", "udp")
)

var (
	// tunIP net.IP
	// ServerIP net.IP
	UserCIDR *net.IPNet
)

func init() {
	// Server actual IP
	//ServerIP = net.ParseIP("10.10.100.1")
	// User actual IP
	_, UserCIDR, _ = net.ParseCIDR("10.10.20.0/24")
}

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

	conn, err := overlay.NewServerConn(
		"tun0",
		net.ParseIP("10.10.10.1"),
		opt.Port,
		[]string{"10.10.20.0/24"},
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

	conn.Run(ctx)

	return nil
}

func runCmd(program string, args ...string) (*exec.Cmd, error) {
	cmd := exec.Command(program, args...)
	err := cmd.Run()
	return cmd, err
}
