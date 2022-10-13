package packet

import (
	"errors"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type Packet struct {
	pkt gopacket.Packet
	l3  *layers.IPv4
}

var (
	ErrNotLayer4Packet = errors.New("not a layer 4 packet")
)

func NewFromLayer4Bytes(b []byte) (*Packet, error) {
	pkt := gopacket.NewPacket(b, layers.LayerTypeIPv4, gopacket.Lazy)
	ipLayer := pkt.Layer(layers.LayerTypeIPv4)

	if ipLayer == nil {
		return nil, ErrNotLayer4Packet
	}

	ip, ok := ipLayer.(*layers.IPv4)
	if !ok {
		return nil, ErrNotLayer4Packet
	}

	return &Packet{
		pkt: pkt,
		l3:  ip,
	}, nil
}

func (p *Packet) GetDst() net.IP {
	return p.l3.DstIP
}

func (p *Packet) GetSrc() net.IP {
	return p.l3.SrcIP
}

func (p *Packet) GetDstPort() string {
	tcpLayer := p.pkt.Layer(layers.LayerTypeTCP)
	if tcpLayer != nil {
		if l4, ok := tcpLayer.(*layers.TCP); ok {
			return l4.DstPort.String()
		}
	}

	udpLayer := p.pkt.Layer(layers.LayerTypeUDP)
	if udpLayer != nil {
		if l4, ok := udpLayer.(*layers.UDP); ok {
			return l4.DstPort.String()
		}
	}
	return ""
}

func (p *Packet) GetSrcPort() string {
	tcpLayer := p.pkt.Layer(layers.LayerTypeTCP)
	if tcpLayer != nil {
		if l4, ok := tcpLayer.(*layers.TCP); ok {
			return l4.SrcPort.String()
		}
	}

	udpLayer := p.pkt.Layer(layers.LayerTypeUDP)
	if udpLayer != nil {
		if l4, ok := udpLayer.(*layers.UDP); ok {
			return l4.SrcPort.String()
		}
	}
	return ""
}

func (p *Packet) Modify(mods ...Mod) {
	for _, m := range mods {
		m(p.l3)
	}
}

func (p *Packet) Serialize() ([]byte, error) {
	opts := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}

	buf := gopacket.NewSerializeBuffer()

	tcpLayer := p.pkt.Layer(layers.LayerTypeTCP)
	if tcpLayer != nil {
		if l4, ok := tcpLayer.(*layers.TCP); ok {
			err := l4.SetNetworkLayerForChecksum(p.l3)
			if err != nil {
				return nil, err
			}
		}
	}

	udpLayer := p.pkt.Layer(layers.LayerTypeUDP)
	if udpLayer != nil {
		if l4, ok := udpLayer.(*layers.UDP); ok {
			err := l4.SetNetworkLayerForChecksum(p.l3)
			if err != nil {
				return nil, err
			}
		}
	}

	err := gopacket.SerializePacket(buf, opts, p.pkt)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
