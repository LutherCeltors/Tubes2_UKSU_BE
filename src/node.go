package src

import "strings"

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
	ID          int
	Type        NodeType
	Tag         string
	Data        string
	Attrs       []Attribute
	Parent      *Node
	Children    []*Node
	NextSibling *Node
	PrevSibling *Node
}

func (n *Node) AppendChild(child *Node) {
	if n == nil || child == nil {
		return
	}
	child.Parent = n

	if len(n.Children) > 0 {
		lastChild := n.Children[len(n.Children)-1]
		lastChild.NextSibling = child
		child.PrevSibling = lastChild
	}

	n.Children = append(n.Children, child)
}

func (n *Node) GetAttribute(name string) (string, bool) {
	if n == nil || n.Type != ElementNode {
		return "", false
	}

	target := strings.ToLower(strings.TrimSpace(name))
	for _, attr := range n.Attrs {
		if strings.ToLower(attr.Name) == target {
			return attr.Value, true
		}
	}
	return "", false
}