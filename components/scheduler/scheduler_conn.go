package scheduler

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"time"

	"github.com/bupt-narc/rinp/pkg/util/iplist"
)

type SchedulerConn struct {
	Conn
}

func NewSchedulerConn(
	listenPort int,
) (*SchedulerConn, error) {
	conn := &SchedulerConn{
		Conn{
			listenPort: listenPort,
		},
	}

	conn.SetDealFunc(conn.deal)

	return conn, nil
}

func (c *SchedulerConn) deal() {
	TCPaddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf(":%d", c.listenPort))
	if err != nil {
		connLog.Errorln("ResolveTCPAddr err:", err)
		return
	}
	listener, err := net.ListenTCP("tcp4", TCPaddr)
	if err != nil {
		connLog.Errorln("ListenTCP err:", err)
		return
	}
	connLog.Infof("listening on port %d", c.listenPort)
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			connLog.Errorln("Accept err:", err)
			continue
		}
		// TODO: do some authentication to prevent malicious connections
		go c.schedule(conn)
	}
}

func (c *SchedulerConn) schedule(conn net.Conn) {
	defer conn.Close()

	// get remote IP
	remoteIP := conn.RemoteAddr().(*net.TCPAddr).IP.String()
	connLog.Debugf("user %s connected", remoteIP)

	// TODO real schedule

	// this is fake schedule
	index := 0

	for {
		ctx, cancel := context.WithCancel(context.Background())
		if c.quit {
			// TODO: refactor using defer cancel() instead, including cancel()s below
			cancel()
			break
		}

		// Get number of proxies
		dbsize, err := redisProxy.Do(ctx, redisProxy.B().Dbsize().Build()).ToInt64()
		if err != nil {
			connLog.Errorf("cannot get number of proxies: %s", err)
			cancel()
			continue
		}

		proxyName := fmt.Sprintf("proxy%d", index+1)
		proxyIP, err := redisProxy.DoCache(ctx, redisProxy.B().Get().Key(proxyName).Cache(), 30*time.Second).ToString()
		if err != nil {
			connLog.Errorf("cannot get public address of proxy %s", proxyName)
			cancel()
			continue
		}

		sendNextProxyAddr(conn, proxyIP)
		index = (index + 1) % int(dbsize)
		time.Sleep(5000 * time.Millisecond)
		cancel()
	}

}

func sendNextProxyAddr(conn net.Conn, addr string) error {
	_, err := conn.Write([]byte(addr + "\n"))
	if err == nil {
		connLog.Debugf("let user %s change to %s", conn.RemoteAddr().(*net.TCPAddr).IP.String(), addr)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// TODO: take network latency into consideration by invalidating the old proxy name with a delay
	// Keep only IP, strip port
	clientConn := conn.RemoteAddr().String()
	host, _, _ := net.SplitHostPort(clientConn)
	err = redisClient.Do(ctx, redisClient.B().Set().Key(host).Value(iplist.ToString(addr)).Build()).Error()
	if err != nil {
		connLog.Errorf("cannot set client %s to proxy %s: %s", host, addr, err)
		return err
	}

	// read command packet from scheduler
	message, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		connLog.Errorf("cannot read command packet: %v", err)
		return err
	}
	message = message[:len(message)-1]
	connLog.Debugf("switched user %s to proxy %s", conn.RemoteAddr().(*net.TCPAddr).IP.String(), message)
	return err
}
