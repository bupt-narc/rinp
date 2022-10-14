package overlay

import (
	"net"
	"time"

	"github.com/bupt-narc/rinp/pkg/util/bytesize"
	"github.com/sirupsen/logrus"
)

type Conn struct {
	udpConn      *net.UDPConn
	tun          *Tun
	rxBytes      uint64
	txBytes      uint64
	lastStatTime time.Time
}

const (
	DefaultMTU      = 1300
	statDuration    = 1 * time.Second
	minimumStatSize = 1024
)

var (
	connLog = logrus.WithField("overlay", "connection")
)

func (c *Conn) stat() {
	c.lastStatTime = time.Now()
	for {
		time.Sleep(statDuration)

		if c.rxBytes < minimumStatSize && c.txBytes < minimumStatSize {
			continue
		}

		d := time.Since(c.lastStatTime).Seconds()
		qRx := bytesize.ByteCountBinary(int64(c.rxBytes))
		qRxPs := bytesize.ByteCountBinary(int64(float64(c.rxBytes) / d))
		qTx := bytesize.ByteCountBinary(int64(c.txBytes))
		qTxPs := bytesize.ByteCountBinary(int64(float64(c.txBytes) / d))

		c.rxBytes = 0
		c.txBytes = 0
		c.lastStatTime = time.Now()

		connLog.Infof("rx: %s (%s/s), tx: %s (%s/s), duration: %.2fs",
			qRx, qRxPs,
			qTx, qTxPs,
			d,
		)
	}
}
