package feeds

import (
	"net"
	"testing"
)

func TestMain(m *testing.M) {
	// InitBlockedIPs appends, so it runs once for the whole package rather than
	// per test.
	InitBlockedIPs()
	m.Run()
}

// isSafeIP reports every address as safe until the block list is populated, so
// an empty list is the one failure mode that silently disables the guard.
func TestBlockedIPRangesAreLoaded(t *testing.T) {
	if len(blockedIPRanges) == 0 {
		t.Fatal("block list is empty, so every host would resolve as safe")
	}
}

func TestIsSafeIP(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want bool
	}{
		{name: "private class a", ip: "10.0.0.1", want: false},
		{name: "private class a broadcast", ip: "10.255.255.255", want: false},
		{name: "private class b low", ip: "172.16.0.1", want: false},
		{name: "private class b high", ip: "172.31.255.255", want: false},
		{name: "private class c", ip: "192.168.1.1", want: false},
		{name: "loopback", ip: "127.0.0.1", want: false},
		{name: "loopback range", ip: "127.1.2.3", want: false},
		{name: "cloud metadata", ip: "169.254.169.254", want: false},
		{name: "link local", ip: "169.254.0.1", want: false},
		{name: "this network", ip: "0.0.0.0", want: false},
		{name: "ipv6 loopback", ip: "::1", want: false},
		{name: "ipv6 unique local", ip: "fc00::1", want: false},
		{name: "ipv6 unique local high", ip: "fdff::1", want: false},
		{name: "ipv6 link local", ip: "fe80::1", want: false},
		{name: "cgnat low", ip: "100.64.0.1", want: false},
		{name: "cgnat high", ip: "100.127.255.254", want: false},
		{name: "benchmarking", ip: "198.18.0.1", want: false},
		{name: "benchmarking high", ip: "198.19.255.255", want: false},
		{name: "multicast", ip: "224.0.0.1", want: false},
		{name: "limited broadcast", ip: "255.255.255.255", want: false},
		{name: "reserved", ip: "240.0.0.1", want: false},
		{name: "ipv6 multicast", ip: "ff02::1", want: false},
		{name: "nat64 embedding loopback", ip: "64:ff9b::7f00:1", want: false},
		{name: "local-use nat64", ip: "64:ff9b:1::a00:1", want: false},
		{name: "6to4 embedding private", ip: "2002:a00:1::1", want: false},
		{name: "teredo", ip: "2001:0:4136:e378::1", want: false},

		{name: "public dns", ip: "1.1.1.1", want: true},
		{name: "public host", ip: "93.184.216.34", want: true},
		{name: "just outside private class b", ip: "172.32.0.1", want: true},
		{name: "just outside link local", ip: "169.255.0.1", want: true},
		{name: "public ipv6", ip: "2606:4700::1111", want: true},
		{name: "just outside cgnat", ip: "100.128.0.1", want: true},
		{name: "just below cgnat", ip: "100.63.255.255", want: true},
		{name: "just outside benchmarking", ip: "198.20.0.1", want: true},
		{name: "public ipv6 sharing teredo's first hextet", ip: "2001:4860:4860::8888", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("could not parse %q", tt.ip)
			}

			if got := isSafeIP(ip); got != tt.want {
				t.Errorf("isSafeIP(%s) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

// An IPv4-mapped address is the same host written a second way, so the block list
// has to reject it without needing its own entry.
func TestIsSafeIPRejectsIPv4MappedForms(t *testing.T) {
	for _, raw := range []string{"::ffff:169.254.169.254", "::ffff:127.0.0.1", "::ffff:10.0.0.1"} {
		ip := net.ParseIP(raw)
		if ip == nil {
			t.Fatalf("could not parse %q", raw)
		}

		if isSafeIP(ip) {
			t.Errorf("isSafeIP(%s) = true, want false", raw)
		}
	}
}

// The dialer hook is the check that actually holds: it sees the address the
// connection is about to use, after any redirect and after DNS.
func TestSafeDialerControlBlocksInternalAddresses(t *testing.T) {
	control := safeDialer().Control

	tests := []struct {
		name      string
		address   string
		wantError bool
	}{
		{name: "public address dials", address: "1.1.1.1:443", wantError: false},
		{name: "public ipv6 dials", address: "[2606:4700::1111]:443", wantError: false},
		{name: "metadata endpoint blocked", address: "169.254.169.254:80", wantError: true},
		{name: "loopback blocked", address: "127.0.0.1:8080", wantError: true},
		{name: "ipv6 loopback blocked", address: "[::1]:8080", wantError: true},
		{name: "unresolved hostname blocked", address: "internal.local:80", wantError: true},
		{name: "address without a port blocked", address: "1.1.1.1", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := control("tcp", tt.address, nil)

			if tt.wantError && err == nil {
				t.Errorf("control(%q) allowed the dial", tt.address)
			}

			if !tt.wantError && err != nil {
				t.Errorf("control(%q) blocked the dial: %v", tt.address, err)
			}
		})
	}
}
