package src

import "fmt"

type ParseError struct {
	Kind string
	Message string
	Offset int
	Line int
	Column int
	Context string
}

func (e *ParseError) Error() string {
	if e==nil {
		return "<nil>"
	}
	base:=fmt.Sprintf("parse error [%s] at %d:%d (offset %d): %s", e.Kind, e.Line, e.Column, e.Offset, e.Message)
	if e.Context!="" {
		return base+" | context: `"+e.Context+"`"
	}
	return base
}

type position struct {
	offset int
	line int
	column int
}

func makeParseError(kind string, msg string, pos position, input string) *ParseError {
	return &ParseError{
		Kind: kind,
		Message: msg,
		Offset: pos.offset,
		Line: pos.line,
		Column: pos.column,
		Context: extractContext(input, pos.offset, 24),
	}
}

func extractContext(input string, offset int, radius int) string {
	if offset<0 {
		offset=0
	}
	if offset>len(input) {
		offset=len(input)
	}

	start:=offset-radius
	if start<0 {
		start=0
	}
	end:=offset+radius
	if end>len(input) {
		end=len(input)
	}
	return input[start:end]
}