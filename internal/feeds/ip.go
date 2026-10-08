package feeds

import (
	"net"
	"net/netip"
)

var blockedIPRanges = []netip.Prefix{
	netip.MustParsePrefix("10.0.0.0/8"),
	netip.MustParsePrefix("172.16.0.0/12"),
	netip.MustParsePrefix("192.168.0.0/16"),

	// Shared address space (CGNAT) - used for internal addressing by some
	// hosting platforms and VPN overlays such as Tailscale
	netip.MustParsePrefix("100.64.0.0/10"),

	netip.MustParsePrefix("127.0.0.0/8"),

	// Link-local / cloud metadata - AWS, GCP, Azure all expose instance
	// metadata (API keys, IAM roles, etc.) at 169.254.169.254
	netip.MustParsePrefix("169.254.0.0/16"),

	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("192.0.0.0/24"),

	netip.MustParsePrefix("198.18.0.0/15"),

	netip.MustParsePrefix("224.0.0.0/4"),
	netip.MustParsePrefix("240.0.0.0/4"),

	netip.MustParsePrefix("::1/128"),
	netip.MustParsePrefix("::/128"),
	netip.MustParsePrefix("::/96"),

	netip.MustParsePrefix("fc00::/7"),

	netip.MustParsePrefix("fe80::/10"),

	netip.MustParsePrefix("ff00::/8"),
	netip.MustParsePrefix("fec0::/10"),
	netip.MustParsePrefix("100::/64"),

	// IPv6 forms that carry an IPv4 address inside them (NAT64, 6to4,
	// Teredo). A gateway can translate them back to an internal IPv4 host,
	// and no real feed is served from them.
	netip.MustParsePrefix("64:ff9b::/96"),
	netip.MustParsePrefix("64:ff9b:1::/48"),
	netip.MustParsePrefix("2002::/16"),
	netip.MustParsePrefix("2001::/32"),
}

func isSafeIP(ip net.IP) bool {
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return false
	}
	addr = addr.Unmap()

	if !addr.IsGlobalUnicast() || addr.IsPrivate() {
		return false
	}

	for _, network := range blockedIPRanges {
		if network.Contains(addr) {
			return false
		}
	}
	return true
}
