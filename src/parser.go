package src

func Parse(input string) (*Node, error) {
	root := &Node{Type: DocumentNode}
	stack := []*Node{root}
	lex := newTokenizer(input)

	for {
		tok := lex.nextToken()

		switch tok.kind {
		case tokenEOF:
			return root, nil
		case tokenText:
			if tok.data == "" {
				continue
			}
			stack[len(stack)-1].AppendChild(&Node{Type: TextNode, Data: tok.data})
		case tokenComment:
			stack[len(stack)-1].AppendChild(&Node{Type: CommentNode, Data: tok.data})
		case tokenDoctype:
			stack[len(stack)-1].AppendChild(&Node{Type: DoctypeNode, Data: tok.data})
		case tokenStartTag:
			seenAttrs := make(map[string]struct{}, len(tok.attrs))
			normalizedAttrs := make([]Attribute, 0, len(tok.attrs))
			for _, attr := range tok.attrs {
				if attr.Name == "" {
					continue
				}
				if _, ok := seenAttrs[attr.Name]; ok {
					continue
				}
				seenAttrs[attr.Name] = struct{}{}
				normalizedAttrs = append(normalizedAttrs, attr)
			}

			element := &Node{Type: ElementNode, Tag: tok.name, Attrs: normalizedAttrs}
			stack[len(stack)-1].AppendChild(element)
			if tok.selfClosing {
				continue
			}
			if _, ok := voidElements[tok.name]; ok {
				continue
			}
			stack = append(stack, element)
		case tokenEndTag:
			if len(stack) <= 1 {
				continue
			}
			matchIdx := -1
			for i := len(stack) - 1; i >= 1; i-- {
				if stack[i].Tag == tok.name {
					matchIdx = i
					break
				}
			}
			if matchIdx == -1 {
				continue
			}
			stack = stack[:matchIdx]
		}
	}
}

var voidElements = map[string]struct{}{
	"area":   {},
	"base":   {},
	"br":     {},
	"col":    {},
	"embed":  {},
	"hr":     {},
	"img":    {},
	"input":  {},
	"link":   {},
	"meta":   {},
	"param":  {},
	"source": {},
	"track":  {},
	"wbr":    {},
}