package src

import (
	"strings"
	"unicode/utf8"
)

type tokenType int

const (
	tokenStartTag tokenType = iota
	tokenEndTag
	tokenText
	tokenComment
	tokenDoctype
	tokenEOF
)

type token struct {
	kind        tokenType
	name        string
	data        string
	attrs       []Attribute
	selfClosing bool
}

type tokenizer struct {
	input      string
	length     int
	pos        int
	rawtextTag string
}

func newTokenizer(input string) *tokenizer {
	return &tokenizer{
		input:  input,
		length: len(input),
	}
}

var rawTextElements = map[string]struct{}{
	"script":   {},
	"style":    {},
	"textarea": {},
	"title":    {},
	"iframe":   {},
	"noscript": {},
	"noframes": {},
	"noembed":  {},
	"xmp":      {},
	"plaintext": {},
}

func (t *tokenizer) nextToken() token {
	tok := t.nextTokenInternal()
	if tok.kind == tokenStartTag && !tok.selfClosing {
		if _, ok := rawTextElements[tok.name]; ok {
			t.rawtextTag = tok.name
		}
	}
	return tok
}

func (t *tokenizer) nextTokenInternal() token {
	if t.rawtextTag != "" {
		return t.readRawText()
	}
	if t.pos >= t.length {
		return token{kind: tokenEOF}
	}

	if t.peekByte() == '<' {
		start := t.pos

		if strings.HasPrefix(t.input[t.pos:], "<!--") {
			return t.readComment()
		}
		if strings.HasPrefix(t.input[t.pos:], "<![CDATA[") {
			return t.readCData()
		}
		if strings.HasPrefix(t.input[t.pos:], "<!") {
			return t.readDeclaration()
		}
		if strings.HasPrefix(t.input[t.pos:], "</") {
			tok, ok := t.readEndTag()
			if ok {
				return tok
			}
			t.pos = start + 1
			return token{kind: tokenText, data: "<"}
		}
		next := t.peekByteAt(1)
		if (next >= 'a' && next <= 'z') || (next >= 'A' && next <= 'Z') {
			tok, ok := t.readStartTag()
			if ok {
				return tok
			}
			t.pos = start + 1
			return token{kind: tokenText, data: "<"}
		}

		t.advanceByte()
		return token{kind: tokenText, data: "<"}
	}

	return t.readText()
}

func (t *tokenizer) readRawText() token {
	rawTag := t.rawtextTag
	begin := t.pos
	needle := "</" + rawTag

	search := t.pos
	for search < t.length {
		idx := indexFoldFrom(t.input, needle, search)
		if idx == -1 {
			data := t.input[begin:]
			t.pos = t.length
			t.rawtextTag = ""
			return token{kind: tokenText, data: data}
		}
		afterIdx := idx + len(needle)
		if afterIdx >= t.length {
			data := t.input[begin:idx]
			t.pos = idx
			t.rawtextTag = ""
			return token{kind: tokenText, data: data}
		}
		ch := t.input[afterIdx]
		if ch == ' ' || ch == '\n' || ch == '\r' || ch == '\t' || ch == '\f' || ch == '>' || ch == '/' {
			t.rawtextTag = ""
			if idx == begin {
				return t.nextTokenInternal()
			}
			data := t.input[begin:idx]
			t.pos = idx
			return token{kind: tokenText, data: data}
		}
		search = afterIdx
	}

	data := t.input[begin:]
	t.pos = t.length
	t.rawtextTag = ""
	return token{kind: tokenText, data: data}
}

func (t *tokenizer) readCData() token {
	t.advanceN(9)
	begin := t.pos
	idx := strings.Index(t.input[t.pos:], "]]>")
	if idx == -1 {
		data := t.input[begin:]
		t.pos = t.length
		return token{kind: tokenText, data: data}
	}
	end := t.pos + idx
	data := t.input[begin:end]
	t.pos = end + 3
	return token{kind: tokenText, data: data}
}

func indexFoldFrom(s, sub string, start int) int {
	if start < 0 {
		start = 0
	}
	if start > len(s) {
		return -1
	}
	idx := strings.Index(strings.ToLower(s[start:]), strings.ToLower(sub))
	if idx == -1 {
		return -1
	}
	return start + idx
}

func (t *tokenizer) readText() token {
	begin := t.pos
	for t.pos < t.length && t.peekByte() != '<' {
		t.advanceByte()
	}
	return token{
		kind: tokenText,
		data: decodeHTMLEntities(t.input[begin:t.pos]),
	}
}

func (t *tokenizer) readComment() token {
	t.advanceN(4)
	begin := t.pos
	idx := strings.Index(t.input[t.pos:], "-->")
	if idx == -1 {
		data := t.input[begin:]
		t.pos = t.length
		return token{kind: tokenComment, data: data}
	}
	end := t.pos + idx
	data := t.input[begin:end]
	t.pos = end + 3
	return token{kind: tokenComment, data: data}
}

