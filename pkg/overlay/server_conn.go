package overlay

import (
	"context"
	"fmt"
	"net"

	"github.com/bupt-narc/rinp/pkg/packet"
	"github.com/dboslee/lru"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	ClientPoolSize = 1024
)

type ServerConn struct {
	Conn
	clientPool *lru.SyncCache[string, *net.UDPAddr]
}

func NewServerConn(
	name string,
	serverIP net.IP,
	listenPort int,
	clientRoutes []string,
) (*ServerConn, error) {
	overlayRoutes, err := stringToRoutes(clientRoutes, serverIP)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot parse routes")
	}

	_, cidr, err := net.ParseCIDR(fmt.Sprintf("%s/32", serverIP.String()))
	if err != nil {
		return nil, errors.Wrapf(err, "cannot parse server ip")
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

	conn := &ServerConn{
		Conn: Conn{
			tun: newTun,
		},
		clientPool: lru.NewSync[string, *net.UDPAddr](lru.WithCapacity(ClientPoolSize)),
	}

	err = conn.SetListenPort(listenPort)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot set server address")
	}

	connLog.Infof("server connection activated, severAddr=%s, clientRoutes=%v", serverIP, clientRoutes)

	return conn, nil
}

func (s *ServerConn) SetListenPort(port int) error {
	udpAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		return err
	}
	connLog.Infof("listening on port %d", port)
	s.udpConn = conn
	return nil
}

func (c *ServerConn) Run(ctx context.Context) {
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

func (s *ServerConn) readTUNAndWriteUDP() {
	buf := make([]byte, 2000)
	for {
		n, err := s.tun.Read(buf)
		if err != nil {
			connLog.Errorf("cannot receive packet: %s", err)
			continue
		}
		packetData := buf[:n]
		connLog.Debugf("reveiced %d bytes", n)
		connLog.Tracef("received packet: %x", packetData)

		pkt, err := packet.NewFromLayer4Bytes(packetData)
		if err != nil {
			connLog.Errorf("error when parsing packet: %s", err)
			continue
		}

		connLog.Debugf("recv from tun, src: %s, dst: %s", pkt.GetSrc(), pkt.GetDst())
		udpAddr, ok := s.clientPool.Get(pkt.GetDst().String())
		if !ok {
			connLog.Errorf("cannot find connection to client %s", pkt.GetDst())
			continue
		}

		n, err = s.udpConn.WriteToUDP(packetData, udpAddr)
		if err != nil {
			connLog.Errorf("cannot send packet: %s", err)
		}
		connLog.Debugf("written %d bytes to udp", n)
	}
}

func (s *ServerConn) readUDPAndSendTUN() {
	buf := make([]byte, 2000)
	for {
		var (
			n       int
			err     error
			udpAddr *net.UDPAddr
		)
		n, udpAddr, err = s.udpConn.ReadFromUDP(buf)
		if err != nil {
			connLog.Errorf("cannot receive packet: %s", err)
			continue
		}

		packetData := buf[:n]
		connLog.Debugf("reveiced %d bytes", n)
		connLog.Tracef("received packet: %x", packetData)

		pkt, err := packet.NewFromLayer4Bytes(packetData)
		if err != nil {
			connLog.Errorf("error when parsing packet: %s", err)
			continue
		}

		connLog.Debugf("recv from udp, src: %s, dst: %s", pkt.GetSrc(), pkt.GetDst())
		if logrus.IsLevelEnabled(logrus.DebugLevel) {
			_, ok := s.clientPool.Peek(pkt.GetSrc().String())
			if ok {
				connLog.Debugf("updating old connection to %s", pkt.GetSrc().String())
			} else {
				connLog.Debugf("adding new connection to %s", pkt.GetSrc().String())
			}
		}
		s.clientPool.Set(pkt.GetSrc().String(), udpAddr)

		n, err = s.tun.Write(packetData)
		if err != nil {
			connLog.Errorf("cannot write outgoing packet: %s", err)
		}
		connLog.Debugf("written %d bytes to tun", n)
	}
}
