package packet

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	pkt0 = Packet{
		PacketVersion: 1,
		IPVersion:     0,
		Type:          5,
		Src:           net.IPv4bcast,
		SrcPort:       12384,
		Dst:           net.IPv4allsys,
		DstPort:       31453,
		DataLength:    2,
		Data:          []byte{0xa2, 0x3b},
	}
	pkt0Golden = []byte{
		0b0101_0001,
		255, 255, 255, 255,
		0x60, 0x30,
		224, 0, 0, 1,
		0xdd, 0x7a,
		0x2, 0x0,
		0xa2, 0x3b,
	}
)

func TestPacket_Marshal(t *testing.T) {
	b, err := Marshal(pkt0)
	assert.NoError(t, err)
	assert.Equal(t, pkt0Golden, b)
}

func TestPacket_UnMarshal(t *testing.T) {
	newPkt := New()
	err := UnMarshal(pkt0Golden, newPkt)
	assert.NoError(t, err)
	assert.Equal(t, &pkt0, newPkt)
}
