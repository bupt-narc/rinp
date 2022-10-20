package auth

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomIP(t *testing.T) {
	cases := map[string]struct {
		CIDR string
	}{
		"/24": {
			CIDR: "10.10.10.0/24",
		},
		"/20": {
			CIDR: "10.10.12.23/20",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			for i := 0; i < 100; i++ {
				_, cidr, _ := net.ParseCIDR(c.CIDR)
				ip := RandomIP(cidr)
				assert.True(t, cidr.Contains(ip))
			}
		})
	}
}
