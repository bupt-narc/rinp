package overlay

import (
	"context"
	"fmt"
	"net"

	"github.com/bupt-narc/rinp/pkg/packet"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type ClientConn struct {
	Conn
}

func NewClientConn(
	name string,
	clientIP net.IP,
	serverRoutes []string,
	serverAddr string,
) (*ClientConn, error) {
	overlayRoutes, err := stringToRoutes(serverRoutes, clientIP)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot parse routes")
	}

	_, cidr, err := net.ParseCIDR(fmt.Sprintf("%s/32", clientIP.String()))
	if err != nil {
		return nil, errors.Wrapf(err, "cannot parse client ip")
	}

	newTun, err := NewTun(
		connLog.Logger,
		name,
		cidr,
		DefaultMTU,
		overlayRoutes,
		500,
		false,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot new tun device")
	}

	err = newTun.Activate()
	if err != nil {
		return nil, errors.Wrapf(err, "cannot activate tun device")
	}

	conn := &ClientConn{
		Conn{
			tun: newTun,
		},
	}

	err = conn.SetServerAddr(serverAddr)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot set server address")
	}

	connLog.Infof("client connection activated, clientAddr=%s, serverRoutes=%v", clientIP, serverRoutes)

	return conn, nil
}

func (c *ClientConn) SetServerAddr(addr string) error {
	s, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp4", nil, s)
	if err != nil {
		return err
	}
	c.udpConn = conn
	return nil
}

func (c *ClientConn) Run(ctx context.Context) {
	ch := make(chan struct{})
	go func() {
		c.readTUNAndWriteUDP()
		close(ch)
	}()
	go func() {
		c.readUDPAndSendTUN()
		close(ch)
	}()

	select {
	case <-ch:
		connLog.Infof("stopped reading")
	case <-ctx.Done():
		c.tun.Close()
		c.udpConn.Close()
	}
}

func (c *ClientConn) readTUNAndWriteUDP() {
	buf := make([]byte, 2000)
	for {
		n, err := c.tun.Read(buf)
		if err != nil {
			connLog.Errorf("cannot receive packet: %s", err)
			continue
		}
		packetData := buf[:n]
		connLog.Debugf("reveiced %d bytes", n)
		connLog.Tracef("received packet: %x", packetData)

		if logrus.IsLevelEnabled(logrus.DebugLevel) {
			pkt, err := packet.NewFromLayer4Bytes(packetData)
			if err != nil {
				connLog.Errorf("error when parsing packet: %s", err)
				continue
			}

			connLog.Debugf("recv from tun, src: %s, dst: %s", pkt.GetSrc(), pkt.GetDst())
		}

		_, err = c.udpConn.Write(packetData)
		if err != nil {
			connLog.Errorf("cannot send packet: %s", err)
		}
	}
}

func (c *ClientConn) readUDPAndSendTUN() {
	buf := make([]byte, 2000)
	for {
		n, err := c.udpConn.Read(buf)
		if err != nil {
			connLog.Errorf("cannot receive packet: %s", err)
			continue
		}

		packetData := buf[:n]
		connLog.Debugf("reveiced %d bytes", n)
		connLog.Tracef("received packet: %x", packetData)

		if logrus.IsLevelEnabled(logrus.DebugLevel) {
			pkt, err := packet.NewFromLayer4Bytes(packetData)
			if err != nil {
				connLog.Errorf("error when parsing packet: %s", err)
				continue
			}

			connLog.Debugf("recv from udp, src: %s, dst: %s", pkt.GetSrc(), pkt.GetDst())
		}

		n, err = c.tun.Write(packetData)
		if err != nil {
			connLog.Errorf("cannot send packet: %s", err)
		}
	}
}
