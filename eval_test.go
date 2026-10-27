package envsubst

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/woodpecker-ci/envsubst/parse"
	"github.com/woodpecker-ci/envsubst/path"
)

// Test cases sourced from tldp.org.
// http://www.tldp.org/LDP/abs/html/parameter-substitution.html

func TestExpand(t *testing.T) {
	expressions := []struct {
		params map[string]string
		input  string
		output string
	}{
		// text-only
		{
			params: map[string]string{},
			input:  "abcdEFGH28ij",
			output: "abcdEFGH28ij",
		},
		// length
		{
			params: map[string]string{"var01": "abcdEFGH28ij"},
			input:  "${#var01}",
			output: "12",
		},
		// uppercase first
		{
			params: map[string]string{"var01": "abcdEFGH28ij"},
			input:  "${var01^}",
			output: "AbcdEFGH28ij",
		},
		// uppercase
		{
			params: map[string]string{"var01": "abcdEFGH28ij"},
			input:  "${var01^^}",
			output: "ABCDEFGH28IJ",
		},
		// lowercase first
		{
			params: map[string]string{"var01": "ABCDEFGH28IJ"},
			input:  "${var01,}",
			output: "aBCDEFGH28IJ",
		},
		// lowercase
		{
			params: map[string]string{"var01": "ABCDEFGH28IJ"},
			input:  "${var01,,}",
			output: "abcdefgh28ij",
		},
		// substring with position
		{
			params: map[string]string{"path_name": "/home/bozo/ideas/thoughts.for.today"},
			input:  "${path_name:11}",
			output: "ideas/thoughts.for.today",
		},
		// substring with position and length
		{
			params: map[string]string{"path_name": "/home/bozo/ideas/thoughts.for.today"},
			input:  "${path_name:11:5}",
			output: "ideas",
		},
		// substring with position and negative length
		{
			params: map[string]string{"path_name": "/home/bozo/ideas/thoughts.for.today"},
			input:  "${path_name:11:-6}",
			output: "ideas/thoughts.for",
		},
		// substring with negative length cutting everything
		{
			params: map[string]string{"var": "abc"},
			input:  "[${var:0:-3}]",
			output: "[]",
		},
		// substring with negative length from a variable
		{
			params: map[string]string{"var": "abcdefgh", "pos": "1", "len": "-2"},
			input:  "${var:${pos}:${len}}",
			output: "bcdef",
		},
		// default not used
		{
			params: map[string]string{"var": "abc"},
			input:  "${var=xyz}",
			output: "abc",
		},
		// default used
		{
			params: map[string]string{},
			input:  "${var=xyz}",
			output: "xyz",
		},
		{
			params: map[string]string{"default_var": "foo"},
			input:  "something ${var=${default_var}}",
			output: "something foo",
		},
		{
			params: map[string]string{"default_var": "foo1"},
			input:  `foo: ${var=${default_var}-suffix}`,
			output: "foo: foo1-suffix",
		},
		{
			params: map[string]string{"default_var": "foo1"},
			input:  `foo: ${var=prefix${default_var}-suffix}`,
			output: "foo: prefixfoo1-suffix",
		},
		{
			params: map[string]string{},
			input:  "${var:=xyz}",
			output: "xyz",
		},
		// replace suffix
		{
			params: map[string]string{"stringZ": "abcABC123ABCabc"},
			input:  "${stringZ/%abc/XYZ}",
			output: "abcABC123ABCXYZ",
		},
		// replace prefix
		{
			params: map[string]string{"stringZ": "abcABC123ABCabc"},
			input:  "${stringZ/#abc/XYZ}",
			output: "XYZABC123ABCabc",
		},
		// replace all
		{
			params: map[string]string{"stringZ": "abcABC123ABCabc"},
			input:  "${stringZ//abc/xyz}",
			output: "xyzABC123ABCxyz",
		},
		// replace first
		{
			params: map[string]string{"stringZ": "abcABC123ABCabc"},
			input:  "${stringZ/abc/xyz}",
			output: "xyzABC123ABCabc",
		},
		// delete shortest match prefix
		{
			params: map[string]string{"filename": "bash.string.txt"},
			input:  "${filename#*.}",
			output: "string.txt",
		},
		{
			params: map[string]string{"filename": "path/to/file"},
			input:  "${filename#*/}",
			output: "to/file",
		},
		{
			params: map[string]string{"filename": "/path/to/file"},
			input:  "${filename#*/}",
			output: "path/to/file",
		},
		// delete longest match prefix
		{
			params: map[string]string{"filename": "bash.string.txt"},
			input:  "${filename##*.}",
			output: "txt",
		},
		{
			params: map[string]string{"filename": "path/to/file"},
			input:  "${filename##*/}",
			output: "file",
		},
		{
			params: map[string]string{"filename": "/path/to/file"},
			input:  "${filename##*/}",
			output: "file",
		},
		// delete shortest match suffix
		{
			params: map[string]string{"filename": "bash.string.txt"},
			input:  "${filename%.*}",
			output: "bash.string",
		},
		// delete longest match suffix
		{
			params: map[string]string{"filename": "bash.string.txt"},
			input:  "${filename%%.*}",
			output: "bash",
		},

		// nested parameters
		{
			params: map[string]string{"var01": "abcdEFGH28ij"},
			input:  "${var=${var01^^}}",
			output: "ABCDEFGH28IJ",
		},
		// escaped
		{
			params: map[string]string{"var01": "abcdEFGH28ij"},
			input:  "$${var01}",
			output: "${var01}",
		},
		{
			params: map[string]string{"var01": "abcdEFGH28ij"},
			input:  "some text ${var01}$${var$${var01}$var01${var01}",
			output: "some text abcdEFGH28ij${var${var01}$var01abcdEFGH28ij",
		},
		{
			params: map[string]string{"default_var": "foo"},
			input:  "something $${var=${default_var}}",
			output: "something ${var=foo}",
		},
		// some common escaping use cases
		{
			params: map[string]string{"stringZ": "foo/bar"},
			input:  `${stringZ/\//-}`,
			output: "foo-bar",
		},
		{
			params: map[string]string{"stringZ": "foo/bar/baz"},
			input:  `${stringZ//\//-}`,
			output: "foo-bar-baz",
		},
		// escape outside of expansion shouldn't be processed
		{
			params: map[string]string{"default_var": "foo"},
			input:  "\\\\something ${var=${default_var}}",
			output: "\\\\something foo",
		},
		// substitute with a blank string
		{
			params: map[string]string{"stringZ": "foo.bar"},
			input:  `${stringZ/./}`,
			output: "foobar",
		},
	}

	for _, expr := range expressions {
		t.Run(expr.input, func(t *testing.T) {
			output, err := Eval(expr.input, func(s string) string {
				return expr.params[s]
			})
			require.NoError(t, err)
			assert.Equal(t, expr.output, output)
		})
	}
}

