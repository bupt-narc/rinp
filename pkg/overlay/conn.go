package overlay

import (
	"context"
	"net"
	"time"

	"github.com/bupt-narc/rinp/pkg/util/bytesize"
	"github.com/sirupsen/logrus"
)

const (
	DefaultMTU      = 1400
	statDuration    = 1 * time.Second
	minimumStatSize = 1024
)

var (
	connLog = logrus.WithField("overlay", "connection")
)

type Conn struct {
	udpConn      *net.UDPConn
	tun          *Tun
	rxBytes      uint64
	txBytes      uint64
	lastStatTime time.Time
	quit         bool
	rxFunc       func()
	txFunc       func()
}

func (c *Conn) Run(ctx context.Context) {
	ch := make(chan struct{})
	go func() {
		c.rxFunc()
		close(ch)
	}()
	go func() {
		c.txFunc()
		close(ch)
	}()
	go func() {
		c.stat()
	}()

	select {
	case <-ch:
		connLog.Infof("stopped reading")
	case <-ctx.Done():
	}

	c.quit = true
	// TODO: make sure it is not nil
	c.tun.Close()
	c.udpConn.Close()
}

func (c *Conn) SetRxFunc(f func()) {
	c.rxFunc = f
}

func (c *Conn) SetTxFunc(f func()) {
	c.txFunc = f
}

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
