package ipv4

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFixedSetContainsPrefix(t *testing.T) {
	s := NewSet()
	s.Insert(_p("10.0.0.0/16"))
	assert.True(t, s.FixedSet().Contains(_p("10.0.0.0/24")))
	assert.True(t, s.FixedSet().Contains(_p("10.0.30.0/27")))
	assert.True(t, s.FixedSet().Contains(_p("10.0.128.0/17")))
	assert.False(t, s.FixedSet().Contains(_p("10.224.0.0/24")))
	assert.False(t, s.FixedSet().Contains(_p("10.1.30.0/27")))
	assert.False(t, s.FixedSet().Contains(_p("10.0.128.0/15")))
}

func TestWalkRanges(t *testing.T) {
	tests := []struct {
		description string
		prefixes    []SetI
		ranges      []Range
	}{
		{
			description: "empty",
			prefixes:    []SetI{},
			ranges:      []Range{},
		}, {
			description: "simple prefix",
			prefixes: []SetI{
				_p("203.0.113.0/24"),
			},
			ranges: []Range{
				_p("203.0.113.0/24").Range(),
			},
		}, {
			description: "adjacent prefixes",
			prefixes: []SetI{
				_p("203.0.113.0/25"),
				_p("203.0.113.128/26"),
			},
			ranges: []Range{
				_r(_a("203.0.113.0"), _a("203.0.113.191")),
			},
		}, {
			description: "disjoint prefixes",
			prefixes: []SetI{
				_p("203.0.113.0/25"),
				_p("203.0.113.192/26"),
			},
			ranges: []Range{
				_r(_a("203.0.113.0"), _a("203.0.113.127")),
				_r(_a("203.0.113.192"), _a("203.0.113.255")),
			},
		}, {
			// This is the reverse of the complicated test from range_test.go
			description: "complicated",
			prefixes: []SetI{
				_p("7.0.37.17/32"),
				_p("7.0.37.18/31"),
				_p("7.0.37.20/30"),
				_p("7.0.37.24/29"),
				_p("7.0.37.32/27"),
				_p("7.0.37.64/26"),
				_p("7.0.37.128/25"),
				_p("7.0.38.0/23"),
				_p("7.0.40.0/21"),
				_p("7.0.48.0/20"),
				_p("7.0.64.0/18"),
				_p("7.0.128.0/17"),
				_p("7.1.0.0/16"),
				_p("7.2.0.0/15"),
				_p("7.4.0.0/14"),
				_p("7.8.0.0/13"),
				_p("7.16.0.0/12"),
				_p("7.32.0.0/11"),
				_p("7.64.0.0/10"),
				_p("7.128.0.0/9"),
				_p("8.0.0.0/6"),
				_p("12.0.0.0/8"),
				_p("13.0.0.0/13"),
				_p("13.8.0.0/17"),
				_p("13.8.128.0/18"),
				_p("13.8.192.0/20"),
				_p("13.8.208.0/21"),
				_p("13.8.216.0/22"),
				_p("13.8.220.0/23"),
				_p("13.8.222.0/26"),
				_p("13.8.222.64/27"),
				_p("13.8.222.96/28"),
				_p("13.8.222.112/31"),
			},
			ranges: []Range{
				_r(_a("7.0.37.17"), _a("13.8.222.113")),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			t.Run("finish", func(t *testing.T) {
				s := NewFixedSet(tt.prefixes...)
				ranges := []Range{}
				finished := s.WalkRanges(func(r Range) bool {
					ranges = append(ranges, r)
					return true
				})
				assert.True(t, finished)
				require.Equal(t, len(tt.ranges), len(ranges))
				for i, r := range tt.ranges {
					assert.Equal(t, r, ranges[i])
				}
			})

			t.Run("don't finish", func(t *testing.T) {
				s := NewFixedSet(tt.prefixes...)
				if s.Size() != 0 {
					ranges := []Range{}
					finished := s.WalkRanges(func(r Range) bool {
						ranges = append(ranges, r)
						return false
					})
					assert.False(t, finished)
					assert.Equal(t, 1, len(ranges))
					assert.Equal(t, tt.ranges[0], ranges[0])
				}
			})
		})
	}
}
