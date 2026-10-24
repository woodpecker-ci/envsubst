package envsubst

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
