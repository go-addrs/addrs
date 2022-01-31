An alternative representation for IP addresses than the one in the [go net
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
opaque, immutable, comparable, small, allocation free, usable as a table key, and
is as interoperable as possible.

// TODO address ipv6 scopes
// TODO address situations that need to return either IPv4 or IPv6 such as a DNS resolver.

## IP Tables

This is a data structure that tables IP addresses to arbitrary `interface{}`
values. It supports the constant-time basic table operations: insert, get, and
remove. It also supports O(n) iteration over all prefix/value pairs in
lexigraphical order of prefixes.

When a table is created, you choose whether the prefixes will be IPv4 (4-byte
representation only) or IPv6 (16-byte) addresses. The two families cannot be
mixed in the same table instance. This is consistent with this library's stance on
not conflating IPv4 with 16-byte IPv4 in IPv6 representation.

Since this data structure was specifically designed to use IP addresses as keys,
it supports a couple of other cool operations.

First, it can efficiently perform a longest prefix match when an exact match is
not available. This operation has the same O(1) efficiency as an exact match.

Second, it supports aggregation of key/values while iterating on the fly. This
has nearly the same O(n) efficiency \*\* as iterating without aggregating. The
rules of aggregation are as follows:

1. The values stored must be comparable. Prefixes get aggregated only where
   their values compare equal.
2. The set of key/value pairs visited is the minimal-size set such that any
   longest prefix match against the aggregated set will always return the same
   value as the same match against the non-aggregated set.
3. The aggregated and non-aggregated sets of prefixes may be disjoint.

Aggregation can be useful, for example, to minimize the number of prefixes
needed to install into a router's datapath to guarantee that all of the next
hops are correct. In general, though, routing protocols should be careful when
passing aggregated routes to neighbors as this will likely lead to poor
comparisions by neighboring routers who receive routes aggregated differently
from different peers.

A future enhancement could efficiently compute the difference in the aggregated
set when inserting or removing elements so that the entire set doesn't need to
be iterated after each mutation. Since the aggregated set of prefixes is
disjoint from the original, either operation could result in both adding and
removing key/value pairs. This makes it tricky but it should be possible.

As a simple example, consider the following key/value pairs inserted into a table.

- 10.224.24.2/31 / true
- 10.224.24.0/32 / true
- 10.224.24.1/32 / true

When iterating over the aggregated set, only the following key/value pair will
be visited.

- 10.224.24.0/30 / true

A slightly more complex example shows how value comparison comes into play.

- 10.224.24.0/30 / true
- 10.224.24.0/31 / false
- 10.224.24.1/32 / true
- 10.224.24.0/32 / false

Iterating the aggregated set:

- 10.224.24.0/30 / true
- 10.224.24.0/31 / false
- 10.224.24.1/32 / true

A more complex example where all values are the same (so they aren't shown)

- 172.21.0.0/20
- 192.68.27.0/25
- 192.168.26.128/25
- 10.224.24.0/32
- 192.68.24.0/24
- 172.16.0.0/12
- 192.68.26.0/24
- 10.224.24.0/30
- 192.168.24.0/24
- 192.168.25.0/24
- 192.168.26.0/25
- 192.68.25.0/24
- 192.168.27.0/24
- 172.20.128.0/19
- 192.68.27.128/25

The aggregrated set is as follows:

- 10.224.24.0/30
- 172.16.0.0/12
- 192.68.24.0/22
- 192.168.24.0/22

\*\* There is one complication that may throw its efficiency slightly off of
     O(n) but I haven't analyzed it yet to be sure. It should be pretty close.

[go net library]: https://golang.org/pkg/net/
[go-netaddr]: https://gopkg.in/netaddr.v1
[inet.af/netaddr]: https://pkg.go.dev/inet.af/netaddr
