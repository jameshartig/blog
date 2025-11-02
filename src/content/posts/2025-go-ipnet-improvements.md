---
title: 'Golang Network Contains Improvements'
pubDate: '2025-10-06'
description: 'IPNet Contains slowdown with Golang 1.21'
---

Go's net package contains the [IPNet](https://pkg.go.dev/net#IPNet) struct
to represent an IP network containing an [IP](https://pkg.go.dev/net#IP) and an
[IPMask](https://pkg.go.dev/net#IPMask). For example, a network like
`192.168.0.1/24` would be stored as `192.168.0.1` and `ffffff00`. The struct
mainly offers a helper function `Contains(IP) bool`, that indicates whether a
given IP is contained within the network. You can use `ParseCIDR` to parse CIDR
notation into an `IPNet` struct.

In Go 1.21, the `ParseIP` method was [changed](https://go-review.googlesource.com/c/go/+/463987)
(and later [documented](https://go-review.googlesource.com/c/go/+/598076)) to
always return a 16-byte IP, representing IPv4 addresses as IPv4-mapped IPv6
addresses. The net package treats IPv4-mapped IPv6 addresses and IPv4 addresses
as equivalent, so this change should not have altered behavior.

However, `Contains` always calls `To4` on the provided IP:

```go
if x := ip.To4(); x != nil {
  ip = x
}
```

This call previously did nothing for IPv4 addresses, but now it ends up slicing
the IP whenever it's an IPv4-mapped IPv6 address (which, after the 1.21 change,
is all the time for IPv4 addresses).

```go
func (ip IP) To4() IP {
	if len(ip) == IPv4len {
		return ip
	}
	if len(ip) == IPv6len &&
		isZeros(ip[0:10]) &&
		ip[10] == 0xff &&
		ip[11] == 0xff {
		return ip[12:16]
	}
	return nil
}
```

The conversion from IPv4-mapped to IPv4 is only a couple nanoseconds slower

```
BenchmarkContainsV4-16                    88157508             12.51 ns/op
BenchmarkContainsV4Mapped-16              64967758             20.06 ns/op
BenchmarkContainsV6-16                    89194792             12.97 ns/op
```

which is insignificant unless you're checking if an IP is contained against a
list of 1,000 networks.

```
BenchmarkContainsV4List-16                  148334              7111 ns/op
BenchmarkContainsV4MappedList-16             90250             13092 ns/op
BenchmarkContainsV6List-16                  153919              7656 ns/op
```

This was discovered during an investigation into an increase in CPU usage
affecting certain servers in our fleet. On some servers, `Contains` accounted for
more than 30% of CPU time and a 7x increase in time spent running the garbage
collector. The difference between servers was related to the proportion of IPv4
vs IPv6 addresses that the server was handling.

## Solution 1: Custom Contains

After we narrowed the problem down to the `To4` method my first attempt at a
solution was to write a custom function that checked for IPv4-mapped IPv6
addresses and handled them separately by checking the mask against the last 4
bytes without reslicing the IP. This solution reduced the time by more than 50%.

```go
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
```

```
BenchmarkCustomContainsV4List-16             377374            3093 ns/op
BenchmarkCustomContainsV4MappedList-16       183717            5904 ns/op
BenchmarkCustomContainsV6-16                 174031            6102 ns/op
```

This restored stability to the service and reduced the CPU usage, but there was
still a large discrepancy between servers, and as the list of networks we checked
against grew, the difference became more pronounced.

## Solution 2: Optimized Lookups

As I worked to improve the performance further I tried several things. First, I
should store the IPv4 and IPv6 networks separately and only check against the
relevant list. Second, I could swap `bytes.HasPrefix` for a string comparison
when determining if an IP is an IPv4-mapped IPv6 address. Finally, I can use the
prefix from the network as a key in a map to further reduce the number of
comparisons needed.

This resulted in something similar to:

```go
type IPNetSet struct {
	m4 map[string][]*net.IPNet
	m6 map[string][]*net.IPNet
}

// Find returns the first IPNet that contains the given ip.
func (s *IPNetSet) Find(ip net.IP) (*net.IPNet, bool) {
	switch {
	case len(ip) == net.IPv4len:
		for _, e := range s.m4[string(ip[:1])] {
			if Contains(e, ip) {
				return e, true
			}
		}
	case len(ip) == net.IPv6len && string(ip[:12]) == v4InV6Prefix:
		ip = ip[12:]
		for _, e := range s.m4[string(ip[:1])] {
			if Contains(e, ip) {
				return e, true
			}
		}
	case len(ip) == net.IPv6len:
		for _, e := range s.m6[string(ip[:2])] {
			if Contains(e, ip) {
				return e, true
			}
		}
	}
	return nil, false
}
```

With this solution, the time was reduced by almost 100x. The IP network
lookups now barely register in CPU usage and we can handle orders of magnitude
more networks if we had to. Additionally, there's almost no difference between
the different forms of IPv4 IPs.

```
BenchmarkMapContainsV4List-16               41507410         28.53 ns/op
BenchmarkMapContainsV4MappedList-16         43106853         28.75 ns/op
BenchmarkMapContainsV6-16                   33591614         32.12 ns/op
```

## Further Optimizations

Currently, the map key only contains the first byte of IPv4 and the first 2 bytes
of IPv6 networks. The sweet spot largely depends on the distribution of the
networks and what your maximum mask is. I'll likely explore tweaking these
further. But for now, the performance is good enough that I have more important
things to focus on.

The rest of the codebase used `net.IP` and switching everything over to the new
`net/netip` package would've been more work than I was willing to do at the time.
I'll be exploring this further in the future as we move to `netip` in general. I
did benchmark the `netip.Prefix` method and it was much faster than `net.IPNet`
along with no difference between the different IP versions.

```
BenchmarkNetIPContainsV4List-16             275042          4318 ns/op
BenchmarkNetIPContainsV4MappedList-16       269730          4592 ns/op
BenchmarkNetIPContainsV6List-16             239674          4715 ns/op
```

_The code and benchmarks above can be found in
[2025-go-ipnet-improvements](https://github.com/jameshartig/blog/tree/main/public/code/2025-go-ipnet-improvements)._
