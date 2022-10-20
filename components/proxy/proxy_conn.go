package proxy

import (
	"fmt"
	"net"

	"github.com/bupt-narc/rinp/pkg/packet"
	"github.com/bupt-narc/rinp/pkg/util/nexthop"
	"github.com/sirupsen/logrus"
)

const (
	defaultServiceCIDR      = "11.22.33.44/32"
	defaultServiceAddress   = "service:12345"
	defaultSchedulerCIDR    = "11.22.33.55/32"
	defaultSchedulerAddress = "scheduler:12345"
)

type ProxyConn struct {
	Conn
}

func NewProxyConn(
	listenPort int,
) (*ProxyConn, error) {
	conn := &ProxyConn{
		Conn{
			NextHop:    nexthop.NewNextHopMap(),
			listenPort: listenPort,
		},
	}

	// TODO get NextHop from etcd
	_, ServiceCIDR, _ := net.ParseCIDR(defaultServiceCIDR)
	conn.NextHop.SetNextHop(ServiceCIDR, defaultServiceAddress)
	_, SchedulerCIDR, _ := net.ParseCIDR(defaultSchedulerCIDR)
	conn.NextHop.SetNextHop(SchedulerCIDR, defaultSchedulerAddress)

	conn.SetDealFunc(conn.deal)

	return conn, nil
}

func (c *ProxyConn) deal() {
	// listen on udp port
	udpAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", c.listenPort))
	if err != nil {
		connLog.Errorln("ResolveUDPAddr err:", err)
		return
	}
	conn, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		connLog.Errorln("ListenUDP err:", err)
		return
	}
	connLog.Infof("listening on port %d", c.listenPort)
	defer conn.Close()

	buf := make([]byte, 2000)
	for {
		if c.quit {
			break
		}

		n, udpAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			connLog.Errorf("cannot receive packet: %s", err)
			continue
		}

		packetData := buf[:n]
		connLog.Debugf("reveiced %d bytes", n)
		connLog.Tracef("received packet: %x", packetData)
		c.rxBytes += uint64(n)

		pkt, err := packet.NewFromLayer3Bytes(packetData)
		if err != nil {
			connLog.Errorf("error when parsing packet: %s", err)
			continue
		}

		connLog.Debugf("recv from udp, src: %s, dst: %s", pkt.GetSrc(), pkt.GetDst())

		// TODO get client NextHop from etcd
		if logrus.IsLevelEnabled(logrus.DebugLevel) {
			_, err := c.NextHop.GetNextHopByString(pkt.GetSrc().String())
			if err == nil {
				connLog.Debugf("updating old connection to %s: %s", pkt.GetSrc().String(), udpAddr.String())
			} else {
				connLog.Debugf("adding new connection to %s: %s", pkt.GetSrc().String(), udpAddr.String())
			}
		}
		c.NextHop.SetNextHopByString(pkt.GetSrc().String(), udpAddr.String())

		if _, err := c.NextHop.GetNextHop(pkt.GetSrc()); err != nil {
			// TODO deal with unknown traffic client
			connLog.Infof("find unknown traffic from %s", pkt.GetSrc().String())
			continue
		}

		if _, err := c.NextHop.GetNextHop(pkt.GetDst()); err != nil {
			connLog.Errorf("find unknown traffic to %s", pkt.GetDst().String())
			continue
		}

		// transfer packet to next hop by UDP
		nextHop, err := c.NextHop.GetNextHop(pkt.GetDst())
		if err != nil {
			connLog.Errorf("cannot find next hop for %s", pkt.GetDst().String())
			continue
		}
		n, err = conn.WriteToUDP(packetData, nextHop)
		if err != nil {
			connLog.Errorf("cannot send packet to %s: %s", nextHop.String(), err)
			continue
		}
		c.txBytes += uint64(n)
		connLog.Debugf("send %d bytes to %s", n, nextHop.String())
	}
}
