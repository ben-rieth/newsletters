package feeds

import "net"

var blockedIPRanges []*net.IPNet

func InitBlockedIPs() {
	blocked := []string{
		// Private network ranges (RFC 1918) - internal corporate/home networks
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",

		// Shared address space (CGNAT) - used for internal addressing by some
		// hosting platforms and VPN overlays such as Tailscale
		"100.64.0.0/10",

		// Loopback - resolves to the server itself (e.g. localhost)
		"127.0.0.0/8",

		// Link-local / cloud metadata - AWS, GCP, Azure all expose instance
		// metadata (API keys, IAM roles, etc.) at 169.254.169.254
		"169.254.0.0/16",

		// "This" network - source-only, should never be a destination
		"0.0.0.0/8",

		// Benchmarking - reserved for testing interconnect devices
		"198.18.0.0/15",

		// Multicast, plus the reserved block that ends in the limited
		// broadcast address 255.255.255.255
		"224.0.0.0/4",
		"240.0.0.0/4",

		// IPv6 loopback - equivalent of 127.0.0.1
		"::1/128",

		// IPv6 unique local - equivalent of RFC 1918 private ranges
		"fc00::/7",

		// IPv6 link-local - equivalent of 169.254.0.0/16
		"fe80::/10",

		// IPv6 multicast
		"ff00::/8",

		// IPv6 forms that carry an IPv4 address inside them (NAT64, 6to4,
		// Teredo). A gateway can translate them back to an internal IPv4 host,
		// and no real feed is served from them.
		"64:ff9b::/96",
		"64:ff9b:1::/48",
		"2002::/16",
		"2001::/32",
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
