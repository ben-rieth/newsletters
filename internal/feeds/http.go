package feeds

import (
	"fmt"
	"net"
	"net/http"
	"syscall"
	"time"
)

func safeDialer() *net.Dialer {
	return &net.Dialer{
		Timeout: 5 * time.Second,
		Control: func(network, address string, c syscall.RawConn) error {
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return err
			}

			ip := net.ParseIP(host)
			if ip == nil {
				return fmt.Errorf("Could not parse IP address")
			}

			if !isSafeIP(ip) {
				return fmt.Errorf("IP address is on block list")
			}

			return nil
		},
	}
}

func newSafeFeedClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DialContext: safeDialer().DialContext,
		},
	}
}
