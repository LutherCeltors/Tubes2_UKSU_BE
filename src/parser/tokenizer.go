package parser

import (
	"strings"
	"unicode/utf8"
)

type tokenType int

const (
	tokenStartTag tokenType=iota
	tokenEndTag
	tokenText
	tokenComment
	tokenDoctype
	tokenEOF
)

type token struct {
	kind tokenType
	name string
	data string
	attrs []Attribute
	selfClosing bool
	pos position
}

type tokenizer struct {
	input string
	length int
	pos int
	line int
	column int
}

func newTokenizer(input string) *tokenizer {
	t:=&tokenizer{
		input: input,
		length: len(input),
		line: 1,
		column: 1,
	}
	return t
}

func (t *tokenizer) nextToken() (token, error) {
	if t.pos>=t.length {
		return token{kind: tokenEOF, pos: t.currentPos()}, nil
	}

	if t.peekByte()=='<' {
		if t.hasPrefix("<!--") {
			return t.readComment()
		}
		if t.hasPrefix("<!") {
			return t.readDeclaration()
		}
		if t.hasPrefix("</") {
			return t.readEndTag()
		}

		if isNameStartChar(t.peekByteAt(1)) {
			return t.readStartTag()
		}

		return token{}, makeParseError("invalid_tag_open", "invalid tag opening after '<'", t.currentPos(), t.input)
	}

	return t.readText(), nil
}

func (t *tokenizer) readText() token {
	start:=t.currentPos()
	begin:=t.pos
	for (t.pos<t.length && t.peekByte()!='<') {
		t.advanceByte()
	}
	return token{kind: tokenText, data: decodeHTMLEntities(t.input[begin:t.pos]), pos: start}
}

func (t *tokenizer) readComment() (token, error) {
	start:=t.currentPos()
	t.advanceN(4)
	begin:=t.pos
	idx:=strings.Index(t.input[t.pos:], "-->")
	if idx==-1 {
		return token{}, makeParseError("unclosed_comment", "comment is not closed with '-->'", start, t.input)
	}
	end:=t.pos+idx
	data:=t.input[begin:end]
	t.advanceN(idx+3)
	return token{kind: tokenComment, data: data, pos: start}, nil
}

func (t *tokenizer) readDeclaration() (token, error) {
	start:=t.currentPos()
	t.advanceN(2)
	begin:=t.pos
	for (t.pos<t.length && t.peekByte()!='>') {
		t.advanceByte()
	}
	if t.pos>=t.length {
		return token{}, makeParseError("unclosed_declaration", "declaration/doctype is not closed with '>'", start, t.input)
	}
	data:=strings.TrimSpace(t.input[begin:t.pos])
	t.advanceByte()

	if !hasPrefixFold(data, "doctype") {
		return token{}, makeParseError("invalid_declaration", "invalid declaration: expected <!DOCTYPE ...>", start, t.input)
	}
	data=strings.TrimSpace(data[7:])
	if data=="" {
		return token{}, makeParseError("invalid_declaration", "doctype declaration must include content", start, t.input)
	}
	return token{kind: tokenDoctype, data: data, pos: start}, nil
}

func (t *tokenizer) readEndTag() (token, error) {
	start:=t.currentPos()
	t.advanceN(2)
	t.skipSpaces()
	if t.pos>=t.length {
		return token{}, makeParseError("unexpected_eof", "unexpected EOF while reading end tag", start, t.input)
	}
	if !isNameStartChar(t.peekByte()) {
		return token{}, makeParseError("malformed_end_tag", "end tag name is invalid", t.currentPos(), t.input)
	}
	name:=strings.ToLower(t.readName())
	t.skipSpaces()
	if t.pos>=t.length {
		return token{}, makeParseError("unexpected_eof", "unexpected EOF while closing end tag", start, t.input)
	}
	if t.peekByte()!='>' {
		return token{}, makeParseError("malformed_end_tag", "end tag has invalid trailing content", t.currentPos(), t.input)
	}
	t.advanceByte()
	return token{kind: tokenEndTag, name: name, pos: start}, nil
}

