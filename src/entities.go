package src

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

func decodeHTMLEntities(s string) string {
	if !strings.Contains(s, "&") {
		return s
	}

	var builder strings.Builder
	for i:=0; i<len(s); {
		if s[i]!='&' {
			builder.WriteByte(s[i])
			i++
			continue
		}

		semiRel:=strings.IndexByte(s[i+1:], ';')
		if semiRel==-1 {
			builder.WriteByte(s[i])
			i++
			continue
		}

		semi:=i+1+semiRel
		entity:=s[i+1:semi]
		decoded, ok:=decodeEntity(entity)
		if !ok {
			builder.WriteString(s[i:semi+1])
		} else {
			builder.WriteString(decoded)
		}
		i=semi+1
	}

	return builder.String()
}

func decodeEntity(entity string) (string, bool) {
	if entity=="" {
		return "", false
	}

	if entity[0]=='#' {
		return decodeNumericEntity(entity[1:])
	}

	value, ok:=namedEntities[strings.ToLower(entity)]
	return value, ok
}

func decodeNumericEntity(raw string) (string, bool) {
	if raw=="" {
		return "", false
	}

	base:=10
	if len(raw)>1 && (raw[0]=='x' || raw[0]=='X') {
		base=16
		raw=raw[1:]
		if raw=="" {
			return "", false
		}
	}

	val, err:=strconv.ParseInt(raw, base, 32)
	if err!=nil || val<0 || val>int64(unicode.MaxRune) {
		return "", false
	}

	r:=rune(val)
	if !utf8.ValidRune(r) {
		return "", false
	}

	return string(r), true
}

var namedEntities=map[string]string{
	"amp": "&",
	"lt": "<",
	"gt": ">",
	"quot": "\"",
	"apos": "'",
	"nbsp": " ",
}