package parser

func Parse(input string) (*Node, error) {
	root:=&Node{Type: DocumentNode}
	stack:=[]*Node{root}
	lex:=newTokenizer(input)

	for {
		tok, err:=lex.nextToken()
		if err!=nil {
			return nil, err
		}

		switch tok.kind {
		case tokenEOF:
			if len(stack)>1 {
				open:=stack[len(stack)-1]
				return nil, makeParseError("unclosed_tag", "unclosed tag at EOF: <"+open.Tag+">", tok.pos, input)
			}
			return root, nil
		case tokenText:
			if tok.data=="" {
				continue
			}
			stack[len(stack)-1].AppendChild(&Node{Type: TextNode, Data: tok.data})
		case tokenComment:
			stack[len(stack)-1].AppendChild(&Node{Type: CommentNode, Data: tok.data})
		case tokenDoctype:
			stack[len(stack)-1].AppendChild(&Node{Type: DoctypeNode, Data: tok.data})
		case tokenStartTag:
			if tok.selfClosing && !isVoidElement(tok.name) {
				return nil, makeParseError("invalid_self_closing", "non-void element cannot be self-closing: <"+tok.name+"/>", tok.pos, input)
			}
			dup:=findDuplicateAttribute(tok.attrs)
			if dup!="" {
				return nil, makeParseError("duplicate_attribute", "duplicate attribute in element <"+tok.name+">: "+dup, tok.pos, input)
			}
			element:=&Node{Type: ElementNode, Tag: tok.name, Attrs: tok.attrs}
			stack[len(stack)-1].AppendChild(element)
			if tok.selfClosing || isVoidElement(tok.name) {
				continue
			}
			stack=append(stack, element)
		case tokenEndTag:
			if len(stack)<=1 {
				return nil, makeParseError("unmatched_closing_tag", "closing tag without matching opening tag: </"+tok.name+">", tok.pos, input)
			}
			top:=stack[len(stack)-1]
			if top.Tag!=tok.name {
				return nil, makeParseError("mismatched_closing_tag", "closing tag mismatch: expected </"+top.Tag+"> but got </"+tok.name+">", tok.pos, input)
			}
			stack=stack[:len(stack)-1]
		}
	}
}

func findDuplicateAttribute(attrs []Attribute) string {
	seen:=make(map[string]struct{}, len(attrs))
	for _, attr:=range attrs {
		if attr.Name=="" {
			continue
		}
		if _, ok:=seen[attr.Name]; ok {
			return attr.Name
		}
		seen[attr.Name]=struct{}{}
	}
	return ""
}

func isVoidElement(tag string) bool {
	_, ok:=voidElements[tag]
	return ok
}

var voidElements=map[string]struct{}{
	"area": {},
	"base": {},
	"br": {},
	"col": {},
	"embed": {},
	"hr": {},
	"img": {},
	"input": {},
	"link": {},
	"meta": {},
	"param": {},
	"source": {},
	"track": {},
	"wbr": {},
}