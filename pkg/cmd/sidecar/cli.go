package sidecar

import (
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/bupt-narc/rinp/pkg/overlay"
	"github.com/bupt-narc/rinp/pkg/packet"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/slackhq/nebula/iputil"
	"github.com/spf13/cobra"
)

var (
	packetLog = logrus.WithField("client", "packet")
	tunLog    = logrus.WithField("client", "tun")
	udpLog    = logrus.WithField("client", "udp")
)

var (
	tunIP    net.IP
	ServerIP net.IP
	UserCIDR *net.IPNet
)

func init() {
	// Server actual IP
	ServerIP = net.ParseIP("10.10.100.1")
	// User actual IP
	_, UserCIDR, _ = net.ParseCIDR("10.10.200.0/24")
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

	// tun IP
	tunIP, cidr, err := net.ParseCIDR("10.10.10.0/24")
	if err != nil {
		return err
	}

	tunLog.Infof("tun IP: %s", tunIP.String())

	vpnIP := iputil.Ip2VpnIp(net.ParseIP("10.10.10.1").To4())
	newTun, err := overlay.NewTun(tunLog.Logger, "mytunsrv", cidr, 1300, []overlay.Route{{
		MTU:    1300,
		Metric: 0,
		Cidr:   UserCIDR,
		Via:    &vpnIP,
	}}, 500, false)
	if err != nil {
		return err
	}
	tunLog.Infof("created device")
	err = newTun.Activate()
	if err != nil {
		return err
	}
	tunLog.Infof("activated device")

	// Connect UDP
	s, err := net.ResolveUDPAddr("udp4", ":32000")
	if err != nil {
		return err
	}

	connection, err := net.ListenUDP("udp4", s)
	if err != nil {
		return err
	}

	go readUDPAndSendTUN(newTun, connection)
	go readTUNAndWriteUDP(newTun, connection)

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

		pkt, err := packet.NewFromLayer4Bytes(packetData)
		if err != nil {
			tunLog.Errorf("error when parsing packet: %s", err)
			continue
		}

		tunLog.Debugf("recv from tun, src: %s, dst: %s", pkt.GetSrc(), pkt.GetDst())
		//pkt.Modify(packet.ModifySrc(ServerIP))
		tunLog.Debugf("udp packet out, src: %s, dst: %s", pkt.GetSrc(), pkt.GetDst())

		out, err := pkt.Serialize()
		if err != nil {
			tunLog.Errorf("error when serializeing packet: %s", err)
			continue
		}

		_, err = udpConn.WriteToUDP(out, udpAddr)

		if err != nil {
			udpLog.Errorf("cannot send packet: %s", err)
		}
	}
}

var udpAddr *net.UDPAddr

func readUDPAndSendTUN(t *overlay.Tun, udpConn *net.UDPConn) {
	buf := make([]byte, 2000)
	for {
		var (
			n   int
			err error
		)
		n, udpAddr, err = udpConn.ReadFromUDP(buf)
		if err != nil {
			udpLog.Errorf("cannot receive packet: %s", err)
			continue
		}
		packetData := buf[:n]
		udpLog.Infof("reveiced %d bytes", n)
		udpLog.Debugf("received packet: %x", packetData)

		pkt, err := packet.NewFromLayer4Bytes(packetData)
		if err != nil {
			udpLog.Errorf("error when parsing packet: %s", err)
			continue
		}

		udpLog.Debugf("recv from udp, src: %s, dst: %s", pkt.GetSrc(), pkt.GetDst())
		udpLog.Debugf("srcPort: %s, dstPort: %s", pkt.GetSrcPort(), pkt.GetDstPort())

		out, err := pkt.Serialize()
		if err != nil {
			udpLog.Errorf("error when serializing packet: %s", err)
			continue
		}

		n, err = t.Write(out)
		if err != nil {
			tunLog.Errorf("cannot write outgoing packet: %s", err)
		}
		tunLog.Debugf("written %d bytes", n)
	}
}

func runCmd(program string, args ...string) (*exec.Cmd, error) {
	cmd := exec.Command(program, args...)
	err := cmd.Run()
	return cmd, err
}
