package ipv4

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	Eights = unsafeParseAddress("8.8.8.8")
	Nines  = unsafeParseAddress("9.9.9.9")

	Ten24          = unsafeParsePrefix("10.0.0.0/24")
	TenOne24       = unsafeParsePrefix("10.0.1.0/24")
	TenTwo24       = unsafeParsePrefix("10.0.2.0/24")
	Ten24128       = unsafeParsePrefix("10.0.0.128/25")
	Ten24Router    = unsafeParseAddress("10.0.0.1")
	Ten24Broadcast = unsafeParseAddress("10.0.0.255")
)

func TestOldSetInit(t *testing.T) {
	set := Set{}

	assert.Equal(t, int64(0), set.Size())
	assert.True(t, set.isValid())
}

func TestOldSetContains(t *testing.T) {
	set := Set{}

	assert.Equal(t, int64(0), set.Size())
	assert.False(t, set.Contains(Eights))
	assert.False(t, set.Contains(Nines))
	assert.True(t, set.isValid())
}

func TestOldSetInsert(t *testing.T) {
	sb := SetBuilder{}

	sb.Insert(Nines)
	assert.Equal(t, int64(1), sb.Set().Size())
	assert.True(t, sb.Set().Contains(Nines))
	assert.False(t, sb.Set().Contains(Eights))
	sb.Insert(Eights)
	assert.Equal(t, 2, sb.trie.NumNodes())
	assert.True(t, sb.Set().Contains(Eights))
	assert.True(t, sb.Set().isValid())
}

func TestOldSetInsertPrefixwork(t *testing.T) {
	sb := SetBuilder{}

	sb.InsertPrefix(Ten24)
	assert.Equal(t, 1, sb.trie.NumNodes())
	assert.Equal(t, int64(256), sb.Set().Size())
	assert.True(t, sb.Set().ContainsPrefix(Ten24))
	assert.True(t, sb.Set().ContainsPrefix(Ten24128))
	assert.False(t, sb.Set().Contains(Nines))
	assert.False(t, sb.Set().Contains(Eights))
	assert.True(t, sb.Set().isValid())
}

func TestOldSetInsertSequential(t *testing.T) {
	sb := SetBuilder{}

	sb.Insert(unsafeParseAddress("192.168.1.0"))
	assert.Equal(t, 1, sb.trie.NumNodes())
	sb.Insert(unsafeParseAddress("192.168.1.1"))
	assert.Equal(t, 1, sb.trie.NumNodes())
	sb.Insert(unsafeParseAddress("192.168.1.2"))
	assert.Equal(t, 2, sb.trie.NumNodes())
	sb.Insert(unsafeParseAddress("192.168.1.3"))
	assert.Equal(t, 1, sb.trie.NumNodes())
	assert.Equal(t, int64(4), sb.Set().Size())

	cidr := unsafeParsePrefix("192.168.1.0/30")
	assert.True(t, sb.Set().ContainsPrefix(cidr))

	cidr = unsafeParsePrefix("192.168.1.4/31")
	sb.InsertPrefix(cidr)
	assert.Equal(t, 2, sb.trie.NumNodes())
	assert.True(t, sb.Set().ContainsPrefix(cidr))

	cidr = unsafeParsePrefix("192.168.1.6/31")
	sb.InsertPrefix(cidr)
	assert.Equal(t, 1, sb.trie.NumNodes())
	assert.True(t, sb.Set().ContainsPrefix(cidr))

	cidr = unsafeParsePrefix("192.168.1.6/31")
	sb.InsertPrefix(cidr)
	assert.Equal(t, 1, sb.trie.NumNodes())
	assert.True(t, sb.Set().ContainsPrefix(cidr))

	cidr = unsafeParsePrefix("192.168.0.240/29")
	sb.InsertPrefix(cidr)
	assert.Equal(t, 2, sb.trie.NumNodes())
	assert.True(t, sb.Set().ContainsPrefix(cidr))

	cidr = unsafeParsePrefix("192.168.0.248/29")
	sb.InsertPrefix(cidr)
	assert.Equal(t, 2, sb.trie.NumNodes())
	assert.True(t, sb.Set().ContainsPrefix(cidr))
	assert.True(t, sb.Set().isValid())
}