func (t *tokenizer) readStartTag() (token, error) {
	start:=t.currentPos()
	t.advanceByte()
	if t.pos>=t.length || !isNameStartChar(t.peekByte()) {
		return token{}, makeParseError("invalid_tag_name", "start tag name is invalid", t.currentPos(), t.input)
	}
	name:=strings.ToLower(t.readName())
	attrs:=make([]Attribute, 0)
	selfClosing:=false

	for {
		t.skipSpaces()
		if t.pos>=t.length {
			return token{}, makeParseError("unexpected_eof", "unexpected EOF while reading start tag", start, t.input)
		}
		if t.hasPrefix("/>") {
			selfClosing=true
			t.advanceN(2)
			break
		}
		if t.peekByte()=='>' {
			t.advanceByte()
			break
		}

		attrPos:=t.currentPos()
		attrName:=t.readAttributeName()
		if attrName=="" {
			return token{}, makeParseError("invalid_attribute", "attribute name is invalid", attrPos, t.input)
		}

		attrValue:=""
		t.skipSpaces()
		if t.pos<t.length && t.peekByte()=='=' {
			t.advanceByte()
			t.skipSpaces()
			v, err:=t.readAttributeValue()
			if err!=nil {
				return token{}, err
			}
			attrValue=v
		}
		attrs=append(attrs, Attribute{Name: strings.ToLower(attrName), Value: decodeHTMLEntities(attrValue)})
	}

	return token{kind: tokenStartTag, name: name, attrs: attrs, selfClosing: selfClosing, pos: start}, nil
}

func (t *tokenizer) readAttributeValue() (string, error) {
	if t.pos>=t.length {
		return "", makeParseError("unexpected_eof", "unexpected EOF while reading attribute value", t.currentPos(), t.input)
	}

	quote:=t.peekByte()
	if quote=='"' || quote=='\'' {
		start:=t.currentPos()
		t.advanceByte()
		begin:=t.pos
		for (t.pos<t.length && t.peekByte()!=quote) {
			t.advanceByte()
		}
		if t.pos>=t.length {
			return "", makeParseError("unclosed_attribute_value", "quoted attribute value is not closed", start, t.input)
		}
		value:=t.input[begin:t.pos]
		t.advanceByte()
		return value, nil
	}

	begin:=t.pos
	for (t.pos<t.length) {
		ch:=t.peekByte()
		if isSpace(ch) || ch=='>' || ch=='/' {
			break
		}
		t.advanceByte()
	}
	return t.input[begin:t.pos], nil
}

func (t *tokenizer) readAttributeName() string {
	begin:=t.pos
	for (t.pos<t.length) {
		ch:=t.peekByte()
		if isSpace(ch) || ch=='=' || ch=='>' || ch=='/' {
			break
		}
		t.advanceByte()
	}
	return strings.TrimSpace(t.input[begin:t.pos])
}

func (t *tokenizer) readName() string {
	begin:=t.pos
	for (t.pos<t.length && isNameChar(t.peekByte())) {
		t.advanceByte()
	}
	return t.input[begin:t.pos]
}

func (t *tokenizer) skipSpaces() {
	for (t.pos<t.length && isSpace(t.peekByte())) {
		t.advanceByte()
	}
}

func (t *tokenizer) hasPrefix(prefix string) bool {
	return strings.HasPrefix(t.input[t.pos:], prefix)
}

func hasPrefixFold(s string, prefix string) bool {
	if len(s)<len(prefix) {
		return false
	}
	return strings.EqualFold(s[:len(prefix)], prefix)
}

func (t *tokenizer) peekByte() byte {
	return t.input[t.pos]
}

func (t *tokenizer) peekByteAt(delta int) byte {
	idx:=t.pos+delta
	if idx<0 || idx>=t.length {
		return 0
	}
	return t.input[idx]
}

func (t *tokenizer) advanceN(n int) {
	for i:=0; i<n && t.pos<t.length; i++ {
		t.advanceByte()
	}
}

func (t *tokenizer) advanceByte() {
	if t.pos>=t.length {
		return
	}
	r, size:=utf8.DecodeRuneInString(t.input[t.pos:])
	if r==utf8.RuneError && size==0 {
		return
	}
	if size<=0 {
		size=1
	}
	t.pos+=size
	if r=='\n' {
		t.line++
		t.column=1
	} else {
		t.column++
	}
}

func (t *tokenizer) currentPos() position {
	return position{offset: t.pos, line: t.line, column: t.column}
}

func isNameStartChar(b byte) bool {
	return (b>='a' && b<='z') || (b>='A' && b<='Z')
}

func isNameChar(b byte) bool {
	return isNameStartChar(b) || (b>='0' && b<='9') || b=='-' || b=='_' || b==':' || b=='.'
}

func isSpace(b byte) bool {
	switch b {
	case ' ', '\n', '\r', '\t', '\f':
		return true
	default:
		return false
	}
}