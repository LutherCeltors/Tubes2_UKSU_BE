package src

import "fmt"

type LogEntry struct {
	NodeID int    `json:"nodeId"`
	Tag    string `json:"tag"`
	Status string `json:"status"`
}

func SearchDFS(root *Node, query string, topN int) ([]*Node, []LogEntry, int, error) {
	if root == nil {
		return nil, nil, 0, fmt.Errorf("root node is nil")
	}
	selector, err := ParseSelector(query)
	if err != nil {
		return nil, nil, 0, err
	}

	var results []*Node
	var logs []LogEntry
	nodesVisited := 0

	var dfs func(n *Node) bool
	dfs = func(n *Node) bool {
		if n == nil {
			return false
		}

		if n.Type == ElementNode || n.Type == DocumentNode {
			nodesVisited++
			match := false
			if n.Type == ElementNode && selector.Match(n) {
				match = true
				results = append(results, n)
			}

			status := "visited"
			if match {
				status = "matched"
			}

			if n.Type == ElementNode {
				logs = append(logs, LogEntry{
					NodeID: n.ID,
					Tag:    n.Tag,
					Status: status,
				})
			}
			if topN > 0 && len(results) >= topN {
				return true
			}
		}
		for _, child := range n.Children {
			if dfs(child) {
				return true
			}
		}
		return false
	}

	dfs(root)

	return results, logs, nodesVisited, nil
}

type JSONNode struct {
	ID         int               `json:"id"`
	Tag        string            `json:"tag"`
	Attributes map[string]string `json:"attributes,omitempty"`
	Children   []*JSONNode       `json:"children,omitempty"`
}

func ConvertToJSONNode(n *Node) *JSONNode {
	if n == nil {
		return nil
	}
	if n.Type != ElementNode && n.Type != DocumentNode {
		return nil
	}
	attrs := make(map[string]string)
	for _, a := range n.Attrs {
		attrs[a.Name] = a.Value
	}
	tag := n.Tag
	if n.Type == DocumentNode {
		tag = "document"
	}
	res := &JSONNode{
		ID:         n.ID,
		Tag:        tag,
		Attributes: attrs,
	}
	for _, child := range n.Children {
		if cJson := ConvertToJSONNode(child); cJson != nil {
			res.Children = append(res.Children, cJson)
		}
	}
	return res
}