func TestOldSetRemove(t *testing.T) {
	sb := SetBuilder{}

	sb.InsertPrefix(Ten24)
	assert.Equal(t, 1, sb.trie.NumNodes())
	sb.RemovePrefix(Ten24128)
	assert.Equal(t, 1, sb.trie.NumNodes())
	assert.Equal(t, int64(128), sb.Set().Size())
	assert.False(t, sb.Set().ContainsPrefix(Ten24))
	assert.False(t, sb.Set().ContainsPrefix(Ten24128))
	cidr := unsafeParsePrefix("10.0.0.0/25")
	assert.True(t, sb.Set().ContainsPrefix(cidr))

	sb.Remove(Ten24Router)
	assert.Equal(t, int64(127), sb.Set().Size())
	assert.Equal(t, 7, sb.trie.NumNodes())
	assert.True(t, sb.Set().isValid())
}

func TestOldSetRemovePrefixworkBroadcast(t *testing.T) {
	sb := SetBuilder{}

	sb.InsertPrefix(Ten24)
	assert.Equal(t, 1, sb.trie.NumNodes())
	sb.Remove(Ten24.Address)
	sb.Remove(Ten24Broadcast)
	assert.Equal(t, int64(254), sb.Set().Size())
	assert.Equal(t, 14, sb.trie.NumNodes())
	assert.False(t, sb.Set().ContainsPrefix(Ten24))
	assert.False(t, sb.Set().ContainsPrefix(Ten24128))
	assert.False(t, sb.Set().Contains(Ten24Broadcast))
	assert.False(t, sb.Set().Contains(Ten24.Address))

	cidr := unsafeParsePrefix("10.0.0.128/26")
	assert.True(t, sb.Set().ContainsPrefix(cidr))
	assert.True(t, sb.Set().Contains(Ten24Router))

	sb.Remove(Ten24Router)
	assert.Equal(t, int64(253), sb.Set().Size())
	assert.Equal(t, 13, sb.trie.NumNodes())
	assert.True(t, sb.Set().isValid())
}

func TestOldSetRemoveAll(t *testing.T) {
	sb := SetBuilder{}

	sb.InsertPrefix(Ten24)
	cidr1 := unsafeParsePrefix("192.168.0.0/25")
	sb.InsertPrefix(cidr1)
	assert.Equal(t, 2, sb.trie.NumNodes())

	cidr2 := unsafeParsePrefix("0.0.0.0/0")
	sb.RemovePrefix(cidr2)
	assert.Equal(t, 0, sb.trie.NumNodes())
	assert.False(t, sb.Set().ContainsPrefix(Ten24))
	assert.False(t, sb.Set().ContainsPrefix(Ten24128))
	assert.False(t, sb.Set().ContainsPrefix(cidr1))
	assert.True(t, sb.Set().isValid())
}

func TestOldSet_RemoveTop(t *testing.T) {
	testSet := SetBuilder{}
	ip1 := unsafeParseAddress("10.0.0.1")
	ip2 := unsafeParseAddress("10.0.0.2")

	testSet.Insert(ip2) // top
	testSet.Insert(ip1) // inserted at left
	testSet.Remove(ip2) // remove top node

	assert.True(t, testSet.Set().Contains(ip1))
	assert.False(t, testSet.Set().Contains(ip2))
	assert.True(t, testSet.Set().isValid())
}

func TestOldSetInsertOverlapping(t *testing.T) {
	sb := SetBuilder{}

	sb.InsertPrefix(Ten24128)
	assert.False(t, sb.Set().ContainsPrefix(Ten24))
	assert.Equal(t, 1, sb.trie.NumNodes())
	sb.InsertPrefix(Ten24)
	assert.Equal(t, 1, sb.trie.NumNodes())
	assert.Equal(t, int64(256), sb.Set().Size())
	assert.True(t, sb.Set().ContainsPrefix(Ten24))
	assert.True(t, sb.Set().Contains(Ten24Router))
	assert.False(t, sb.Set().Contains(Eights))
	assert.False(t, sb.Set().Contains(Nines))
	assert.True(t, sb.Set().isValid())
}

func TestOldSetUnion(t *testing.T) {
	set1, set2 := SetBuilder{}, SetBuilder{}

	set1.InsertPrefix(Ten24)
	cidr := unsafeParsePrefix("192.168.0.248/29")
	set2.InsertPrefix(cidr)

	set := set1.Set().Union(set2.Set())
	assert.True(t, set.ContainsPrefix(Ten24))
	assert.True(t, set.ContainsPrefix(cidr))
	assert.True(t, set1.Set().isValid())
	assert.True(t, set2.Set().isValid())
}

