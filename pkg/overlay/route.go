package overlay

import (
	"net"
	"runtime"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/slackhq/nebula/cidr"
	"github.com/slackhq/nebula/iputil"
)

type Route struct {
	MTU    int
	Metric int
	Cidr   *net.IPNet
	Via    *iputil.VpnIp
}

func makeRouteTree(l *logrus.Logger, routes []Route, allowMTU bool) (*cidr.Tree4, error) {
	routeTree := cidr.NewTree4()
	for _, r := range routes {
		if !allowMTU && r.MTU > 0 {
			l.WithField("route", r).Warnf("route MTU is not supported in %s", runtime.GOOS)
		}

		if r.Via != nil {
			routeTree.AddCIDR(r.Cidr, *r.Via)
		}
	}
	return routeTree, nil
}

func StringToCIDRs(str []string) ([]*net.IPNet, error) {
	var ret []*net.IPNet
	for _, r := range str {
		_, cidr, err := net.ParseCIDR(r)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot parse %s", r)
		}
		ret = append(ret, cidr)
	}
	return ret, nil
}

func ipNetToRoutes(cidrs []*net.IPNet, via net.IP) ([]Route, error) {
	var routes []Route
	for _, r := range cidrs {
		via := iputil.Ip2VpnIp(via.To4())
		_r := Route{
			MTU:    DefaultMTU,
			Metric: 0,
			Cidr:   r,
			Via:    &via,
		}
		routes = append(routes, _r)
	}
	return routes, nil
}
