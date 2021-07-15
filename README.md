An alternative representation for IPv4 addresses than the one in the [go net
library]. After authoring [go-netaddr] -- aimed at addressing some frustrations
with the net package -- and seeing other similar packages, such as
[inet.af/netaddr] started to address other frustrations, I realized that it
still wasn't quite right.

This package's position is that IPv4 and IPv6 are two different protocols. When
you have an address in either one, you always know which one it is. There is no
16-byte representation of an IPv4 address. That myth leads to confusion. There
is a way to project IPv4 addresses into IPv6 but let's be honest; it isn't very
useful.

If you put a packet analyzer on the wire and capture packets, you may find a
mixture of ARP/IPv4 and IPv6 packets but the difference between them is always
clear.

 - the ethertype in the L2 header is `0x0800` for IPv4 and `0x86DD` for IPv6
 - you will find a 4-byte address in an IPv4 packet and 16-byte in IPv6
 - the software stacks handling the two are more different than they are the same

Conflating the two protocols by creating a single address type to represent
both is the wrong thing to do in the long run.

Beyond that, since this package creates an all new IP address type, it tries to
learn from other examples such as [inet.af/netaddr] to make a type that is
opaque, immutable, comparable, small, allocation free, usable as a map key, and
is as interoperable as possible.

// TODO address ipv6 scopes
// TODO address situations that need to return either IPv4 or IPv6 such as a DNS resolver.

[go net library]: https://golang.org/pkg/net/
[go-netaddr]: https://gopkg.in/netaddr.v1
[inet.af/netaddr]: https://pkg.go.dev/inet.af/netaddr
