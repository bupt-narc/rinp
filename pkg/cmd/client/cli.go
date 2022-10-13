package client

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
	UserIP         net.IP
	IPString       = "10.10.20.0/24"
	IP             net.IP
	CIDR           *net.IPNet
	ServerCIDR     *net.IPNet
	ServerIPString = "172.17.0.2:32000"
)

func init() {
	var err error
	IP, CIDR, err = net.ParseCIDR(IPString)
	if err != nil {
		panic(err)
	}
	// User actual IP
	UserIP = net.ParseIP("10.10.200.1")
	// Server exposed IP
	_, ServerCIDR, _ = net.ParseCIDR("10.10.100.0/24")
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

	vpnIP := iputil.Ip2VpnIp(net.ParseIP("10.10.20.1").To4())
	newTun, err := overlay.NewTun(tunLog.Logger, "mytun", CIDR, 1300, []overlay.Route{{
		MTU:    1300,
		Metric: 0,
		Cidr:   ServerCIDR,
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
	go readUDPAndSendTUN(newTun, c)

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
		pkt.Modify(packet.ModifySrc(UserIP))
		tunLog.Debugf("udp packet out, src: %s, dst: %s", pkt.GetSrc(), pkt.GetDst())

		out, err := pkt.Serialize()
		if err != nil {
			tunLog.Errorf("error when serializing packet: %s", err)
			continue
		}

		_, err = udpConn.Write(out)
		if err != nil {
			udpLog.Errorf("cannot send packet: %s", err)
		}
	}
}

func readUDPAndSendTUN(t *overlay.Tun, udpConn *net.UDPConn) {
	buf := make([]byte, 2000)
	for {
		n, _, err := udpConn.ReadFromUDP(buf)
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
		//pkt.Modify(packet.ModifyDst(net.ParseIP("127.0.0.1")))
		udpLog.Debugf("tun packet out, src: %s, dst: %s", pkt.GetSrc(), pkt.GetDst())

		out, err := pkt.Serialize()
		if err != nil {
			udpLog.Errorf("error when serializing packet: %s", err)
			continue
		}

		n, err = t.Write(out)
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
