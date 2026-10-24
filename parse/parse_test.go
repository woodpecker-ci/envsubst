package parse

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var tests = []struct {
	Text string
	Node Node
}{
	//
	// text only
	//
	{
		Text: "text",
		Node: &TextNode{Value: "text"},
	},
	{
		Text: "}text",
		Node: &TextNode{Value: "}text"},
	},
	{
		Text: "http://github.com",
		Node: &TextNode{Value: "http://github.com"}, // should not escape double slash
	},
	{
		Text: "$${string}",
		Node: &TextNode{Value: "${string}"}, // should not escape double dollar
	},
	{
		Text: "$$string",
		Node: &TextNode{Value: "$string"}, // should not escape double dollar
	},
	{
		Text: `\\.\pipe\pipename`,
		Node: &TextNode{Value: `\\.\pipe\pipename`},
	},

	//
	// variable only
	//
	{
		Text: "${string}",
		Node: &FuncNode{Param: "string"},
	},

	//
	// text transform functions
	//
	{
		Text: "${string,}",
		Node: &FuncNode{
			Param: "string",
			Name:  ",",
			Args:  nil,
		},
	},
	{
		Text: "${string,,}",
		Node: &FuncNode{
			Param: "string",
			Name:  ",,",
			Args:  nil,
		},
	},
	{
		Text: "${string^}",
		Node: &FuncNode{
			Param: "string",
			Name:  "^",
			Args:  nil,
		},
	},
	{
		Text: "${string^^}",
		Node: &FuncNode{
			Param: "string",
			Name:  "^^",
			Args:  nil,
		},
	},

	//
	// substring functions
	//
	{
		Text: "${string:position}",
		Node: &FuncNode{
			Param: "string",
			Name:  ":",
			Args: []Node{
				&TextNode{Value: "position"},
			},
		},
	},
	{
		Text: "${string:position:length}",
		Node: &FuncNode{
			Param: "string",
			Name:  ":",
			Args: []Node{
				&TextNode{Value: "position"},
				&TextNode{Value: "length"},
			},
		},
	},

	//
	// string removal functions
	//
	{
		Text: "${string#substring}",
		Node: &FuncNode{
			Param: "string",
			Name:  "#",
			Args: []Node{
				&TextNode{Value: "substring"},
			},
		},
	},
	{
		Text: "${string##substring}",
		Node: &FuncNode{
			Param: "string",
			Name:  "##",
			Args: []Node{
				&TextNode{Value: "substring"},
			},
		},
	},
	{
		Text: "${string%substring}",
		Node: &FuncNode{
			Param: "string",
			Name:  "%",
			Args: []Node{
				&TextNode{Value: "substring"},
			},
		},
	},
	{
		Text: "${string%%substring}",
		Node: &FuncNode{
			Param: "string",
			Name:  "%%",
			Args: []Node{
				&TextNode{Value: "substring"},
			},
		},
	},

	//
	// string replace functions
	//
	{
		Text: "${string/substring/replacement}",
		Node: &FuncNode{
			Param: "string",
			Name:  "/",
			Args: []Node{
				&TextNode{Value: "substring"},
				&TextNode{Value: "replacement"},
			},
		},
	},
	{
		Text: "${string//substring/replacement}",
		Node: &FuncNode{
			Param: "string",
			Name:  "//",
			Args: []Node{
				&TextNode{Value: "substring"},
				&TextNode{Value: "replacement"},
			},
		},
	},
	{
		Text: "${string/#substring/replacement}",
		Node: &FuncNode{
			Param: "string",
			Name:  "/#",
			Args: []Node{
				&TextNode{Value: "substring"},
				&TextNode{Value: "replacement"},
			},
		},
	},
	{
		Text: "${string/%substring/replacement}",
		Node: &FuncNode{
			Param: "string",
			Name:  "/%",
			Args: []Node{
				&TextNode{Value: "substring"},
				&TextNode{Value: "replacement"},
			},
		},
	},

	//
	// default value functions
	//
	{
		Text: "${string=default}",
		Node: &FuncNode{
			Param: "string",
			Name:  "=",
			Args: []Node{
				&TextNode{Value: "default"},
			},
		},
	},
	{
		Text: "${string:=default}",
		Node: &FuncNode{
			Param: "string",
			Name:  ":=",
			Args: []Node{
				&TextNode{Value: "default"},
			},
		},
	},
	{
		Text: "${string:-default}",
		Node: &FuncNode{
			Param: "string",
			Name:  ":-",
			Args: []Node{
				&TextNode{Value: "default"},
			},
		},
	},
	{
		Text: "${string:?default}",
		Node: &FuncNode{
			Param: "string",
			Name:  ":?",
			Args: []Node{
				&TextNode{Value: "default"},
			},
		},
	},
	{
		Text: "${string:+default}",
		Node: &FuncNode{
			Param: "string",
			Name:  ":+",
			Args: []Node{
				&TextNode{Value: "default"},
			},
		},
	},

	//
	// length function
	//
	{
		Text: "${#string}",
		Node: &FuncNode{
			Param: "string",
			Name:  "#",
		},
	},

	//
	// special characters in argument
	//
	{
		Text: "${string#$%:*{}",
		Node: &FuncNode{
			Param: "string",
			Name:  "#",
			Args: []Node{
				&TextNode{Value: "$%:*{"},
			},
		},
	},

	// text before and after function
	{
		Text: "hello ${#string} world",
		Node: &ListNode{
			Nodes: []Node{
				&TextNode{
					Value: "hello ",
				},
				&ListNode{
					Nodes: []Node{
						&FuncNode{
							Param: "string",
							Name:  "#",
						},
						&TextNode{
							Value: " world",
						},
					},
				},
			},
		},
	},
	// text before and after function with \\ outside of function
	{
		Text: `\\ hello ${#string} world \\`,
		Node: &ListNode{
			Nodes: []Node{
				&TextNode{
					Value: `\\ hello `,
				},
				&ListNode{
					Nodes: []Node{
						&FuncNode{
							Param: "string",
							Name:  "#",
						},
						&TextNode{
							Value: ` world \\`,
						},
					},
				},
			},
		},
	},

	// escaped function arguments
	{
		Text: `${string/\/position/length}`,
		Node: &FuncNode{
			Param: "string",
			Name:  "/",
			Args: []Node{
				&TextNode{
					Value: "/position",
				},
				&TextNode{
					Value: "length",
				},
			},
		},
	},
	{
		Text: `${string/\/position\\/length}`,
		Node: &FuncNode{
			Param: "string",
			Name:  "/",
			Args: []Node{
				&TextNode{
					Value: "/position\\",
				},
				&TextNode{
					Value: "length",
				},
			},
		},
	},
	{
		Text: `${string/position/\/length}`,
		Node: &FuncNode{
			Param: "string",
			Name:  "/",
			Args: []Node{
				&TextNode{
					Value: "position",
				},
				&TextNode{
					Value: "/length",
				},
			},
		},
	},
	{
		Text: `${string/position/\/length\\}`,
		Node: &FuncNode{
			Param: "string",
			Name:  "/",
			Args: []Node{
				&TextNode{
					Value: "position",
				},
				&TextNode{
					Value: "/length\\",
				},
			},
		},
	},
	{
		Text: `${string/position/\/leng\\th}`,
		Node: &FuncNode{
			Param: "string",
			Name:  "/",
			Args: []Node{
				&TextNode{
					Value: "position",
				},
				&TextNode{
					Value: "/leng\\th",
				},
			},
		},
	},

	// functions in functions
	{
		Text: "${string:${position}}",
		Node: &FuncNode{
			Param: "string",
			Name:  ":",
			Args: []Node{
				&FuncNode{
					Param: "position",
				},
			},
		},
	},
	{
		Text: "${string:${stringy:position:length}:${stringz,,}}",
		Node: &FuncNode{
			Param: "string",
			Name:  ":",
			Args: []Node{
				&FuncNode{
					Param: "stringy",
					Name:  ":",
					Args: []Node{
						&TextNode{Value: "position"},
						&TextNode{Value: "length"},
					},
				},
				&FuncNode{
					Param: "stringz",
					Name:  ",,",
				},
			},
		},
	},
	{
		Text: "${string#${stringz}}",
		Node: &FuncNode{
			Param: "string",
			Name:  "#",
			Args: []Node{
				&FuncNode{Param: "stringz"},
			},
		},
	},
	{
		Text: "${string=${stringz}}",
		Node: &FuncNode{
			Param: "string",
			Name:  "=",
			Args: []Node{
				&FuncNode{Param: "stringz"},
			},
		},
	},
	{
		Text: "${string=prefix-${var}}",
		Node: &FuncNode{
			Param: "string",
			Name:  "=",
			Args: []Node{
				&TextNode{Value: "prefix-"},
				&FuncNode{Param: "var"},
			},
		},
	},
	{
		Text: "${string=${var}-suffix}",
		Node: &FuncNode{
			Param: "string",
			Name:  "=",
			Args: []Node{
				&FuncNode{Param: "var"},
				&TextNode{Value: "-suffix"},
			},
		},
	},
	{
		Text: "${string=prefix-${var}-suffix}",
		Node: &FuncNode{
			Param: "string",
			Name:  "=",
			Args: []Node{
				&TextNode{Value: "prefix-"},
				&FuncNode{Param: "var"},
				&TextNode{Value: "-suffix"},
			},
		},
	},
	{
		Text: "${string=prefix${var} suffix}",
		Node: &FuncNode{
			Param: "string",
			Name:  "=",
			Args: []Node{
				&TextNode{Value: "prefix"},
				&FuncNode{Param: "var"},
				&TextNode{Value: " suffix"},
			},
		},
	},
	{
		Text: "${string//${stringy}/${stringz}}",
		Node: &FuncNode{
			Param: "string",
			Name:  "//",
			Args: []Node{
				&FuncNode{Param: "stringy"},
				&FuncNode{Param: "stringz"},
			},
		},
	},
}