func TestOldSetDifference(t *testing.T) {
	set1, set2 := SetBuilder{}, SetBuilder{}

	set1.InsertPrefix(Ten24)
	cidr := unsafeParsePrefix("192.168.0.248/29")
	set2.InsertPrefix(cidr)

	set := set1.Set().Difference(set2.Set())
	assert.True(t, set.ContainsPrefix(Ten24))
	assert.False(t, set.ContainsPrefix(cidr))
	assert.True(t, set1.Set().isValid())
	assert.True(t, set2.Set().isValid())
}

func TestOldIntersectionAinB1(t *testing.T) {
	case1 := []string{"10.0.16.0/20", "10.5.8.0/24", "10.23.224.0/23"}
	case2 := []string{"10.0.20.0/30", "10.5.8.0/29", "10.23.224.0/27"}
	output := []string{"10.23.224.0/27", "10.0.20.0/30", "10.5.8.0/29"}
	testIntersection(t, case1, case2, output)

}

func TestOldIntersectionAinB2(t *testing.T) {
	case1 := []string{"10.10.0.0/30", "10.5.8.0/29", "10.23.224.0/27"}
	case2 := []string{"10.10.0.0/20", "10.5.8.0/24", "10.23.224.0/23"}
	output := []string{"10.10.0.0/30", "10.5.8.0/29", "10.23.224.0/27"}
	testIntersection(t, case1, case2, output)
}

func TestOldIntersectionAinB3(t *testing.T) {
	case1 := []string{"10.0.5.0/24", "10.5.8.0/29", "10.23.224.0/27"}
	case2 := []string{"10.6.0.0/24", "10.9.9.0/29", "10.23.6.0/23"}
	output := []string{}
	testIntersection(t, case1, case2, output)
}

func TestOldIntersectionAinB4(t *testing.T) {
	case1 := []string{"10.23.6.0/24", "10.5.8.0/29", "10.23.224.0/27"}
	case2 := []string{"10.6.0.0/24", "10.9.9.0/29", "10.23.6.0/29"}
	output := []string{"10.23.6.0/29"}
	testIntersection(t, case1, case2, output)
}

func TestOldIntersectionAinB5(t *testing.T) {
	case1 := []string{"10.0.23.0/27", "10.0.20.0/27", "10.0.15.0/27"}
	case2 := []string{"10.0.23.0/24", "10.0.20.0/24", "10.0.15.0/24"}
	output := []string{"10.0.23.0/27", "10.0.20.0/27", "10.0.15.0/27"}
	testIntersection(t, case1, case2, output)
}

func TestOldIntersectionAinB6(t *testing.T) {
	case1 := []string{"10.0.23.0/24", "10.0.20.0/24", "10.0.15.0/24"}
	case2 := []string{"10.0.23.0/27", "10.0.20.0/27", "10.0.15.0/27"}
	output := []string{"10.0.15.0/27", "10.0.20.0/27", "10.0.23.0/27"}
	testIntersection(t, case1, case2, output)
}

func TestOldIntersectionAinB7(t *testing.T) {
	case1 := []string{"10.0.23.0/24", "10.0.20.0/24", "10.0.15.0/24"}
	case2 := []string{"10.0.14.0/27", "10.0.10.0/27", "10.0.8.0/27"}
	output := []string{}
	testIntersection(t, case1, case2, output)
}

func TestOldIntersectionAinB8(t *testing.T) {
	case1 := []string{"10.0.23.0/24", "10.0.20.0/24", "172.16.1.0/24"}
	case2 := []string{"10.0.14.0/27", "10.0.10.0/27", "172.16.1.0/28"}
	output := []string{"172.16.1.0/28"}
	testIntersection(t, case1, case2, output)
}

func TestOldIntersectionAinB9(t *testing.T) {
	case1 := []string{"10.5.8.0/29"}
	case2 := []string{"10.10.0.0/20", "10.5.8.0/24", "10.23.224.0/23"}
	output := []string{"10.5.8.0/29"}
	testIntersection(t, case1, case2, output)
}

