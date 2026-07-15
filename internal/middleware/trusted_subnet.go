package middleware

import (
	"fmt"
	"net"
	"net/http"
)

func TrustedSubnetMiddleware(trustedSubnet string) (func(http.Handler) http.Handler, error) {
	var subnet *net.IPNet

	if trustedSubnet != "" {
		_, parsed, err := net.ParseCIDR(trustedSubnet)
		if err != nil {
			return nil, fmt.Errorf("parse trusted subnet %q: %w", trustedSubnet, err)
		}
		subnet = parsed
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if subnet == nil {
				next.ServeHTTP(w, r)
				return
			}

			ip := net.ParseIP(r.Header.Get("X-Real-IP"))
			if ip == nil || !subnet.Contains(ip) {
				http.Error(w, "ip address is not in trusted subnet", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}, nil
}
