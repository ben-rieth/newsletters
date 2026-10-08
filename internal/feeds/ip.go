package feeds

import "net"

var blockedIPRanges []*net.IPNet

func InitBlockedIPs() {
	blocked := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",

		// Shared address space (CGNAT) - used for internal addressing by some
		// hosting platforms and VPN overlays such as Tailscale
		"100.64.0.0/10",

		"127.0.0.0/8",

		// Link-local / cloud metadata - AWS, GCP, Azure all expose instance
		// metadata (API keys, IAM roles, etc.) at 169.254.169.254
		"169.254.0.0/16",

		"0.0.0.0/8",
		"192.0.0.0/24",

		"198.18.0.0/15",

		"224.0.0.0/4",
		"240.0.0.0/4",

		"::1/128",
		"::/128",

		"fc00::/7",

		"fe80::/10",

		"ff00::/8",
		"fec0::/10",
		"100::/64",

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
