package packet

import (
	"net"
)

type Packet struct {
	// Byte [0,1)
	// [0-3)
	PacketVersion Version
	// [3-4)
	IPVersion IPVersion
	// [4-8)
	Type Type

	// Byte [1,5)
	// IPVersion=IPv4 [8-40)
	Src net.IP

	// Byte [5,7)
	// [40-56)
	SrcPort uint16

	// Byte [7,11)
	// IPVersion=IPv4 [56-88)
	Dst net.IP

	// Byte [11,13)
	// [88-104)
	DstPort uint16

	// Byte [13, 15)
	// [104-120)
	DataLength uint16

	// Byte [15,...)
	// [120-120+DataLength)
	Data []byte
}

type Version uint

const (
	Version0 = iota
	Version1
	Version2
	Version3
	Version4
	Version5
	Version6
	Version7
)

type IPVersion uint

const (
	IPv4 = 0x0
	IPv6 = 0x1
)

type Type uint

const (
	DataTransfer = iota
)
