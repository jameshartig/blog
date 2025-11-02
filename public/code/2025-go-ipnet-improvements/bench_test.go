package go_ipnet_slowdown

import (
	"bufio"
	"bytes"
	_ "embed"
	"net"
	"net/netip"
	"testing"
)

//go:embed netlist.txt
var netlist []byte

var list = make([]*net.IPNet, 0, 1000)
var set IPNetSet
var netipList = make([]netip.Prefix, 0, 1000)

var ok bool

func init() {
	s := bufio.NewScanner(bytes.NewReader(netlist))
	for s.Scan() {
		line := s.Text()
		if line == "" {
			continue
		}
		_, n, err := net.ParseCIDR(line)
		if err != nil {
			panic(err)
		}
		set.Add(n)
		list = append(list, n)
		netipList = append(netipList, netip.MustParsePrefix(line))
	}
}

func BenchmarkContainsV4(b *testing.B) {
	ipNet := net.IPNet{
		IP:   net.IPv4(192, 168, 0, 0).To4(),
		Mask: net.CIDRMask(24, 32),
	}
	ip2 := net.IPv4(192, 168, 0, 2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ok = ipNet.Contains(ip2)
	}
}

func BenchmarkContainsV4Mapped(b *testing.B) {
	ipNet := net.IPNet{
		IP:   net.IPv4(192, 168, 0, 0),
		Mask: net.CIDRMask(24, 32),
	}
	ip2 := net.IPv4(192, 168, 0, 2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ok = ipNet.Contains(ip2)
	}
}

func BenchmarkContainsV6(b *testing.B) {
	ipNet := net.IPNet{
		IP:   net.IP{38, 0, 23, 0, 65, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		Mask: net.CIDRMask(48, 128),
	}
	ip2 := net.IP{38, 0, 23, 0, 65, 3, 43, 26, 156, 78, 26, 47, 179, 213, 200, 230}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ok = ipNet.Contains(ip2)
	}
}

func BenchmarkContainsV4List(b *testing.B) {
	ip2 := net.IPv4(192, 168, 0, 2).To4()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, ipNet := range list {
			ok = ipNet.Contains(ip2)
		}
	}
}

func BenchmarkContainsV4MappedList(b *testing.B) {
	ip2 := net.IPv4(192, 168, 0, 2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, ipNet := range list {
			ok = ipNet.Contains(ip2)
		}
	}
}

func BenchmarkContainsV6List(b *testing.B) {
	ip2 := net.IP{38, 0, 23, 0, 65, 3, 43, 26, 156, 78, 26, 47, 179, 213, 200, 230}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, ipNet := range list {
			ok = ipNet.Contains(ip2)
		}
	}
}

var v4InV6PrefixBytes = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff}

func Contains(ipn *net.IPNet, ip net.IP) bool {
	// explicitly check for ipv4-mapped ipv6 addresses
	if len(ip) == net.IPv6len && bytes.HasPrefix(ip, v4InV6PrefixBytes) {
		// make sure ipnet is an ipv4 address
		if len(ipn.IP) != net.IPv4len {
			return false
		}
		// we only look at bytes 12 though 16
		for i := range ipn.IP {
			if ipn.IP[i] != ip[i+12]&ipn.Mask[i] {
				return false
			}
		}
		return true
	}
	if len(ipn.IP) != len(ip) {
		return false
	}
	for i := range ipn.IP {
		if ipn.IP[i] != ip[i]&ipn.Mask[i] {
			return false
		}
	}
	return true
}

func BenchmarkCustomContainsV4List(b *testing.B) {
	ip2 := net.IPv4(192, 168, 0, 2).To4()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, ipNet := range list {
			ok = Contains(ipNet, ip2)
		}
	}
}

func BenchmarkCustomContainsV4MappedList(b *testing.B) {
	ip2 := net.IPv4(192, 168, 0, 2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, ipNet := range list {
			ok = Contains(ipNet, ip2)
		}
	}
}

func BenchmarkCustomContainsV6(b *testing.B) {
	ip2 := net.IP{38, 0, 23, 0, 65, 3, 43, 26, 156, 78, 26, 47, 179, 213, 200, 230}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, ipNet := range list {
			ok = Contains(ipNet, ip2)
		}
	}
}

