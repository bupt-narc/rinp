package scheduler

import (
	"bufio"
	"fmt"
	"net"
	"time"
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
		if c.quit {
			break
		}
		sendNextProxyAddr(conn, fmt.Sprintf("proxy%d:5114", index+1))
		index = (index + 1) % 3
		time.Sleep(5000 * time.Millisecond)
	}

}

func sendNextProxyAddr(conn net.Conn, addr string) error {
	_, err := conn.Write([]byte(addr + "\n"))
	if err == nil {
		connLog.Debugf("let user %s change to %s", conn.RemoteAddr().(*net.TCPAddr).IP.String(), addr)
	}
	// read command packet from scheduler
	message, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		connLog.Errorf("cannot read command packet: %v", err)
		return err
	}
	message = message[:len(message)-1]
	connLog.Debugf("user %s change to %s", conn.RemoteAddr().(*net.TCPAddr).IP.String(), message)
	return err
}