func TestExpandError(t *testing.T) {
	for input, want := range map[string]error{
		"${var:abc}":  ErrBadSubstringPosition,
		"${var:1:x}":  ErrBadSubstringLength,
		"${var:3:-9}": ErrBadSubstringLength,
		"${var:8:-1}": ErrBadSubstringLength,
		"${var##[}":   path.ErrBadPattern,
	} {
		t.Run(input, func(t *testing.T) {
			_, err := Eval(input, func(string) string { return "abcdefgh" })
			assert.ErrorIs(t, err, want)
		})
	}
}

func TestExpandErrorUnsetVariable(t *testing.T) {
	// a negative length on an unset variable used to panic
	_, err := Eval("${unset:0:-1}", func(string) string { return "" })
	assert.ErrorIs(t, err, ErrBadSubstringLength)
}

func TestExpandErrorInNestedArgument(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "in the argument of a function", input: "${a:-${b:1:x}}"},
		{name: "after plain text", input: "text ${a:1:x}"},
		{name: "before another substitution", input: "${a:1:x}${b}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Eval(tt.input, func(string) string { return "abcdefgh" })
			assert.ErrorIs(t, err, ErrBadSubstringLength)
		})
	}
}

func TestEvalParseError(t *testing.T) {
	// the input is returned unchanged next to the error
	input := "before ${unclosed"
	output, err := Eval(input, func(string) string { return "value" })
	assert.ErrorIs(t, err, parse.ErrMissingClosingBrace)
	assert.Equal(t, input, output)
}

func TestEvalEnv(t *testing.T) {
	t.Setenv("ENVSUBST_TEST_SET", "value")
	t.Setenv("ENVSUBST_TEST_EMPTY", "")

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "set variable", input: "${ENVSUBST_TEST_SET}", want: "value"},
		{name: "undefined variable is empty", input: "[${ENVSUBST_TEST_UNDEFINED}]", want: "[]"},
		{name: "default for empty variable", input: "${ENVSUBST_TEST_EMPTY:-fallback}", want: "fallback"},
		{name: "function on set variable", input: "${ENVSUBST_TEST_SET^^}", want: "VALUE"},
		{name: "no substitution", input: "plain text", want: "plain text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EvalEnv(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

	t.Run("parse error", func(t *testing.T) {
		_, err := EvalEnv("${ENVSUBST_TEST_SET")
		assert.ErrorIs(t, err, parse.ErrMissingClosingBrace)
	})
}
