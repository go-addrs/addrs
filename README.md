# Overview

`addrs` provides IP address related types and data structures for the Go
programming language with a clean and complete API and many nice features when
compared with correspanding types in the Go `net` package. The basic types are
opaque, immutable, comparable, space efficient, and defined as simple structs
that don't require extra memory allocation.

One key difference which sets this library apart from others is that it
maintains a clear distinction between IPv4 and IPv6 addresses and related types.

## Immutable Types

The types described in this section are opaque, immutable, comparable, space
efficient and do not require allocation. They can be used as map keys.

### Address

An `Address` does not store anything more than an IP address. Its size is 32
bits for IPv4 and 128 bits for IPv6 -- exactly the same size when they appear
network packet headers. However, they are not stored as bytes in network order
and cannot be directly serialized as such.

### Mask

A Msak is like an `Address` (same size) except that it must be in the format of
a string of 1 bits in the most significant positions, followed by 0 bits to fill
out the remainder of the bits.

### Prefix

A `Prefix` is an `Address` plus a `Mask`. However, the mask is stored more
efficiently as a length indicating the number of 1s.

Any Address can be used With a `Prefix`. It is not limited to 0s where the
`Mask` has 0s. If you need this, you can call `.Network()` to get another Prefix
with these bits zeroed out.

### Range

A `Range` represents the set of all addresses between a first and a last address
(inclusive) and is store efficiently as such.

One thing to note is that there is no valid representation of an empty `Range`.
The API will not return one in any case and the zero value of a `Range` has one
address in it (`Address{}` - `Address{}`).

## Mutable Types

The two most complex types in this library are mutable -- `Set` and `Table`.

These behave like Go maps in a few ways:

1. They are reference types where each instance points to a share data
   structure. These can be efficiently passed and returned to and from functions
   and methods without passing pointers.

2. They must be initialized to be modifiable. An unitialized `Set` or `Table`
   will behave like it is empty when you read from it but any attempt to modify
   it (e.g. insert entries) will result in a panic. Each type provides a factory
   function which will return a fully-initialized instance which can then be
   modified.

3. Unlike the simpler, immutable types mentioned above, The memory for `Set` and
   `Table` is allocated from the heap.

There are a few ways in which these types do not behave like a Go `map`:

1. Both types have a corresponding immutable represation. Converting between
   mutable and immutable instances is as efficient as copying a pointer. If you
   convert from mutable -> immutable -> mutable you end up with a mutable clone
   which is independent from the original.

2. It is safe to read and write concurrently. All read operations work on a
   consistent representation of the underlying datastructure which is not
   affected by concurrent writes. However, subsequent reads on the same instance
   may, depending on timing, reflect concurrent writes from other goroutines.

3. Concurrent writes *will* cause a panic.

A nice pattern to ensure consistency is to reserve writing to a single goroutine
and then send fixed snapshots of the `Set` or `Table` through channels to other
goroutines to consume it.

### Set and FixedSet

A `Set` contains any arbitrary combination of individual, distinct `Address`
values. `Address`, `Prefix`, and `Range` are similar to sets but are more
constrained. The API provides methods to convert freely between these types.

The fixed representation of a `Set` is called `FixedSet`. One can be efficiently
obtained by calling `.FixedSet()` on an `Address`, `Prefix`, `Range`, or `Set`.

The memory required to store a `Set` is proportional to the minimum number of
`Prefix`es required to exactly cover all of its `Address`es. This
proportionality is maintained as modifications occur. For example,
subsequentially inserting two equally sized, and properly aligned `Prefix`es
will result in changes to the underlying structure to represent them both with a
single `Prefix`. It can get arbitrarily complicated and this will hold true.

The above is especially important when it comes to storing IPv6 addresses. For
example, an entire `/64` `Prefix` has a massive number of distinct `Address`es
but is stored in a very small space.

### Table and FixedTable

A `Table` maps `Prefix`es to arbitrary values. They use Go generics so that any
type of value can be stored and retreived in a type-safe manner. Bits in the
`Address` part that would be masked off by 0s in the `Mask` are ignored when
using it as a key in a `Table`.

`ITable` is equivalent but does not require generics. If you do not have at
least Go version `1.18`, you can use this type to map to `interface{}`es
instead. Arbitrary types can still be stored but it is up to you to dynamically
cast them to the type you want.

At first glance, `Table` may seem similar but more restrictive than Go's `map`.
Afterall, a `Prefix` can be used as a `map` key so why is it necessary?

`Table` is more capable than Go `map` in a few very important ways besides the
ones mentioned above.

1. Walking a `Table` always orders the keys lexigraphically. Much like strings,
   shorter `Prefix`es come first followed by longer ones that it contains.
   `Prefix`es of the same length are ordered by their upper bits, up to that
   length.

2. It supports an efficient longest prefix match. When you search using a
   `Prefix`, it will return the entry whose key is closest to it (longest) yet
   still contains it.

3. It can convert itself to an aggregated form containing the minimum number of
   entries required such that any search using an `Address` as the search key
   will return the same value as the same search on the original. (This
   operation requires that the values be of a comparable type, either by using a
   type that is inherently comparable with `==` and `!=` or by implementing the
   `EqualComparable` interface.

4. It supports an efficient diff operation so that you can iterate the entries
   removed, added, or changed from one to the other. Given a large table to
   start with, if you make a small number of modifications to it and then diff
   the before and after snapshots, the diff operation efficiency will be very
   good, proportional to the changes made between the two snapshots.
