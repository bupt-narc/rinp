package packet

import (
	"net"

	"github.com/google/gopacket/layers"
)

type Mod func(l *layers.IPv4)

func ModifyDst(dstIP net.IP) Mod {
	return func(l *layers.IPv4) {
		l.DstIP = dstIP
	}
}

func ModifySrc(srcIP net.IP) Mod {
	return func(l *layers.IPv4) {
		l.SrcIP = srcIP
	}
}

func ModifyData(b []byte) Mod {
	return func(l *layers.IPv4) {
		l.Contents = b
	}
}
