package src

import (
	"fmt"
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

type AttrSelector struct {
	Name  string
	Op    string
	Value string
}

func (a *AttrSelector) Match(n *Node) bool {
	val, ok := n.GetAttribute(a.Name)
	if !ok {
		return false
	}
	switch a.Op {
	case "":
		return true
	case "=":
		return val == a.Value
	case "~=":
		if a.Value == "" {
			return false
		}
		for _, w := range strings.Fields(val) {
			if w == a.Value {
				return true
			}
		}
		return false
	case "|=":
		return val == a.Value || strings.HasPrefix(val, a.Value+"-")
	case "^=":
		return a.Value != "" && strings.HasPrefix(val, a.Value)
	case "$=":
		return a.Value != "" && strings.HasSuffix(val, a.Value)
	case "*=":
		return a.Value != "" && strings.Contains(val, a.Value)
	}
	return false
}

type SimpleSelector struct {
	Tag         string
	ID          string
	Classes     []string
	Attrs       []AttrSelector
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
	for i := range s.Attrs {
		if !s.Attrs[i].Match(n) {
			return false
		}
	}
	if !s.IsUniversal && s.Tag == "" && s.ID == "" && len(s.Classes) == 0 && len(s.Attrs) == 0 {
		return false
	}
	return true
}

type ComplexSelector struct {
	Simple *SimpleSelector
	Comb   Combinator
	Left   *ComplexSelector
}

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

	tokens, err := tokenizeSelector(query)
	if err != nil {
		return nil, err
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

func tokenizeSelector(query string) ([]string, error) {
	var tokens []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			tokens = append(tokens, cur.String())
			cur.Reset()
		}
	}
	isCombinator := func(s string) bool {
		return s == ">" || s == "+" || s == "~"
	}

	n := len(query)
	for i := 0; i < n; {
		c := query[i]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			flush()
			j := i
			for j < n && (query[j] == ' ' || query[j] == '\t' || query[j] == '\n' || query[j] == '\r') {
				j++
			}
			emit := len(tokens) > 0 && j < n
			if emit {
				next := query[j]
				if next == '>' || next == '+' || next == '~' {
					emit = false
				}
				if len(tokens) > 0 && isCombinator(tokens[len(tokens)-1]) {
					emit = false
				}
			}
			if emit {
				tokens = append(tokens, " ")
			}
			i = j
		case c == '>' || c == '+' || c == '~':
			flush()
			tokens = append(tokens, string(c))
			i++
		case c == '[':
			cur.WriteByte(c)
			i++
			for i < n && query[i] != ']' {
				if query[i] == '"' || query[i] == '\'' {
					quote := query[i]
					cur.WriteByte(query[i])
					i++
					for i < n && query[i] != quote {
						cur.WriteByte(query[i])
						i++
					}
					if i >= n {
						return nil, fmt.Errorf("unterminated string in selector")
					}
					cur.WriteByte(query[i])
					i++
				} else {
					cur.WriteByte(query[i])
					i++
				}
			}
			if i >= n {
				return nil, fmt.Errorf("unterminated attribute selector")
			}
			cur.WriteByte(query[i])
			i++
		default:
			cur.WriteByte(c)
			i++
		}
	}
	flush()
	return tokens, nil
}

func parseSimpleSelector(s string) (*SimpleSelector, error) {
	simple := &SimpleSelector{}
	n := len(s)
	if n == 0 {
		return nil, fmt.Errorf("invalid simple selector: %s", s)
	}

	readIdent := func(i int) (string, int) {
		start := i
		for i < n {
			c := s[i]
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' {
				i++
				continue
			}
			break
		}
		return s[start:i], i
	}

	hasToken := false
	for i := 0; i < n; {
		c := s[i]
		switch c {
		case '*':
			simple.IsUniversal = true
			hasToken = true
			i++
		case '#':
			i++
			id, next := readIdent(i)
			if id == "" {
				return nil, fmt.Errorf("invalid id in selector: %s", s)
			}
			simple.ID = id
			hasToken = true
			i = next
		case '.':
			i++
			cls, next := readIdent(i)
			if cls == "" {
				return nil, fmt.Errorf("invalid class in selector: %s", s)
			}
			simple.Classes = append(simple.Classes, cls)
			hasToken = true
			i = next
		case '[':
			i++
			for i < n && (s[i] == ' ' || s[i] == '\t') {
				i++
			}
			name, next := readIdent(i)
			if name == "" {
				return nil, fmt.Errorf("invalid attribute name in selector: %s", s)
			}
			i = next
			for i < n && (s[i] == ' ' || s[i] == '\t') {
				i++
			}
			attr := AttrSelector{Name: name}
			if i < n && s[i] != ']' {
				if i+1 < n && (s[i] == '~' || s[i] == '|' || s[i] == '^' || s[i] == '$' || s[i] == '*') && s[i+1] == '=' {
					attr.Op = s[i : i+2]
					i += 2
				} else if s[i] == '=' {
					attr.Op = "="
					i++
				} else {
					return nil, fmt.Errorf("invalid attribute operator in selector: %s", s)
				}
				for i < n && (s[i] == ' ' || s[i] == '\t') {
					i++
				}
				if i >= n {
					return nil, fmt.Errorf("missing attribute value in selector: %s", s)
				}
				if s[i] == '"' || s[i] == '\'' {
					quote := s[i]
					i++
					vstart := i
					for i < n && s[i] != quote {
						i++
					}
					if i >= n {
						return nil, fmt.Errorf("unterminated string in selector: %s", s)
					}
					attr.Value = s[vstart:i]
					i++
				} else {
					vstart := i
					for i < n && s[i] != ']' && s[i] != ' ' && s[i] != '\t' {
						i++
					}
					attr.Value = s[vstart:i]
				}
				for i < n && (s[i] == ' ' || s[i] == '\t') {
					i++
				}
			}
			if i >= n || s[i] != ']' {
				return nil, fmt.Errorf("unterminated attribute selector: %s", s)
			}
			i++
			simple.Attrs = append(simple.Attrs, attr)
			hasToken = true
		default:
			if hasToken {
				return nil, fmt.Errorf("invalid character %q in selector: %s", c, s)
			}
			tag, next := readIdent(i)
			if tag == "" {
				return nil, fmt.Errorf("invalid simple selector: %s", s)
			}
			simple.Tag = tag
			hasToken = true
			i = next
		}
	}

	if !hasToken {
		return nil, fmt.Errorf("invalid simple selector: %s", s)
	}
	return simple, nil
}
