package overlay

import (
	"fmt"
	"net"
	"strings"

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
	serverRoutes []*net.IPNet,
	serverAddr string,
) (*ClientConn, error) {
	overlayRoutes, err := ipNetToRoutes(serverRoutes, clientIP)
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

	err = conn.SetProxyAddr(serverAddr)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot set server address")
	}

	conn.SetRxFunc(func() {
		conn.readUDPAndSendTUN()
	})

	conn.SetTxFunc(func() {
		conn.readTUNAndWriteUDP()
	})

	connLog.Infof("client connection activated, clientAddr=%s, serverRoutes=%v", clientIP, serverRoutes)

	return conn, nil
}

func (c *ClientConn) SetProxyAddr(addr string) error {
	s, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		connLog.Errorf("cannot resolve proxy address: %s", err)
		return err
	}
	conn, err := net.DialUDP("udp4", nil, s)
	if err != nil {
		connLog.Errorf("cannot dial proxy: %s", err)
		return err
	}
	oldConn := c.udpConn
	c.udpConn = conn
	if oldConn != nil {
		oldConn.Close()
	}
	connLog.Infof("set proxy address to: %s", addr)
	return nil
}

func (c *ClientConn) readTUNAndWriteUDP() {
	buf := make([]byte, 2000)
	for {
		n, err := c.tun.Read(buf)
		if err != nil {
			if c.quit {
				break
			}
			if strings.Contains(err.Error(), "closed network connection") {
				continue
			}
			connLog.Errorf("cannot receive packet: %s", err)
			continue
		}
		packetData := buf[:n]

		if logrus.IsLevelEnabled(logrus.DebugLevel) {
			connLog.Debugf("reveiced %d bytes", n)
			connLog.Tracef("received packet: %x", packetData)

			pkt, err := packet.NewFromLayer3Bytes(packetData)
			if err != nil {
				connLog.Errorf("error when parsing packet: %s", err)
				continue
			}

			connLog.Debugf("recv from tun, src: %s, dst: %s", pkt.GetSrc(), pkt.GetDst())
		}

		_, err = c.udpConn.Write(packetData)
		if err != nil {
			if c.quit {
				break
			}
			if strings.Contains(err.Error(), "closed network connection") {
				continue
			}
			connLog.Errorf("cannot send packet: %s", err)
		}
		c.txBytes += uint64(n)
	}
}

func (c *ClientConn) readUDPAndSendTUN() {
	buf := make([]byte, 2000)
	for {
		n, err := c.udpConn.Read(buf)
		if err != nil {
			if c.quit {
				break
			}
			if strings.Contains(err.Error(), "closed network connection") {
				continue
			}
			connLog.Errorf("cannot receive packet: %s", err)
			continue
		}

		packetData := buf[:n]

		if logrus.IsLevelEnabled(logrus.DebugLevel) {
			connLog.Debugf("reveiced %d bytes", n)
			connLog.Tracef("received packet: %x", packetData)

			pkt, err := packet.NewFromLayer3Bytes(packetData)
			if err != nil {
				connLog.Errorf("error when parsing packet: %s", err)
				continue
			}

			connLog.Debugf("recv from udp, src: %s, dst: %s", pkt.GetSrc(), pkt.GetDst())
		}

		n, err = c.tun.Write(packetData)
		if err != nil {
			if c.quit {
				break
			}
			if strings.Contains(err.Error(), "closed network connection") {
				continue
			}
			connLog.Errorf("cannot send packet: %s", err)
		}
		c.rxBytes += uint64(n)
	}
}
