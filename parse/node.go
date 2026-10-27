package parse

// Node is an element in the parse tree.
type Node interface {
	node()
}

// Empty string node.
var empty = new(TextNode)

// a template is represented by a tree consisting of one
// or more of the following nodes.
type (
	// TextNode represents a string of text.
	TextNode struct {
		Value string
	}

	// FuncNode represents a string function.
	FuncNode struct {
		Param string
		Name  string
		Args  []Node
	}

	// ListNode represents a list of nodes.
	ListNode struct {
		Nodes []Node
	}
)

// newTextNode returns a new TextNode.
func newTextNode(text string) *TextNode {
	return &TextNode{Value: text}
}

// newListNode returns a new ListNode.
func newListNode(nodes ...Node) *ListNode {
	return &ListNode{Nodes: nodes}
}

// newFuncNode returns a new FuncNode.
func newFuncNode(name string) *FuncNode {
	return &FuncNode{Param: name}
}

// The node method defines the node in a parse tree.

func (*TextNode) node() {}
func (*ListNode) node() {}
func (*FuncNode) node() {}
