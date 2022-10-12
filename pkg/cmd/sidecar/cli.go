package sidecar

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
	"github.com/spf13/cobra"
)

var (
	packetLog = logrus.WithField("client", "packet")
	tunLog    = logrus.WithField("client", "tun")
	udpLog    = logrus.WithField("client", "udp")
)

var (
	tunIP net.IP
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

	tunIP, cidr, err := net.ParseCIDR("10.10.20.0/24")
	if err != nil {
		return err
	}

	tunLog.Infof("tun IP: %s", tunIP.String())

	newTun, err := overlay.NewTun(tunLog.Logger, "mytunsrv", cidr, 1300, []overlay.Route{}, 500, false)
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
	s, err := net.ResolveUDPAddr("udp4", ":2345")
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

		pkt := gopacket.NewPacket(buf[:n], layers.LayerTypeIPv4, gopacket.Lazy)
		ipLayer := pkt.Layer(layers.LayerTypeIPv4)
		if ipLayer == nil {
			tunLog.Errorf("not a ip layer packet")
			continue
		}

		ip := ipLayer.(*layers.IPv4)
		dst := ip.DstIP.String()
		src := ip.SrcIP.String()
		tunLog.Infof("src: %s", src)

		dstIP := net.ParseIP("10.10.10.1")

		tunLog.Infof("dst: %s, will be changed to %s", dst, dstIP.String())

		ip.DstIP = dstIP

		opts := gopacket.SerializeOptions{
			ComputeChecksums: true,
			FixLengths:       true,
		}

		newBuffer := gopacket.NewSerializeBuffer()
		err = gopacket.SerializePacket(newBuffer, opts, pkt)
		if err != nil {
			tunLog.Errorf("cannot serialize packet: %s", err)
			continue
		}

		outgoingPacket := newBuffer.Bytes()

		_, err = udpConn.Write(outgoingPacket)
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
		udpLog.Debugf("reveived packet: %x", buf[:n])

		pkt := gopacket.NewPacket(buf[:n], layers.LayerTypeIPv4, gopacket.Lazy)
		ipLayer := pkt.Layer(layers.LayerTypeIPv4)
		if ipLayer == nil {
			tunLog.Errorf("not a ip layer packet")
			continue
		}

		ip := ipLayer.(*layers.IPv4)
		dst := ip.DstIP.String()
		src := ip.SrcIP.String()

		srcIP := net.ParseIP("10.10.20.1")

		tunLog.Infof("src: %s, will be changed to %s", src, srcIP.String())
		tunLog.Infof("dst: %s", dst)

		ip.SrcIP = srcIP

		opts := gopacket.SerializeOptions{
			ComputeChecksums: true,
			FixLengths:       true,
		}

		newBuffer := gopacket.NewSerializeBuffer()
		err = gopacket.SerializePacket(newBuffer, opts, pkt)
		if err != nil {
			tunLog.Errorf("cannot serialize packet: %s", err)
			continue
		}

		outgoingPacket := newBuffer.Bytes()

		_, err = t.Write(outgoingPacket)
		if err != nil {
			tunLog.Errorf("cannot write outgoing packet: %s", err)
		}
	}
}

func runCmd(program string, args ...string) (*exec.Cmd, error) {
	cmd := exec.Command(program, args...)
	err := cmd.Run()
	return cmd, err
}
