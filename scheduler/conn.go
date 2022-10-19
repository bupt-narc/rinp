package scheduler

import (
	"context"
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
	connLog = logrus.WithField("scheduler", "connection")
)

type Conn struct {
	rxBytes      uint64
	txBytes      uint64
	lastStatTime time.Time
	quit         bool
	dealFunc     func()
	listenPort   int
}

func (c *Conn) Run(ctx context.Context) {
	ch := make(chan struct{})
	go func() {
		c.dealFunc()
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
}

func (c *Conn) SetDealFunc(f func()) {
	c.dealFunc = f
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
