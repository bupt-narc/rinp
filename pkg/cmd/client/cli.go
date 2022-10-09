package client

import (
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/bupt-narc/rinp/pkg/packet"
	"github.com/pkg/errors"
	"github.com/songgao/water"
	"golang.org/x/net/ipv4"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	packetLog = logrus.WithField("client", "packet")
	tunLog    = logrus.WithField("client", "tun")
	udpLog    = logrus.WithField("client", "udp")
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

	ifce, err := water.New(water.Config{
		DeviceType: water.TUN,
	})
	if err != nil {
		return errors.Wrap(err, "cannot create TUN device")
	}
	tunLog.Infof("created TUN device: %s", ifce.Name())

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
	s, err := net.ResolveUDPAddr("udp4", "127.0.0.1:3456") // FIXME
	if err != nil {
		return err
	}
	c, err := net.DialUDP("udp4", nil, s)
	if err != nil {
		return err
	}

	udpLog.Infof("connected to udp server %s", c.RemoteAddr().String())
	defer c.Close()

	go readTUNAndWriteUDP(ifce, c)

	// Listen to termination signals.
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM)
	signal.Notify(sigterm, syscall.SIGINT)
	<-sigterm

	return nil
}

func readTUNAndWriteUDP(ifce *water.Interface, udpConn *net.UDPConn) {
	buf := make([]byte, 2000)
	for {
		n, err := ifce.Read(buf)
		if err != nil {
			tunLog.Errorf("cannot receive packet: %s", err)
			continue
		}
		tunLog.Infof("reveiced %d bytes", n)
		tunLog.Debugf("received packet: %x", buf[:n])

		header := ipv4.Header{}
		err = header.Parse(buf[:n])
		if err != nil {
			packetLog.Errorf("cannot parse ipv4 header: %s", err)
			continue
		}
		packetLog.Infof("packet is to %s", header.Dst.String())
		packetLog.Debugf("%#v", header)

		pkt := packet.Packet{
			PacketVersion: packet.Version0,
			IPVersion:     packet.IPv4,
			Type:          packet.DataTransfer,
			Src:           net.IPv4(127, 0, 0, 1),
			SrcPort:       0,
			Dst:           net.IPv4(127, 0, 0, 1),
			DstPort:       0,
			DataLength:    uint16(n),
			Data:          buf[:n],
		}
		pktBytes, err := packet.Marshal(pkt)
		if err != nil {
			packetLog.Errorf("cannot marshal packet: %s", err)
			return
		}

		n, err = udpConn.Write(pktBytes)
		if err != nil {
			udpLog.Errorf("cannot send packet: %s", err)
		}
		udpLog.Infof("sent %d bytes", n)
		udpLog.Debugf("sent packet: %x", pktBytes)
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
