package proxy

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/bupt-narc/rinp/pkg/packet"
	"github.com/bupt-narc/rinp/pkg/repository"
	"github.com/bupt-narc/rinp/pkg/util/nexthop"
	"github.com/sirupsen/logrus"
)

type ProxyConn struct {
	Conn
}

func NewProxyConn(
	ctx context.Context,
	listenPort int,
) (*ProxyConn, error) {
	conn := &ProxyConn{
		Conn{
			NextHop:    nexthop.NewNextHopMap(),
			listenPort: listenPort,
		},
	}

	// TODO get NextHop from etcd
	nextHops, err := repository.GetServicesNextHop(ctx, redisSidecar)
	if err != nil {
		return nil, err
	}
	for service, nextHop := range nextHops {
		_, cidr, err := net.ParseCIDR(service + "/32")
		if err != nil {
			return nil, err
		}
		conn.NextHop.SetNextHop(cidr, nextHop)
		logrus.Infof("set next hop: %s %s", cidr, nextHop)
	}

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

		// Only allow traffic that is scheduled to this proxy
		src := pkt.GetSrc().String()
		if !isClientSchedulerHere(src) && !isServiceAddr(src) {
			// TODO: lower the log level to DEBUG
			// This is not a error. Just to make it obvious for now.
			connLog.Errorf("traffic from client %s is not scheduled to this proxy", src)
			continue
		}

		connLog.Debugf("recv from udp, src: %s, dst: %s", pkt.GetSrc(), pkt.GetDst())

		if logrus.IsLevelEnabled(logrus.DebugLevel) {
			_, err := c.NextHop.GetNextHopByString(pkt.GetSrc().String())
			if err == nil {
				connLog.Debugf("updating old connection to %s: %s", pkt.GetSrc().String(), udpAddr.String())
			} else {
				connLog.Debugf("adding new connection to %s: %s", pkt.GetSrc().String(), udpAddr.String())
			}
		}

		// Get nextHop (client or sidecar)
		nextHop := findNextHop(pkt.GetDst().String(), c.NextHop)
		if nextHop == nil {
			connLog.Errorf("cannot get nexthop for %s", pkt.GetDst().String())
			continue
		}

		// Save connection to client. Packets returned from sidecar will use it.
		// connection to sidecar will not be used later.
		c.NextHop.SetNextHopByString(pkt.GetSrc().String(), udpAddr.String())

		n, err = conn.WriteToUDP(packetData, nextHop)
		if err != nil {
			connLog.Errorf("cannot send packet to %s: %s", nextHop.String(), err)
			continue
		}

		c.txBytes += uint64(n)
		connLog.Debugf("send %d bytes to %s", n, nextHop.String())
	}
}

func isClientSchedulerHere(addr string) bool {
	// TODO: DO NOT USE BUILT-IN CLIENT SIDE CACHING
	// Because when cache misses, it will send a request to Redis.
	// But all malicious packets are cache misses, this will cause a huge pressure on Redis.
	// Instead, use manual cache that is synced with Redis.

	ctx := context.Background()
	//defer cancel()

	clientMsg := redisClient.DoCache(ctx, redisClient.B().Get().Key(addr).Cache(), 30*time.Second)

	if clientMsg.NonRedisError() != nil {
		connLog.Errorf("error when getting client %s from redis: %s", addr, clientMsg.NonRedisError())
		return false
	}

	_, err := clientMsg.ToMessage()

	if err != nil && !strings.Contains(err.Error(), "nil message") {
		return false
	}

	actualProxyName, _ := clientMsg.ToString()

	actualHost, _, _ := net.SplitHostPort(actualProxyName)

	return actualHost == opt.PublicIP
}

func isServiceAddr(addr string) bool {
	// TODO: DO NOT USE BUILT-IN CLIENT SIDE CACHING
	// Because when cache misses, it will send a request to Redis.
	// But all malicious packets are cache misses, this will cause a huge pressure on Redis.
	// Instead, use manual cache that is synced with Redis.
	//ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	//defer cancel()
	ctx := context.Background()

	sidecarMsg := redisSidecar.DoCache(ctx, redisSidecar.B().Get().Key(addr).Cache(), 30*time.Second)

	if sidecarMsg.NonRedisError() != nil {
		connLog.Errorf("error when getting sidecar address from redis: %s", sidecarMsg.NonRedisError())
		return false
	}

	_, err := sidecarMsg.ToMessage()

	if err != nil {
		return false
	} else {
		// Got a sidecar addr
		return true
	}

	return false
}

// TODO: use local cache; refactor;
func findNextHop(addr string, clientMap nexthop.NextHopMap) *net.UDPAddr {
	isClientRoute := false
	udpAddr, err := clientMap.GetNextHopByString(addr)
	var clientAddr *net.UDPAddr
	if err == nil {
		clientAddr = udpAddr
		isClientRoute = true
	}

	connLog.Debugf("find nexthop vip=%s, msg=%s", addr, udpAddr.String())

	if isClientRoute {
		connLog.Debugf("returning client addr")
		return clientAddr
	}

	isServerRoute := false

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	sidecarMsg := redisSidecar.DoCache(ctx, redisSidecar.B().Get().Key(addr).Cache(), 30*time.Second)
	if sidecarMsg.NonRedisError() != nil {
		connLog.Errorf("error when getting sidecar address from redis: %s", sidecarMsg.NonRedisError())
		return nil
	}

	connLog.Debugf("find sidecar vip=%s, msg=%s", addr, sidecarMsg)

	_, err = sidecarMsg.ToMessage()

	if err == nil {
		// Got a sidecar addr
		isServerRoute = true
	}

	if !isClientRoute && !isServerRoute {
		return nil
	}

	// FIXME: use newer information from Redis to update local cache
	// Currently, local cache have higher priority, so we will miss Redis updates.

	if isServerRoute {
		addr, err := sidecarMsg.ToString()
		if err != nil {
			return nil
		}
		connLog.Debugf("returning sidecar addr")
		nextUDPAddr, err := net.ResolveUDPAddr("udp4", addr)
		if err != nil {
			connLog.Errorf("cannot resolve udp addr: %s", err)
			return nil
		}
		return nextUDPAddr
	}

	return nil
}
