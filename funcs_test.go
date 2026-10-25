package envsubst

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/woodpecker-ci/envsubst/path"
)

func TestFuncs(t *testing.T) {
	tests := []struct {
		name string
		fn   substituteFunc
		in   string
		args []string
		want string
	}{
		{name: "len", fn: toLen, in: "Hello World", want: "11"},
		{name: "lower", fn: toLower, in: "Hello World", want: "hello world"},
		{name: "lower first", fn: toLowerFirst, in: "HELLO WORLD", want: "hELLO WORLD"},
		{name: "lower first of empty string", fn: toLowerFirst, in: "", want: ""},
		{name: "upper", fn: toUpper, in: "Hello World", want: "HELLO WORLD"},
		{name: "upper first", fn: toUpperFirst, in: "hello world", want: "Hello world"},
		{name: "upper first of empty string", fn: toUpperFirst, in: "", want: ""},
		{name: "default uses variable value", fn: toDefault, in: "Hello World", args: []string{"Hola Mundo"}, want: "Hello World"},
		{name: "default uses default value when variable is empty", fn: toDefault, in: "", args: []string{"Hola Mundo"}, want: "Hola Mundo"},
		{name: "default concatenates args when variable is empty", fn: toDefault, in: "", args: []string{"Hola Mundo", "-Bonjour le monde", "-Halló heimur"}, want: "Hola Mundo-Bonjour le monde-Halló heimur"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.fn(tt.in, tt.args...)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSubstr(t *testing.T) {
	tests := []struct {
		name string
		in   string
		args []string
		want string
	}{
		{name: "cut from beginning to length", in: "123456789123456789", args: []string{"0", "8"}, want: "12345678"},
		{name: "cut from offset to length", in: "123456789123456789", args: []string{"1", "8"}, want: "23456789"},
		{name: "cut beginning with offset", in: "123456789123456789", args: []string{"9"}, want: "123456789"},
		{name: "ignore length out of bound", in: "123456789123456789", args: []string{"9", "50"}, want: "123456789"},
		{name: "negative offset counts from the end", in: "123456789123456789", args: []string{"-3", "2"}, want: "78"},
		{name: "negative offset beyond length starts at the beginning", in: "123456789123456789", args: []string{"-300", "3"}, want: "123"},
		{name: "offset out of bound", in: "12345678", args: []string{"9", "1"}, want: ""},
		{name: "negative length counts from the end", in: "123456789123456789", args: []string{"1", "-2"}, want: "234567891234567"},
		{name: "offset out of bound without length", in: "12345678", args: []string{"9"}, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toSubstr(tt.in, tt.args...)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// want values match bash 5.2 ${s:pos:len}
func TestSubstrNegativeLength(t *testing.T) {
	tests := []struct {
		name   string
		in     string
		pos    string
		length string
		want   string
	}{
		{name: "drop last", in: "abcdefgh", pos: "0", length: "-1", want: "abcdefg"},
		{name: "offset and negative length", in: "abcdefgh", pos: "1", length: "-2", want: "bcdef"},
		{name: "plus sign offset", in: "abcdefgh", pos: "+1", length: "-2", want: "bcdef"},
		{name: "negative offset and negative length", in: "abcdefgh", pos: "-3", length: "-1", want: "fg"},
		{name: "negative offset at the beginning", in: "abcdefgh", pos: "-8", length: "-1", want: "abcdefg"},
		{name: "end equals position at the beginning", in: "abcdefgh", pos: "0", length: "-8", want: ""},
		{name: "end equals position at the end", in: "abcdefgh", pos: "7", length: "-1", want: ""},
		{name: "end equals position with negative offset", in: "abcdefgh", pos: "-3", length: "-3", want: ""},
		{name: "minus zero is a zero length", in: "abcdefgh", pos: "1", length: "-0", want: ""},
		{name: "position past the end", in: "abcdefgh", pos: "9", length: "-1", want: ""},
		{name: "maximum position", in: "abcdefgh", pos: "9223372036854775807", length: "-1", want: ""},
		{name: "empty string", in: "", pos: "1", length: "-1", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toSubstr(tt.in, tt.pos, tt.length)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// bash fails on all of these with "substring expression < 0"
func TestSubstrNegativeLengthBeforePosition(t *testing.T) {
	tests := []struct {
		name   string
		in     string
		pos    string
		length string
	}{
		{name: "before the beginning", in: "abcdefgh", pos: "0", length: "-9"},
		{name: "before the position", in: "abcdefgh", pos: "3", length: "-9"},
		{name: "negative offset before the position", in: "abcdefgh", pos: "-3", length: "-5"},
		{name: "position at the end", in: "abcdefgh", pos: "8", length: "-1"},
		{name: "empty or unset variable", in: "", pos: "0", length: "-1"},
		{name: "must not wrap around", in: "abcdefgh", pos: "0", length: "-9223372036854775808"},
		{name: "out of int range", in: "abcdefgh", pos: "0", length: "-9223372036854775809"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toSubstr(tt.in, tt.pos, tt.length)
			assert.ErrorIs(t, err, ErrBadSubstringLength)
			assert.Empty(t, got)
		})
	}
}

func TestSubstrBadArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want error
	}{
		{name: "no arguments", args: nil, want: ErrMissingArgument},
		{name: "position is not a number", args: []string{"x"}, want: ErrBadSubstringPosition},
		{name: "position is not a number with length", args: []string{"x", "1"}, want: ErrBadSubstringPosition},
		{name: "length is not a number", args: []string{"1", "x"}, want: ErrBadSubstringLength},
		{name: "negative length is not a number", args: []string{"1", "-x"}, want: ErrBadSubstringLength},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toSubstr("abcdefgh", tt.args...)
			assert.ErrorIs(t, err, tt.want)
			assert.Empty(t, got)
		})
	}
}

func TestReplaceFuncs(t *testing.T) {
	tests := []struct {
		name string
		fn   substituteFunc
		in   string
		args []string
		want string
	}{
		{name: "replace all", fn: replaceAll, in: "a-b-c", args: []string{"-", "+"}, want: "a+b+c"},
		{name: "replace all without replacement removes", fn: replaceAll, in: "a-b-c", args: []string{"-"}, want: "abc"},
		{name: "replace all without match", fn: replaceAll, in: "a-b-c", args: []string{"x", "+"}, want: "a-b-c"},
		{name: "replace first", fn: replaceFirst, in: "a-b-c", args: []string{"-", "+"}, want: "a+b-c"},
		{name: "replace first without replacement removes", fn: replaceFirst, in: "a-b-c", args: []string{"-"}, want: "ab-c"},
		{name: "replace first without match", fn: replaceFirst, in: "a-b-c", args: []string{"x", "+"}, want: "a-b-c"},
		{name: "replace prefix", fn: replacePrefix, in: "foobar", args: []string{"foo", "x"}, want: "xbar"},
		{name: "replace prefix only at the beginning", fn: replacePrefix, in: "barfoo", args: []string{"foo", "x"}, want: "barfoo"},
		{name: "replace suffix", fn: replaceSuffix, in: "foobar", args: []string{"bar", "x"}, want: "foox"},
		{name: "replace suffix only at the end", fn: replaceSuffix, in: "barfoo", args: []string{"bar", "x"}, want: "barfoo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.fn(tt.in, tt.args...)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTrimFuncs(t *testing.T) {
	tests := []struct {
		name string
		fn   substituteFunc
		in   string
		arg  string
		want string
	}{
		{name: "shortest prefix", fn: trimShortestPrefix, in: "foo.bar.baz", arg: "*.", want: "bar.baz"},
		{name: "longest prefix", fn: trimLongestPrefix, in: "foo.bar.baz", arg: "*.", want: "baz"},
		{name: "shortest suffix", fn: trimShortestSuffix, in: "foo.bar.baz", arg: ".*", want: "foo.bar"},
		{name: "longest suffix", fn: trimLongestSuffix, in: "foo.bar.baz", arg: ".*", want: "foo"},
		{name: "literal prefix", fn: trimShortestPrefix, in: "foobar", arg: "foo", want: "bar"},
		{name: "literal suffix", fn: trimShortestSuffix, in: "foobar", arg: "bar", want: "foo"},
		{name: "character class", fn: trimLongestPrefix, in: "123abc", arg: "[0-9]*", want: ""},
		{name: "shortest prefix without match", fn: trimShortestPrefix, in: "foobar", arg: "x*", want: "foobar"},
		{name: "longest prefix without match", fn: trimLongestPrefix, in: "foobar", arg: "x*", want: "foobar"},
		{name: "shortest suffix without match", fn: trimShortestSuffix, in: "foobar", arg: "*x", want: "foobar"},
		{name: "longest suffix without match", fn: trimLongestSuffix, in: "foobar", arg: "*x", want: "foobar"},
		{name: "suffix with multibyte characters", fn: trimShortestSuffix, in: "héllo.txt", arg: ".*", want: "héllo"},
		{name: "prefix with multibyte characters", fn: trimLongestPrefix, in: "héllo.txt", arg: "*é", want: "llo.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.fn(tt.in, tt.arg)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTrimFuncsBadPattern(t *testing.T) {
	tests := []struct {
		name string
		fn   substituteFunc
	}{
		{name: "shortest prefix", fn: trimShortestPrefix},
		{name: "longest prefix", fn: trimLongestPrefix},
		{name: "shortest suffix", fn: trimShortestSuffix},
		{name: "longest suffix", fn: trimLongestSuffix},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.fn("foobar", "[")
			assert.ErrorIs(t, err, path.ErrBadPattern)
			assert.Empty(t, got)
		})
	}
}

func TestFuncsMissingArgument(t *testing.T) {
	tests := []struct {
		name string
		fn   substituteFunc
		args []string
	}{
		{name: "replace all", fn: replaceAll},
		{name: "replace first", fn: replaceFirst},
		{name: "replace prefix", fn: replacePrefix},
		{name: "replace prefix with only a pattern", fn: replacePrefix, args: []string{"foo"}},
		{name: "replace suffix", fn: replaceSuffix},
		{name: "replace suffix with only a pattern", fn: replaceSuffix, args: []string{"foo"}},
		{name: "shortest prefix", fn: trimShortestPrefix},
		{name: "longest prefix", fn: trimLongestPrefix},
		{name: "shortest suffix", fn: trimShortestSuffix},
		{name: "longest suffix", fn: trimLongestSuffix},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.fn("foobar", tt.args...)
			assert.ErrorIs(t, err, ErrMissingArgument)
			assert.Empty(t, got)
		})
	}
}
