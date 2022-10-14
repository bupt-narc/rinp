package overlay

import (
	"net"

	"github.com/sirupsen/logrus"
)

type Conn struct {
	udpConn *net.UDPConn
	tun     *Tun
}

const (
	DefaultMTU = 1300
)

var (
	connLog = logrus.WithField("overlay", "connection")
)
