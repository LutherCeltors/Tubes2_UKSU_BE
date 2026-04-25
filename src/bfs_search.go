package src

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func BFSSearch(root *Node, query string, topN int) ([]*Node, []LogEntry, int, error) {
	if root == nil {
		return nil, nil, 0, fmt.Errorf("root node is nil")
	}
	selector, err := ParseSelector(query)
	if err != nil {
		return nil, nil, 0, err
	}

	type levelResult struct {
		matched  *Node
		logEntry *LogEntry
		children []*Node
	}

	var (
		results      []*Node
		logs         []LogEntry
		nodesVisited int32
	)

	currentLevel := []*Node{root}
	levelIndex := 0

	for len(currentLevel) > 0 {
		levelResults := make([]levelResult, len(currentLevel))
		var wg sync.WaitGroup

		for i, node := range currentLevel {
			wg.Add(1)
			go func(idx int, n *Node, batch int) {
				defer wg.Done()
				r := levelResult{children: n.Children}
				switch n.Type {
				case ElementNode:
					atomic.AddInt32(&nodesVisited, 1)
					match := selector.Match(n)
					if match {
						r.matched = n
					}
					status := "visited"
					if match {
						status = "matched"
					}
					entry := LogEntry{NodeID: n.ID, Tag: n.Tag, Status: status, Batch: batch}
					r.logEntry = &entry
				case DocumentNode:
					atomic.AddInt32(&nodesVisited, 1)
				case TextNode:
					atomic.AddInt32(&nodesVisited, 1)
					entry := LogEntry{NodeID: n.ID, Tag: "#text", Status: "visited", Batch: batch}
					r.logEntry = &entry
				}
				levelResults[idx] = r
			}(i, node, levelIndex)
		}

		wg.Wait()

		var nextLevel []*Node
		stop := false
		for _, r := range levelResults {
			if r.logEntry != nil {
				logs = append(logs, *r.logEntry)
			}
			if r.matched != nil {
				results = append(results, r.matched)
			}
			nextLevel = append(nextLevel, r.children...)
			if topN > 0 && len(results) >= topN {
				stop = true
				break
			}
		}

		if stop {
			break
		}
		currentLevel = nextLevel
		levelIndex++
	}

	return results, logs, int(atomic.LoadInt32(&nodesVisited)), nil
}

func BFSSearchSingle(root *Node, query string, topN int) ([]*Node, []LogEntry, int, error) {
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
	batchIndex := 0

	queue := []*Node{root}

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		switch node.Type {
		case ElementNode:
			nodesVisited++
			match := selector.Match(node)
			status := "visited"
			if match {
				status = "matched"
				results = append(results, node)
			}
			logs = append(logs, LogEntry{NodeID: node.ID, Tag: node.Tag, Status: status, Batch: batchIndex})
			batchIndex++
		case DocumentNode:
			nodesVisited++
		case TextNode:
			nodesVisited++
			logs = append(logs, LogEntry{NodeID: node.ID, Tag: "#text", Status: "visited", Batch: batchIndex})
			batchIndex++
		}

		if topN > 0 && len(results) >= topN {
			break
		}

		queue = append(queue, node.Children...)
	}

	return results, logs, nodesVisited, nil
}
