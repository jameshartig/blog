package go_ipnet_slowdown

import (
	"bufio"
	"bytes"
	_ "embed"
	"net"
	"testing"
)

//go:embed netlist.txt
var netlist []byte

var list = make([]*net.IPNet, 0, 500)

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
		list = append(list, n)
	}
}

func BenchmarkContainsV4(b *testing.B) {
	ipNet := net.IPNet{
		IP:   net.IPv4(192, 168, 0, 1).To4(),
		Mask: net.CIDRMask(24, 32),
	}
	ip2 := net.IPv4(192, 168, 0, 2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ipNet.Contains(ip2)
	}
}

func BenchmarkContainsV4Mapped(b *testing.B) {
	ipNet := net.IPNet{
		IP:   net.IPv4(192, 168, 0, 1),
		Mask: net.CIDRMask(24, 32),
	}
	ip2 := net.IPv4(192, 168, 0, 2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ipNet.Contains(ip2)
	}
}

func BenchmarkContainsV4List(b *testing.B) {
	ip2 := net.IPv4(192, 168, 0, 2).To4()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, ipNet := range list {
			ipNet.Contains(ip2)
		}
	}
}

func BenchmarkContainsV4MappedList(b *testing.B) {
	ip2 := net.IPv4(192, 168, 0, 2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, ipNet := range list {
			ipNet.Contains(ip2)
		}
	}
}