func testIntersection(t *testing.T, input1 []string, input2 []string, output []string) {
	set1, set2, interSect := SetBuilder{}, SetBuilder{}, SetBuilder{}
	for i := 0; i < len(input1); i++ {
		cidr := unsafeParsePrefix(input1[i])
		set1.InsertPrefix(cidr)
	}
	for j := 0; j < len(input2); j++ {
		cidr := unsafeParsePrefix(input2[j])
		set2.InsertPrefix(cidr)
	}
	for k := 0; k < len(output); k++ {
		cidr := unsafeParsePrefix(output[k])
		interSect.InsertPrefix(cidr)
	}
	set := set1.Set().Intersection(set2.Set())
	assert.True(t, interSect.Set().EqualInterface(set))
	assert.True(t, set1.Set().isValid())
	assert.True(t, set2.Set().isValid())
	assert.True(t, set.isValid())

}

func TestOldSetAllocateDeallocate(t *testing.T) {
	rand.Seed(29)

	sb := SetBuilder{}

	bigNet := unsafeParsePrefix("15.1.0.0/16")
	sb.InsertPrefix(bigNet)

	assert.Equal(t, int64(65536), sb.Set().Size())

	ips := make([]Address, 0, sb.Set().Size())
	sb.Set().Iterate(func(ip Address) bool {
		ips = append(ips, ip)
		return true
	})

	allocated := SetBuilder{}
	for i := 0; i != 16384; i++ {
		allocated.Insert(ips[rand.Intn(65536)])
	}
	assert.Equal(t, int64(14500), allocated.Set().Size())
	allocated.Set().Iterate(func(ip Address) bool {
		assert.True(t, sb.Set().Contains(ip))
		return true
	})

	available := sb.Set().Difference(allocated.Set())
	assert.Equal(t, int64(51036), available.Size())
	available.Iterate(func(ip Address) bool {
		assert.True(t, sb.Set().Contains(ip))
		assert.False(t, allocated.Set().Contains(ip))
		return true
	})
	assert.Equal(t, int64(51036), available.Size())
	assert.True(t, sb.Set().isValid())
	assert.True(t, allocated.Set().isValid())
	assert.True(t, available.isValid())
}

func TestOldEqualTrivial(t *testing.T) {
	a, b := SetBuilder{}, SetBuilder{}
	assert.True(t, a.Set().EqualInterface(b.Set()))

	a.InsertPrefix(unsafeParsePrefix("10.0.0.0/24"))
	assert.False(t, a.Set().EqualInterface(b.Set()))
	assert.False(t, b.Set().EqualInterface(a.Set()))
	assert.True(t, a.Set().EqualInterface(a.Set()))
	assert.True(t, b.Set().EqualInterface(b.Set()))
	assert.True(t, a.Set().isValid())
	assert.True(t, b.Set().isValid())
}

func TestOldEqualSingleNode(t *testing.T) {
	a, b := SetBuilder{}, SetBuilder{}
	a.InsertPrefix(unsafeParsePrefix("10.0.0.0/24"))
	b.InsertPrefix(unsafeParsePrefix("10.0.0.0/24"))

	assert.True(t, a.Set().EqualInterface(b.Set()))
	assert.True(t, b.Set().EqualInterface(a.Set()))
	assert.True(t, a.Set().isValid())
	assert.True(t, b.Set().isValid())
}

