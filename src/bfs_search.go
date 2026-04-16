package src

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type combinatorType int

const (
	combinatorNone combinatorType=iota
	combinatorDescendant
	combinatorChild
	combinatorAdjacentSibling
	combinatorGeneralSibling
)

type selectorStep struct {
	raw        string
	combinator combinatorType
}

func bfsSearchAlgorithm(selector string, domTree *Node) *Node {
	result, err:=BFSSearch(selector, domTree, "top n kemunculan", 1)
	if err!=nil || result==nil || len(result.Matches)==0 {
		return nil
	}
	return result.Matches[0]
}

func BFSSearch(selector string, domTree *Node, resultMode string, topN int) (*SearchOutput, error) {
	start:=time.Now()

	if domTree==nil {
		return &SearchOutput{
			Matches:      nil,
			TraversalLog: nil,
			Count: 0,
			TraversalTime: time.Since(start).Milliseconds(),
		}, nil
	}

	steps, err:=parseSelectorSteps(selector)
	if err!=nil {
		return nil, err
	}
	if len(steps)==0 {
		return &SearchOutput{
			Matches:      nil,
			TraversalLog: nil,
			Count: 0,
			TraversalTime: time.Since(start).Milliseconds(),
		}, nil
	}

	allElements:=collectElementsBFS(domTree)
	if len(allElements)==0 {
		return &SearchOutput{
			Matches:      nil,
			TraversalLog: nil,
			Count: 0,
			TraversalTime: time.Since(start).Milliseconds(),
		}, nil
	}

	bfsIndex:=make(map[*Node]int, len(allElements))
	for i, node:=range allElements {
		bfsIndex[node]=i
	}

	current:=make([]*Node, 0)
	traversalLog:=make([]*Node, 0)
	checkedCount:=0
	for _, node:=range allElements {
		traversalLog=append(traversalLog, node)
		checkedCount++
		if matchesSimpleSelector(node, steps[0].raw) {
			current=append(current, node)
		}
	}

	for i:=1; i < len(steps); i++ {
		candidates:=relatedCandidates(current, steps[i].combinator)

		seen:=make(map[*Node]struct{}, len(candidates))
		next:=make([]*Node, 0, len(candidates))
		for _, candidate:=range candidates {
			traversalLog=append(traversalLog, candidate)
			checkedCount++
			if !matchesSimpleSelector(candidate, steps[i].raw) {
				continue
			}
			if _, ok:=seen[candidate]; ok {
				continue
			}
			seen[candidate]=struct{}{}
			next=append(next, candidate)
		}

		sort.Slice(next, func(a, b int) bool {
			return bfsIndex[next[a]] < bfsIndex[next[b]]
		})

		current=next
		if len(current)==0 {
			break
		}
	}

	matches, err:=applyResultMode(current, resultMode, topN)
	if err!=nil {
		return nil, err
	}

	return &SearchOutput{
		Matches:      matches,
		TraversalLog: traversalLog,
		Count: checkedCount,
		TraversalTime: time.Since(start).Milliseconds(),
	}, nil
}

func parseSelectorSteps(selector string) ([]selectorStep, error) {
	s:=strings.TrimSpace(selector)
	if s=="" {
		return nil, fmt.Errorf("Selector kosong")
	}

	steps:=make([]selectorStep, 0)
	pendingCombinator:=combinatorNone
	expectSelector:=true

	for i:=0; i < len(s); {
		spaceStart:=i
		for i < len(s) && isSpace(s[i]) {
			i++
		}
		hadSpace:=i > spaceStart

		if i>=len(s) {
			if expectSelector {
				return nil, fmt.Errorf("Selector tidak valid")
			}
			break
		}

		if isCombinatorChar(s[i]) {
			if expectSelector {
				return nil, fmt.Errorf("Selector tidak valid")
			}
			pendingCombinator=combinatorFromChar(s[i])
			expectSelector=true
			i++
			continue
		}

		if hadSpace && !expectSelector {
			pendingCombinator=combinatorDescendant
			expectSelector=true
		}

		if !expectSelector {
			return nil, fmt.Errorf("Selector tidak valid")
		}

		start:=i
		for i < len(s) && !isSpace(s[i]) && !isCombinatorChar(s[i]) {
			i++
		}
		raw:=s[start:i]
		if err:=validateSimpleSelector(raw); err!=nil {
			return nil, err
		}

		combinator:=pendingCombinator
		if len(steps)==0 {
			combinator=combinatorNone
		} else if combinator==combinatorNone {
			return nil, fmt.Errorf("Selector tidak valid")
		}

		steps=append(steps, selectorStep{
			raw:        raw,
			combinator: combinator,
		})
		pendingCombinator=combinatorNone
		expectSelector=false
	}

	return steps, nil
}

