package feeds

import "net"

var blockedIPRanges []*net.IPNet

func InitBlockedIPs() {
	blocked := []string{
        // Private network ranges (RFC 1918) - internal corporate/home networks
        "10.0.0.0/8",
        "172.16.0.0/12",
        "192.168.0.0/16",

        // Loopback - resolves to the server itself (e.g. localhost)
        "127.0.0.0/8",

        // Link-local / cloud metadata - AWS, GCP, Azure all expose instance
        // metadata (API keys, IAM roles, etc.) at 169.254.169.254
        "169.254.0.0/16",

        // "This" network - source-only, should never be a destination
        "0.0.0.0/8",

        // IPv6 loopback - equivalent of 127.0.0.1
        "::1/128",

        // IPv6 unique local - equivalent of RFC 1918 private ranges
        "fc00::/7",

        // IPv6 link-local - equivalent of 169.254.0.0/16
        "fe80::/10",
    }

	for _, cidr := range blocked {
		_, network, _ := net.ParseCIDR(cidr)
		blockedIPRanges = append(blockedIPRanges, network)
	}
}

func isSafeIP(ip net.IP) bool {
	for _, network := range blockedIPRanges {
		if network.Contains(ip) {
			return false
		}
	}
	return true
}