const v4InV6Prefix = "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\xff\xff"

func Contains2(ipn *net.IPNet, ip net.IP) bool {
	// explicitly check for ipv4-mapped ipv6 addresses
	if len(ip) == net.IPv6len && string(ip[:12]) == v4InV6Prefix {
		// make sure ipnet is an ipv4 address
		if len(ipn.IP) != net.IPv4len {
			return false
		}
		// we only look at bytes 12 though 16
		for i := range ipn.IP {
			if ipn.IP[i] != ip[i+12]&ipn.Mask[i] {
				return false
			}
		}
		return true
	}
	if len(ipn.IP) != len(ip) {
		return false
	}
	for i := range ipn.IP {
		if ipn.IP[i] != ip[i]&ipn.Mask[i] {
			return false
		}
	}
	return true
}

type IPNetSet struct {
	m4 map[string][]*net.IPNet
	m6 map[string][]*net.IPNet
}

// Add adds the given IPNet to the set.
func (s *IPNetSet) Add(ipNet *net.IPNet) {
	ip := ipNet.IP
	switch {
	case len(ip) == net.IPv4len:
		if s.m4 == nil {
			s.m4 = make(map[string][]*net.IPNet)
		}
		s.m4[string(ip[:1])] = append(s.m4[string(ip[:1])], ipNet)
	case len(ip) == net.IPv6len && string(ip[:12]) == v4InV6Prefix:
		if s.m4 == nil {
			s.m4 = make(map[string][]*net.IPNet)
		}
		s.m4[string(ip[12:13])] = append(s.m4[string(ip[12:13])], ipNet)
	case len(ip) == net.IPv6len:
		if s.m6 == nil {
			s.m6 = make(map[string][]*net.IPNet)
		}
		s.m6[string(ip[:2])] = append(s.m6[string(ip[:2])], ipNet)
	}
}

// Find returns the first IPNet that contains the given ip.
func (s *IPNetSet) Find(ip net.IP) (*net.IPNet, bool) {
	switch {
	case len(ip) == net.IPv4len:
		for _, e := range s.m4[string(ip[:1])] {
			if Contains2(e, ip) {
				return e, true
			}
		}
	case len(ip) == net.IPv6len && string(ip[:12]) == v4InV6Prefix:
		ip = ip[12:]
		for _, e := range s.m4[string(ip[:1])] {
			if Contains2(e, ip) {
				return e, true
			}
		}
	case len(ip) == net.IPv6len:
		for _, e := range s.m6[string(ip[:2])] {
			if Contains2(e, ip) {
				return e, true
			}
		}
	}
	return nil, false
}

func BenchmarkMapContainsV4List(b *testing.B) {
	ip2 := net.IPv4(192, 168, 0, 2).To4()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, ok = set.Find(ip2)
	}
}

func BenchmarkMapContainsV4MappedList(b *testing.B) {
	ip2 := net.IPv4(192, 168, 0, 2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, ok = set.Find(ip2)
	}
}

func BenchmarkMapContainsV6(b *testing.B) {
	ip2 := net.IP{38, 0, 23, 0, 65, 3, 43, 26, 156, 78, 26, 47, 179, 213, 200, 230}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, ok = set.Find(ip2)
	}
}

func BenchmarkNetIPContainsV4List(b *testing.B) {
	ip2 := netip.AddrFrom4([4]byte{192, 168, 0, 2})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, p := range netipList {
			ok = p.Contains(ip2)
		}
	}
}

func BenchmarkNetIPContainsV4MappedList(b *testing.B) {
	ip2 := netip.AddrFrom4([4]byte{192, 168, 0, 2})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, p := range netipList {
			ok = p.Contains(ip2)
		}
	}
}

func BenchmarkNetIPContainsV6List(b *testing.B) {
	ip2 := netip.AddrFrom16([16]byte{38, 0, 23, 0, 65, 3, 43, 26, 156, 78, 26, 47, 179, 213, 200, 230})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, p := range netipList {
			ok = p.Contains(ip2)
		}
	}
}