func (t *tokenizer) readDeclaration() token {
	t.advanceN(2)
	begin := t.pos
	for t.pos < t.length && t.peekByte() != '>' {
		t.advanceByte()
	}

	data := strings.TrimSpace(t.input[begin:t.pos])
	if t.pos < t.length && t.peekByte() == '>' {
		t.advanceByte()
	}

	if len(data) < 7 || !strings.EqualFold(data[:7], "doctype") {
		return token{kind: tokenComment, data: data}
	}
	data = strings.TrimSpace(data[7:])
	return token{kind: tokenDoctype, data: data}
}

func (t *tokenizer) readEndTag() (token, bool) {
	start := t.pos
	t.advanceN(2)
	t.skipSpaces()
	if t.pos >= t.length {
		t.pos = start
		return token{}, false
	}
	first := t.peekByte()
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z')) {
		t.pos = start
		return token{}, false
	}

	name := strings.ToLower(t.readName())
	t.skipSpaces()
	if t.pos >= t.length || t.peekByte() != '>' {
		t.pos = start
		return token{}, false
	}
	t.advanceByte()
	return token{kind: tokenEndTag, name: name}, true
}

func (t *tokenizer) readStartTag() (token, bool) {
	start := t.pos
	t.advanceByte()
	if t.pos >= t.length {
		t.pos = start
		return token{}, false
	}
	first := t.peekByte()
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z')) {
		t.pos = start
		return token{}, false
	}

	name := strings.ToLower(t.readName())
	attrs := make([]Attribute, 0)
	selfClosing := false

	for {
		t.skipSpaces()
		if t.pos >= t.length {
			return token{
				kind:        tokenStartTag,
				name:        name,
				attrs:       attrs,
				selfClosing: selfClosing,
			}, true
		}
		if strings.HasPrefix(t.input[t.pos:], "/>") {
			selfClosing = true
			t.advanceN(2)
			break
		}
		if t.peekByte() == '>' {
			t.advanceByte()
			break
		}

		attrName := t.readAttributeName()
		if attrName == "" {
			if t.pos < t.length {
				t.advanceByte()
			}
			continue
		}

		attrValue := ""
		t.skipSpaces()
		if t.pos < t.length && t.peekByte() == '=' {
			t.advanceByte()
			t.skipSpaces()
			attrValue = t.readAttributeValue()
		}

		attrs = append(attrs, Attribute{
			Name:  strings.ToLower(attrName),
			Value: decodeHTMLEntities(attrValue),
		})
	}

	return token{
		kind:        tokenStartTag,
		name:        name,
		attrs:       attrs,
		selfClosing: selfClosing,
	}, true
}

func (t *tokenizer) readAttributeValue() string {
	if t.pos >= t.length {
		return ""
	}

	quote := t.peekByte()
	if quote == '"' || quote == '\'' {
		t.advanceByte()
		begin := t.pos
		for t.pos < t.length && t.peekByte() != quote {
			t.advanceByte()
		}
		value := t.input[begin:t.pos]
		if t.pos < t.length && t.peekByte() == quote {
			t.advanceByte()
		}
		return value
	}

	begin := t.pos
	for t.pos < t.length {
		ch := t.peekByte()
		if ch == ' ' || ch == '\n' || ch == '\r' || ch == '\t' || ch == '\f' || ch == '>' || ch == '/' {
			break
		}
		t.advanceByte()
	}
	return t.input[begin:t.pos]
}

func (t *tokenizer) readAttributeName() string {
	begin := t.pos
	for t.pos < t.length {
		ch := t.peekByte()
		if ch == ' ' || ch == '\n' || ch == '\r' || ch == '\t' || ch == '\f' || ch == '=' || ch == '>' || ch == '/' || ch == '<' {
			break
		}
		t.advanceByte()
	}
	return strings.TrimSpace(t.input[begin:t.pos])
}

func (t *tokenizer) readName() string {
	begin := t.pos
	for t.pos < t.length {
		ch := t.peekByte()
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '_' || ch == ':' || ch == '.' {
			t.advanceByte()
			continue
		}
		break
	}
	return t.input[begin:t.pos]
}

func (t *tokenizer) skipSpaces() {
	for t.pos < t.length {
		ch := t.peekByte()
		if ch != ' ' && ch != '\n' && ch != '\r' && ch != '\t' && ch != '\f' {
			break
		}
		t.advanceByte()
	}
}

func (t *tokenizer) peekByte() byte {
	return t.input[t.pos]
}

func (t *tokenizer) peekByteAt(delta int) byte {
	idx := t.pos + delta
	if idx < 0 || idx >= t.length {
		return 0
	}
	return t.input[idx]
}

func (t *tokenizer) advanceN(n int) {
	for i := 0; i < n && t.pos < t.length; i++ {
		t.advanceByte()
	}
}

func (t *tokenizer) advanceByte() {
	if t.pos >= t.length {
		return
	}
	_, size := utf8.DecodeRuneInString(t.input[t.pos:])
	if size <= 0 {
		size = 1
	}
	t.pos += size
}