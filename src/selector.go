package src

import (
	"fmt"
	"regexp"
	"strings"
)

type Combinator string

const (
	CombNone       Combinator = ""
	CombDescendant Combinator = " "
	CombChild      Combinator = ">"
	CombAdjSibling Combinator = "+"
	CombGenSibling Combinator = "~"
)

type SimpleSelector struct {
	Tag         string
	ID          string
	Classes     []string
	IsUniversal bool
}

func (s *SimpleSelector) Match(n *Node) bool {
	if n == nil || n.Type != ElementNode {
		return false
	}
	if !s.IsUniversal && s.Tag != "" && strings.ToLower(s.Tag) != strings.ToLower(n.Tag) {
		return false
	}
	if s.ID != "" {
		if id, ok := n.GetAttribute("id"); !ok || id != s.ID {
			return false
		}
	}
	if len(s.Classes) > 0 {
		classStr, _ := n.GetAttribute("class")
		nodeClasses := strings.Fields(classStr)
		classMap := make(map[string]bool)
		for _, c := range nodeClasses {
			classMap[c] = true
		}

		for _, c := range s.Classes {
			if !classMap[c] {
				return false
			}
		}
	}
	if !s.IsUniversal && s.Tag == "" && s.ID == "" && len(s.Classes) == 0 {
		return false
	}
	return true
}

type ComplexSelector struct {
	Simple *SimpleSelector
	Comb   Combinator
	Left   *ComplexSelector
}

// Match evaluates right-to-left (bottom-up)
func (cs *ComplexSelector) Match(n *Node) bool {
	if n == nil {
		return false
	}
	if !cs.Simple.Match(n) {
		return false
	}
	if cs.Left == nil {
		return true
	}

	switch cs.Comb {
	case CombChild:
		return cs.Left.Match(n.Parent)
	case CombDescendant:
		curr := n.Parent
		for curr != nil {
			if cs.Left.Match(curr) {
				return true
			}
			curr = curr.Parent
		}
		return false
	case CombAdjSibling:
		curr := n.PrevSibling
		for curr != nil && curr.Type != ElementNode {
			curr = curr.PrevSibling
		}
		return cs.Left.Match(curr)
	case CombGenSibling:
		curr := n.PrevSibling
		for curr != nil {
			if curr.Type == ElementNode && cs.Left.Match(curr) {
				return true
			}
			curr = curr.PrevSibling
		}
		return false
	}
	return false
}

func ParseSelector(query string) (*ComplexSelector, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("empty selector")
	}

	query = regexp.MustCompile(`\s*([>+~])\s*`).ReplaceAllString(query, "$1")
	query = regexp.MustCompile(`\s+`).ReplaceAllString(query, " ")

	var tokens []string
	lexer := regexp.MustCompile(`([>+~\s])|([^>+~\s]+)`)
	matches := lexer.FindAllStringSubmatch(query, -1)

	for _, m := range matches {
		if m[1] != "" { // Combinator
			tokens = append(tokens, m[1])
		} else if m[2] != "" {
			tokens = append(tokens, m[2])
		}
	}

	if len(tokens) == 0 {
		return nil, fmt.Errorf("invalid selector")
	}

	var root *ComplexSelector
	var currentComb Combinator = CombNone

	for i := 0; i < len(tokens); i++ {
		t := tokens[i]
		if t == ">" || t == "+" || t == "~" || t == " " {
			if currentComb != CombNone {
				return nil, fmt.Errorf("unexpected combinator: %v", t)
			}
			currentComb = Combinator(t)
			continue
		}

		simple, err := parseSimpleSelector(t)
		if err != nil {
			return nil, err
		}

		if root == nil {
			root = &ComplexSelector{Simple: simple}
		} else {
			root = &ComplexSelector{
				Simple: simple,
				Comb:   currentComb,
				Left:   root,
			}
		}
		currentComb = CombNone
	}

	if currentComb != CombNone {
		return nil, fmt.Errorf("dangling combinator")
	}

	return root, nil
}

func parseSimpleSelector(s string) (*SimpleSelector, error) {
	simple := &SimpleSelector{}
	re := regexp.MustCompile(`(\*)|(?:^([a-zA-Z0-9_-]+))|(#([a-zA-Z0-9_-]+))|(\.([a-zA-Z0-9_-]+))`)
	matches := re.FindAllStringSubmatch(s, -1)

	if len(matches) == 0 {
		return nil, fmt.Errorf("invalid simple selector: %s", s)
	}

	for _, m := range matches {
		if m[1] != "" {
			simple.IsUniversal = true
		} else if m[2] != "" {
			simple.Tag = m[2]
		} else if m[4] != "" {
			simple.ID = m[4]
		} else if m[6] != "" {
			simple.Classes = append(simple.Classes, m[6])
		}
	}
	return simple, nil
}