func TestOldEqualAllIPv4(t *testing.T) {
	a, b, c := SetBuilder{}, SetBuilder{}, SetBuilder{}
	// Insert the entire IPv4 space into set a in one shot
	a.InsertPrefix(unsafeParsePrefix("0.0.0.0/0"))

	// Insert the entire IPv4 space piece by piece into b and c

	// This list was constructed starting with 0.0.0.0/32 and 0.0.0.1/32,
	// then adding 0.0.0.2/31, 0.0.0.4/30, ..., 128.0.0./1, and then
	// randomizing the list.
	bNets := []Prefix{
		unsafeParsePrefix("0.0.0.0/32"),
		unsafeParsePrefix("0.0.0.1/32"),
		unsafeParsePrefix("0.0.0.128/25"),
		unsafeParsePrefix("0.0.0.16/28"),
		unsafeParsePrefix("0.0.0.2/31"),
		unsafeParsePrefix("0.0.0.32/27"),
		unsafeParsePrefix("0.0.0.4/30"),
		unsafeParsePrefix("0.0.0.64/26"),
		unsafeParsePrefix("0.0.0.8/29"),
		unsafeParsePrefix("0.0.1.0/24"),
		unsafeParsePrefix("0.0.128.0/17"),
		unsafeParsePrefix("0.0.16.0/20"),
		unsafeParsePrefix("0.0.2.0/23"),
		unsafeParsePrefix("0.0.32.0/19"),
		unsafeParsePrefix("0.0.4.0/22"),
		unsafeParsePrefix("0.0.64.0/18"),
		unsafeParsePrefix("0.0.8.0/21"),
		unsafeParsePrefix("0.1.0.0/16"),
		unsafeParsePrefix("0.128.0.0/9"),
		unsafeParsePrefix("0.16.0.0/12"),
		unsafeParsePrefix("0.2.0.0/15"),
		unsafeParsePrefix("0.32.0.0/11"),
		unsafeParsePrefix("0.4.0.0/14"),
		unsafeParsePrefix("0.64.0.0/10"),
		unsafeParsePrefix("0.8.0.0/13"),
		unsafeParsePrefix("1.0.0.0/8"),
		unsafeParsePrefix("128.0.0.0/1"),
		unsafeParsePrefix("16.0.0.0/4"),
		unsafeParsePrefix("2.0.0.0/7"),
		unsafeParsePrefix("32.0.0.0/3"),
		unsafeParsePrefix("4.0.0.0/6"),
		unsafeParsePrefix("64.0.0.0/2"),
		unsafeParsePrefix("8.0.0.0/5"),
	}

	for _, n := range bNets {
		assert.False(t, a.Set().EqualInterface(b.Set()))
		assert.False(t, b.Set().EqualInterface(a.Set()))
		b.InsertPrefix(n)
		assert.False(t, b.Set().EqualInterface(c.Set()))
		assert.False(t, c.Set().EqualInterface(b.Set()))
	}

	// Constructed a different way
	cNets := []Prefix{
		unsafeParsePrefix("255.255.255.240/29"),
		unsafeParsePrefix("0.0.0.0/1"),
		unsafeParsePrefix("255.255.128.0/18"),
		unsafeParsePrefix("255.255.240.0/21"),
		unsafeParsePrefix("254.0.0.0/8"),
		unsafeParsePrefix("255.240.0.0/13"),
		unsafeParsePrefix("255.224.0.0/12"),
		unsafeParsePrefix("248.0.0.0/6"),
		unsafeParsePrefix("255.0.0.0/9"),
		unsafeParsePrefix("255.252.0.0/15"),
		unsafeParsePrefix("255.255.224.0/20"),
		unsafeParsePrefix("255.255.255.224/28"),
		unsafeParsePrefix("255.255.255.0/25"),
		unsafeParsePrefix("252.0.0.0/7"),
		unsafeParsePrefix("192.0.0.0/3"),
		unsafeParsePrefix("255.192.0.0/11"),
		unsafeParsePrefix("255.255.255.248/30"),
		unsafeParsePrefix("255.255.252.0/23"),
		unsafeParsePrefix("255.248.0.0/14"),
		unsafeParsePrefix("255.255.255.192/27"),
		unsafeParsePrefix("255.255.0.0/17"),
		unsafeParsePrefix("255.254.0.0/16"),
		unsafeParsePrefix("255.255.255.255/32"),
		unsafeParsePrefix("128.0.0.0/2"),
		unsafeParsePrefix("255.128.0.0/10"),
		unsafeParsePrefix("255.255.255.128/26"),
		unsafeParsePrefix("240.0.0.0/5"),
		unsafeParsePrefix("255.255.255.252/31"),
		unsafeParsePrefix("255.255.192.0/19"),
		unsafeParsePrefix("255.255.254.0/24"),
		unsafeParsePrefix("255.255.248.0/22"),
		unsafeParsePrefix("224.0.0.0/4"),
		unsafeParsePrefix("255.255.255.254/32"),
	}

	for _, n := range cNets {
		assert.False(t, c.Set().EqualInterface(a.Set()))
		assert.False(t, c.Set().EqualInterface(b.Set()))
		c.InsertPrefix(n)
		assert.True(t, a.Set().isValid())
		assert.True(t, b.Set().isValid())
		assert.True(t, c.Set().isValid())
	}

	// At this point, all three should have the entire IPv4 space
	assert.True(t, a.Set().EqualInterface(b.Set()))
	assert.True(t, a.Set().EqualInterface(c.Set()))
	assert.True(t, b.Set().EqualInterface(a.Set()))
	assert.True(t, b.Set().EqualInterface(c.Set()))
	assert.True(t, c.Set().EqualInterface(a.Set()))
	assert.True(t, c.Set().EqualInterface(b.Set()))
}
