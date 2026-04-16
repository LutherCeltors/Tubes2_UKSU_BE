package src

import "strings"

type NodeType int

const (
	DocumentNode NodeType=iota
	DoctypeNode
	ElementNode
	TextNode
	CommentNode
)

type Attribute struct {
	Name string
	Value string
}

type Node struct {
	Type NodeType
	Tag string
	Data string
	Attrs []Attribute
	Parent *Node
	Children []*Node
}

func (n *Node) AppendChild(child *Node) {
	if n==nil || child==nil {
		return
	}
	child.Parent=n
	n.Children=append(n.Children, child)
}

func (n *Node) GetAttribute(name string) (string, bool) {
	if n==nil || n.Type!=ElementNode {
		return "", false
	}

	target:=strings.ToLower(strings.TrimSpace(name))
	for _, attr:=range n.Attrs {
		if strings.ToLower(attr.Name)==target {
			return attr.Value, true
		}
	}
	return "", false
}

func FindFirstByTag(root *Node, tag string) *Node {
	if root==nil {
		return nil
	}

	target:=strings.ToLower(strings.TrimSpace(tag))
	if root.Type==ElementNode && root.Tag==target {
		return root
	}

	for _, child:=range root.Children {
		if found:=FindFirstByTag(child, target); found!=nil {
			return found
		}
	}
	return nil
}

func FindAllByTag(root *Node, tag string) []*Node {
	if root==nil {
		return nil
	}

	target:=strings.ToLower(strings.TrimSpace(tag))
	results:=make([]*Node, 0)

	var walk func(*Node)
	walk=func(curr *Node) {
		if curr==nil {
			return
		}
		if curr.Type==ElementNode && curr.Tag==target {
			results=append(results, curr)
		}
		for _, child:=range curr.Children {
			walk(child)
		}
	}

	walk(root)
	return results
}

func RawTextContent(root *Node) string {
	if root==nil {
		return ""
	}

	var builder strings.Builder

	var walk func(*Node)
	walk=func(curr *Node) {
		if curr==nil {
			return
		}
		if curr.Type==TextNode {
			builder.WriteString(curr.Data)
			return
		}
		for _, child:=range curr.Children {
			walk(child)
		}
	}

	walk(root)
	return builder.String()
}

func TextContent(root *Node) string {
	raw:=RawTextContent(root)
	return strings.Join(strings.Fields(raw), " ")
}