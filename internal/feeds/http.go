package feeds

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"syscall"
	"time"
)

const maxFeedRedirects = 10

var errInsecureRedirect = errors.New("Feed redirected to a non-HTTPS URL")

func safeDialer() *net.Dialer {
	return &net.Dialer{
		Timeout: 5 * time.Second,
		Control: func(network, address string, c syscall.RawConn) error {
			host, port, err := net.SplitHostPort(address)
			if err != nil {
				return err
			}

			if !allowedFeedPorts[port] {
				return fmt.Errorf("Port %s is not allowed", port)
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

// Feed URLs are checked for HTTPS up front, but Go follows redirects to any
// scheme by default, so a feed could otherwise quietly downgrade to plain HTTP.
func checkFeedRedirect(req *http.Request, via []*http.Request) error {
	if req.URL.Scheme != "https" {
		return errInsecureRedirect
	}

	if len(via) >= maxFeedRedirects {
		return fmt.Errorf("stopped after %d redirects", maxFeedRedirects)
	}

	return nil
}

func newSafeFeedClient() *http.Client {
	return &http.Client{
		Timeout:       10 * time.Second,
		CheckRedirect: checkFeedRedirect,
		Transport: &http.Transport{
			DialContext: safeDialer().DialContext,
		},
	}
}
