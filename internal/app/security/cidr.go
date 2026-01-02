package security

import (
	"fmt"
	"net"
	"net/http"
)

// CIDRAllowlist holds parsed CIDR ranges.
type CIDRAllowlist struct {
	nets []*net.IPNet
}

// NewCIDRAllowlist parses the CIDR strings. Empty list means deny all.
func NewCIDRAllowlist(cidrs []string) (*CIDRAllowlist, error) {
	list := make([]*net.IPNet, 0, len(cidrs))
	for _, cidr := range cidrs {
		_, netw, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, fmt.Errorf("parse cidr %s: %w", cidr, err)
		}
		list = append(list, netw)
	}
	return &CIDRAllowlist{nets: list}, nil
}

// Middleware denies requests outside the configured CIDRs.
func (c *CIDRAllowlist) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "invalid remote addr", http.StatusForbidden)
			return
		}
		ip := net.ParseIP(host)
		if ip == nil {
			http.Error(w, "invalid ip", http.StatusForbidden)
			return
		}
		if len(c.nets) == 0 {
			http.Error(w, "access denied", http.StatusForbidden)
			return
		}
		for _, netw := range c.nets {
			if netw.Contains(ip) {
				next.ServeHTTP(w, r)
				return
			}
		}
		http.Error(w, "access denied", http.StatusForbidden)
	})
}
