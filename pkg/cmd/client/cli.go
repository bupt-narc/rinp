package client

import (
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/bupt-narc/rinp/pkg/overlay"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/songgao/water"
	"github.com/spf13/cobra"
)

var (
	packetLog = logrus.WithField("client", "packet")
	tunLog    = logrus.WithField("client", "tun")
	udpLog    = logrus.WithField("client", "udp")
)

var (
	IPString = "10.10.10.1/24"
	IP       net.IP
	CIDR     *net.IPNet

	ServerIPString = "127.0.0.1:2345"
)

func init() {
	var err error
	IP, CIDR, err = net.ParseCIDR(IPString)
	if err != nil {
		panic(err)
	}
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

	newTun, err := overlay.NewTun(tunLog.Logger, "mytun", CIDR, 1300, []overlay.Route{}, 500, false)
	if err != nil {
		return err
	}
	tunLog.Infof("created device")
	err = newTun.Activate()
	if err != nil {
		return err
	}
	tunLog.Infof("activated device")

	//_, err = runCmd("ip", "addr", "add", "10.255.255.1/24", "dev", ifce.Name())
	//if err != nil {
	//	return errors.Wrapf(err, "cannot add address to %s", ifce.Name())
	//}
	//
	//_, err = runCmd("ip", "link", "set", "dev", ifce.Name(), "up")
	//if err != nil {
	//	return errors.Wrapf(err, "cannot start %s", ifce.Name())
	//}

	// Connect UDP
	s, err := net.ResolveUDPAddr("udp4", ServerIPString) // FIXME
	if err != nil {
		return err
	}
	c, err := net.DialUDP("udp4", nil, s)
	if err != nil {
		return err
	}

	udpLog.Infof("connected to udp server %s", c.RemoteAddr().String())
	defer c.Close()

	go readTUNAndWriteUDP(newTun, c)

	// Listen to termination signals.
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM)
	signal.Notify(sigterm, syscall.SIGINT)
	<-sigterm

	return nil
}

func readTUNAndWriteUDP(t *overlay.Tun, udpConn *net.UDPConn) {
	buf := make([]byte, 2000)
	for {
		n, err := t.Read(buf)
		if err != nil {
			tunLog.Errorf("cannot receive packet: %s", err)
			continue
		}
		packetData := buf[:n]
		tunLog.Infof("reveiced %d bytes", n)
		tunLog.Debugf("received packet: %x", packetData)

		pkt := gopacket.NewPacket(packetData, layers.LayerTypeIPv4, gopacket.Lazy)

		tunLog.Infof("src: %s", pkt.NetworkLayer().NetworkFlow().Src().String())
		tunLog.Infof("dst: %s", pkt.NetworkLayer().NetworkFlow().Dst().String())

		_, err = udpConn.Write(packetData)
		if err != nil {
			udpLog.Errorf("cannot send packet: %s", err)
		}
	}
}

func readUDPAndSendTUN(ifce *water.Interface, udpConn *net.UDPConn) {
	buf := make([]byte, 2000)
	for {
		n, _, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			udpLog.Errorf("cannot receive packet: %s", err)
			continue
		}
		udpLog.Debugf("reveived packet: %x", buf[:n])

		n, err = ifce.Write(buf[:n])
		if err != nil {
			tunLog.Errorf("cannot send packet: %s", err)
		}
	}
}

func runCmd(program string, args ...string) (*exec.Cmd, error) {
	cmd := exec.Command(program, args...)
	err := cmd.Run()
	return cmd, err
}
