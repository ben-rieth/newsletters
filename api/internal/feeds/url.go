package feeds

import (
	"context"
	"errors"
	"net"
	"net/url"
)

var invalidUrlError = errors.New("Invalid URL provided")
var httpsError = errors.New("Only HTTPS URLs are supported")
var hostResolutionError = errors.New("Could not resolve the host")
var invalidIPError = errors.New("Host resolves to an invalid IP address")

func IsSafeFeedUrl(rawUrl string) error {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return invalidUrlError
	}

	if parsedUrl.Scheme != "https" {
		return httpsError
	}

	hostname := parsedUrl.Hostname()

	addrs, err := net.DefaultResolver.LookupIPAddr(context.Background(), hostname)
	if err != nil {
		return hostResolutionError
	}

	for _, addr := range addrs {
		if !isSafeIP(addr.IP) {
			return invalidIPError
		}
	}

	return nil
}
