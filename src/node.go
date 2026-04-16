package src

type NodeType int

const (
	DocumentNode NodeType = iota
	DoctypeNode
	ElementNode
	TextNode
	CommentNode
)

type Attribute struct {
	Name  string
	Value string
}

type Node struct {
	Type     NodeType
	Tag      string
	Data     string
	Attrs    []Attribute
	Parent   *Node
	Children []*Node
}

func (n *Node) AppendChild(child *Node) {
	if n == nil || child == nil {
		return
	}
	child.Parent = n
	n.Children = append(n.Children, child)
}