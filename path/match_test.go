package path

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		in      string
		want    bool
	}{
		{name: "literal", pattern: "abc", in: "abc", want: true},
		{name: "literal mismatch", pattern: "abc", in: "abd", want: false},
		{name: "name longer than pattern", pattern: "abc", in: "abcd", want: false},
		{name: "name shorter than pattern", pattern: "abcd", in: "abc", want: false},
		{name: "empty pattern and name", pattern: "", in: "", want: true},
		{name: "empty pattern", pattern: "", in: "a", want: false},
		{name: "star matches empty name", pattern: "*", in: "", want: true},
		{name: "star matches anything", pattern: "*", in: "abc", want: true},
		{name: "trailing star", pattern: "a*", in: "abc", want: true},
		{name: "trailing star matches nothing", pattern: "a*", in: "a", want: true},
		{name: "leading star", pattern: "*c", in: "abc", want: true},
		{name: "leading star mismatch", pattern: "*c", in: "abd", want: false},
		{name: "star in the middle", pattern: "a*c", in: "abbbc", want: true},
		{name: "star in the middle matches nothing", pattern: "a*b", in: "ab", want: true},
		{name: "several stars", pattern: "a*b*c", in: "aXbYc", want: true},
		{name: "several stars mismatch", pattern: "a*b*c", in: "ab", want: false},
		{name: "question mark", pattern: "a?c", in: "abc", want: true},
		{name: "question mark needs a character", pattern: "a?c", in: "ac", want: false},
		{name: "question mark matches a multibyte character", pattern: "?", in: "é", want: true},
		{name: "character class", pattern: "[abc]", in: "b", want: true},
		{name: "character class mismatch", pattern: "[abc]", in: "d", want: false},
		{name: "range", pattern: "[a-c]x", in: "bx", want: true},
		{name: "range mismatch", pattern: "[a-c]x", in: "dx", want: false},
		{name: "several ranges", pattern: "[a-cx-z]", in: "y", want: true},
		{name: "negated class", pattern: "[^a-c]x", in: "dx", want: true},
		{name: "negated class mismatch", pattern: "[^a-c]x", in: "bx", want: false},
		{name: "range with a multibyte character", pattern: "[à-ü]", in: "ö", want: true},
		{name: "class followed by star", pattern: "[a-c]*", in: "bxx", want: true},
		{name: "star inside a class is literal", pattern: "[*]x", in: "*x", want: true},
		{name: "escaped star", pattern: `\*`, in: "*", want: true},
		{name: "escaped star mismatch", pattern: `\*`, in: "a", want: false},
		{name: "escaped bracket in a class", pattern: `[\]]`, in: "]", want: true},
		{name: "escaped dash in a class", pattern: `[a\-c]`, in: "-", want: true},
		{name: "escaped dash is not a range", pattern: `[a\-c]`, in: "b", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Match(tt.pattern, tt.in)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMatchBadPattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
	}{
		{name: "unclosed class", pattern: "["},
		{name: "unclosed class with characters", pattern: "[a"},
		{name: "unclosed range", pattern: "[a-"},
		{name: "empty class", pattern: "[]"},
		{name: "range without start", pattern: "[-a]"},
		{name: "range without end", pattern: "[a-]"},
		{name: "trailing backslash", pattern: `\`},
		{name: "trailing backslash in a class", pattern: `[\`},
		{name: "bad pattern after a star", pattern: "*["},
		{name: "bad pattern after a match", pattern: "a["},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Match(tt.pattern, "ab")
			assert.ErrorIs(t, err, ErrBadPattern)
			assert.False(t, got)
		})
	}
}