func TestParse(t *testing.T) {
	for _, test := range tests {
		t.Run(test.Text, func(t *testing.T) {
			got, err := Parse(test.Text)
			require.NoError(t, err)
			assert.Equal(t, test.Node, got.Root)
		})
	}
}

func TestParseError(t *testing.T) {
	tests := []struct {
		text string
		want error
	}{
		{text: "${", want: ErrParseVariableName},
		{text: "${}", want: ErrParseVariableName},
		{text: "${!a}", want: ErrParseVariableName},
		{text: "${a", want: ErrMissingClosingBrace},
		{text: "${a:-${b", want: ErrMissingClosingBrace},
		{text: "${a:", want: ErrBadSubstitution},
		{text: "${a:1", want: ErrBadSubstitution},
		{text: "${a:1:2", want: ErrBadSubstitution},
		{text: "${a/x", want: ErrBadSubstitution},
		{text: "${a/x/y", want: ErrBadSubstitution},
		{text: "${a//x/y", want: ErrBadSubstitution},
		{text: "${a/#x/y", want: ErrBadSubstitution},
		{text: "${a/%x/y", want: ErrBadSubstitution},
		{text: "${a^", want: ErrBadSubstitution},
		{text: "${a^^", want: ErrBadSubstitution},
		{text: "${a,", want: ErrBadSubstitution},
		{text: "${a,,", want: ErrBadSubstitution},
		{text: "${#", want: ErrBadSubstitution},
		{text: "${#a", want: ErrBadSubstitution},
		{text: "${a:1:", want: ErrParseFuncSubstitution},
		{text: "${a:-", want: ErrParseFuncSubstitution},
		{text: "${a:-x", want: ErrParseFuncSubstitution},
		{text: "${a:-${b}", want: ErrParseFuncSubstitution},
		{text: "${a/", want: ErrParseFuncSubstitution},
		{text: "${a/x/", want: ErrParseFuncSubstitution},
		{text: "${a#", want: ErrParseFuncSubstitution},
		{text: "${a##", want: ErrParseFuncSubstitution},
		{text: "${a%", want: ErrParseFuncSubstitution},
		{text: "${a%%", want: ErrParseFuncSubstitution},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			_, err := Parse(tt.text)
			assert.ErrorIs(t, err, tt.want)
		})
	}
}