func validateSimpleSelector(raw string) error {
	if raw=="" {
		return fmt.Errorf("Selector kosong")
	}
	if raw=="*" {
		return nil
	}
	if strings.HasPrefix(raw, ".") || strings.HasPrefix(raw, "#") {
		if len(raw)==1 {
			return fmt.Errorf("Selector tidak valid")
		}
		return nil
	}
	return nil
}

func isSpace(ch byte) bool {
	return ch==' ' || ch=='\n' || ch=='\r' || ch=='\t' || ch=='\f'
}

func isCombinatorChar(ch byte) bool {
	return ch=='>' || ch=='+' || ch=='~'
}

func combinatorFromChar(ch byte) combinatorType {
	switch ch {
	case '>':
		return combinatorChild
	case '+':
		return combinatorAdjacentSibling
	case '~':
		return combinatorGeneralSibling
	default:
		return combinatorNone
	}
}

func collectElementsBFS(root *Node) []*Node {
	if root==nil {
		return nil
	}

	queue:=[]*Node{root}
	elements:=make([]*Node, 0)
	for len(queue) > 0 {
		head:=queue[0]
		queue=queue[1:]

		if head.Type==ElementNode {
			elements=append(elements, head)
		}
		queue=append(queue, head.Children...)
	}
	return elements
}

func relatedCandidates(nodes []*Node, relation combinatorType) []*Node {
	candidates:=make([]*Node, 0)

	for _, node:=range nodes {
		switch relation {
		case combinatorDescendant:
			queue:=make([]*Node, 0, len(node.Children))
			queue=append(queue, node.Children...)
			for len(queue) > 0 {
				head:=queue[0]
				queue=queue[1:]
				if head.Type==ElementNode {
					candidates=append(candidates, head)
				}
				queue=append(queue, head.Children...)
			}
		case combinatorChild:
			for _, child:=range node.Children {
				if child.Type==ElementNode {
					candidates=append(candidates, child)
				}
			}
		case combinatorAdjacentSibling:
			if sibling:=nextElementSibling(node); sibling!=nil {
				candidates=append(candidates, sibling)
			}
		case combinatorGeneralSibling:
			candidates=append(candidates, followingElementSiblings(node)...)
		}
	}

	return candidates
}

func matchesSimpleSelector(node *Node, raw string) bool {
	if node==nil || node.Type!=ElementNode {
		return false
	}

	switch {
	case raw=="*":
		return true
	case strings.HasPrefix(raw, "."):
		className:=raw[1:]
		return hasClass(node, className)
	case strings.HasPrefix(raw, "#"):
		idName:=raw[1:]
		return getAttrValue(node, "id")==idName
	default:
		return strings.EqualFold(node.Tag, raw)
	}
}

func hasClass(node *Node, className string) bool {
	if className=="" {
		return false
	}
	classes:=strings.Fields(getAttrValue(node, "class"))
	for _, cls:=range classes {
		if cls==className {
			return true
		}
	}
	return false
}

func getAttrValue(node *Node, name string) string {
	if node==nil {
		return ""
	}
	for _, attr:=range node.Attrs {
		if strings.EqualFold(attr.Name, name) {
			return attr.Value
		}
	}
	return ""
}

func nextElementSibling(node *Node) *Node {
	if node==nil || node.Parent==nil {
		return nil
	}

	children:=node.Parent.Children
	found:=false
	for _, child:=range children {
		if !found {
			if child==node {
				found=true
			}
			continue
		}
		if child.Type==ElementNode {
			return child
		}
	}
	return nil
}

func followingElementSiblings(node *Node) []*Node {
	if node==nil || node.Parent==nil {
		return nil
	}

	children:=node.Parent.Children
	found:=false
	siblings:=make([]*Node, 0)
	for _, child:=range children {
		if !found {
			if child==node {
				found=true
			}
			continue
		}
		if child.Type==ElementNode {
			siblings=append(siblings, child)
		}
	}
	return siblings
}

func applyResultMode(nodes []*Node, resultMode string, topN int) ([]*Node, error) {
	mode:=strings.Join(strings.Fields(strings.ToLower(resultMode)), " ")

	switch mode {
	case "", "all", "semua", "semua kemunculan":
		return nodes, nil
	case "top", "top n", "topn", "top n kemunculan":
		if topN<=0 {
			return nil, fmt.Errorf("TopN tidak valid")
		}
		if topN>=len(nodes) {
			return nodes, nil
		}
		return nodes[:topN], nil
	default:
		return nil, fmt.Errorf("ResultMode tidak valid")
	}
}
