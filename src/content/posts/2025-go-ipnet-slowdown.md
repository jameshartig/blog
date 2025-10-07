---
title: 'Golang IP Network Slowdown'
pubDate: '2025-10-06'
description: 'IPNet slowdown with Golang 1.21'
---

Go's net package conaints the [IPNet](https://pkg.go.dev/net#IPNet) struct
to represent an IP network contaning an [IP](https://pkg.go.dev/net#IP) and an
[IPMask](https://pkg.go.dev/net#IPMask). So a network like `192.168.0.1/24` would
be stored as `192.168.0.1` and `ffffff00`. The struct mainly offers a helper
function `Contains(IP) bool` indicating if the sent IP is contained within the
network. You can use `ParseCIDR` to parse CIDR notation into an `IPNet` struct.

In Go 1.21 the `ParseIP` method was [changed](https://go-review.googlesource.com/c/go/+/463987)
(and later [documented](https://go-review.googlesource.com/c/go/+/598076)) to
always return a 16-byte IP, representing IPv4 addresses as IPv4-mapped IPv6
addresses. The net package treats IPv4-mapped IPv6 addresses and IPv4 addresses
as equilivant so the change shouldn't have changed behavior.

However, `Contains` always calls `To4` on the sent IP

```go
if x := ip.To4(); x != nil {
  ip = x
}
```

which previously did nothing but now ends up slicing the IP whenver it's an
IPv4-mapped IPv6 address (now all the time for IPv4 addresses).

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
BenchmarkContainsV4-16                  88221022                15.06 ns/op
BenchmarkContainsV4Mapped-16            50024178                21.05 ns/op
```

until you need to check if an IP is contained in a block list with 1000 entries.

```
BenchmarkContainsV4List-16                148334              7111 ns/op
BenchmarkContainsV4MappedList-16           90250             13092 ns/op
```

The code for this post can be found in
[2025-go-ipnet-slowdown](https://github.com/jameshartig/blog/tree/main/public/code/2025-go-ipnet-slowdown).
