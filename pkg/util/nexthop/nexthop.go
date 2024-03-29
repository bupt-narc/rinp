package nexthop

import (
	"errors"
	"net"
)

// NextHopMap is a map of CIDR to next hop IP address.
type NextHopMap map[string]*net.UDPAddr // FIXME: do not use pointers

// SetNextHop adds a CIDR to next hop IP address mapping. // TODO: make nextHop *net.UDPAddr
func (m NextHopMap) SetNextHop(cidr *net.IPNet, nextHop string) error {
	// transfer string to net.UDPAddr
	udpAddr, err := net.ResolveUDPAddr("udp4", nextHop)
	if err != nil {
		return err
	}
	m[cidr.String()] = udpAddr
	return nil
}

// SetNextHop by IP string
func (m NextHopMap) SetNextHopByString(ip, nextHop string) error {
	_, cidr, err := net.ParseCIDR(ip + "/32")
	if err != nil {
		return err
	}
	m.SetNextHop(cidr, nextHop)
	return nil
}

// GetNextHop returns the next hop IP address for the given IP address.
// TODO: change CIDR to single IP
func (m NextHopMap) GetNextHop(ip net.IP) (*net.UDPAddr, error) {
	// TODO improve performance
	for cidrStr, nextHop := range m {
		_, cidr, _ := net.ParseCIDR(cidrStr)
		if cidr.Contains(ip) {
			return nextHop, nil
		}
	}
	return nil, errors.New("no next hop found for" + ip.String())
}

// GetNextHopByString returns the next hop IP address for the given IP address.
func (m NextHopMap) GetNextHopByString(ip string) (*net.UDPAddr, error) {
	IP := net.ParseIP(ip)
	if IP == nil {
		return nil, errors.New("invalid IP address")
	}
	return m.GetNextHop(IP)
}

// RemoveNextHop removes a CIDR to next hop IP address mapping.
func (m NextHopMap) RemoveNextHop(cidr *net.IPNet) {
	delete(m, cidr.String())
}

func NewNextHopMap() NextHopMap {
	return make(NextHopMap)
}
