package packet

import (
	"errors"
	"net"

	pkgerrors "github.com/pkg/errors"
)

func New() *Packet {
	return &Packet{}
}

func NewFromByteStream(in []byte) (*Packet, error) {
	p := New()
	err := UnMarshal(in, p)
	return p, err
}

const (
	packetVersionMask byte = 0b0000_0111
	ipVersionMask     byte = 0b0000_1000
	typeMask          byte = 0b1111_0000
)

var (
	ErrInvalidVersion   = errors.New("invalid version")
	ErrInvalidIPVersion = errors.New("invalid ip version")
	ErrInvalidType      = errors.New("invalid type")
)

func UnMarshal(in []byte, p *Packet) error {
	// TODO(charlie0129): check len(in)

	byte0 := in[0:1][0]
	ver := byte0 & packetVersionMask
	if !(ver <= Version7) {
		return pkgerrors.Wrapf(ErrInvalidVersion, "%d is invalid", ver)
	}
	p.PacketVersion = Version(ver)

	ipVer := (byte0 & ipVersionMask) >> 3
	if !(ipVer == IPv4 || ipVer == IPv6) {
		return pkgerrors.Wrapf(ErrInvalidIPVersion, "%d is invalid", ipVer)
	}
	p.IPVersion = IPVersion(ipVer)

	packetType := (byte0 & typeMask) >> 4
	if !(packetType <= 15) {
		return pkgerrors.Wrapf(ErrInvalidType, "%d is invalid", packetType)
	}
	p.Type = Type(packetType)

	p.Src = net.IPv4(in[1], in[2], in[3], in[4])

	p.SrcPort = byteSliceToUint16(in[5:7])

	p.Dst = net.IPv4(in[7], in[8], in[9], in[10])

	p.DstPort = byteSliceToUint16(in[11:13])

	p.DataLength = byteSliceToUint16(in[13:15])

	p.Data = in[15:]

	return nil
}

func Marshal(p Packet) ([]byte, error) {
	// TODO(charlie0129): check validity
	// TODO(charlie0129): normalize to big endian
	byte0_1 := []byte{byte(int(p.PacketVersion) + int(p.IPVersion)<<3 + int(p.Type)<<4)}

	byte1_5 := p.Src.To4()

	byte5_7 := uint16ToByteSlice(p.SrcPort)

	byte7_11 := p.Dst.To4()

	byte11_13 := uint16ToByteSlice(p.DstPort)

	byte13_15 := uint16ToByteSlice(p.DataLength)

	byte15_ := p.Data

	return appendSlices[byte](
		byte0_1,
		byte1_5,
		byte5_7,
		byte7_11,
		byte11_13,
		byte13_15,
		byte15_,
	), nil
}
