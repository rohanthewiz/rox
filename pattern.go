package rox

import (
	"fmt"
	"regexp"
	"strings"
	"unsafe"
)

// Pattern a parsed representation of path pattern.
// eg github.com/googleapis/googleapis/google/api/http.proto.
type Pattern struct {
	key     string   // the key value used for trie.
	fields  []string // list of fields names to be bound by this pattern
	verb    string   // the tail static part in the pattern,eg VERB of URL path.
	pattern string   // original pattern (example: /v1/users/{id})
}

// NewPattern creates a default style's new Pattern from the given original pattern.
// "regexps" is a list of regular expressions shared between multiple patterns.
//
// The syntax of the pattern string is as follows:
//
// 	Pattern		= "/" Segments
// 	Segments	= Segment { "/" Segment }
// 	Segment		= LITERAL | Parameter
//	Parameter	= Anonymous | Named
//	Anonymous	= ":" | "*"
//	Named		= ":" FieldPath [ "=" Regexp ] | "*" FieldPath
// 	FieldPath	= IDENT { "." IDENT }
//
func NewPattern(pattern string, regexps *[]*regexp.Regexp) (p Pattern, err error) {
	var fields []string
	kbuilder := make([]byte, 0, len(pattern))
	segments := pattern

	prevChar := byte(0)
	c := byte(0)
	for i := 0; i < len(segments); i, prevChar = i+1, c {
		c = segments[i]
		kbuilder = append(kbuilder, c)

		if prevChar != '/' {
			continue
		}
		if c == '/' {
			err = fmt.Errorf("pattern include empty segment - %q", segments)
			return
		}

		if c == ':' { // named parameter
			m := strings.IndexByte(segments[i:], '/')
			var nameAndRe string
			if m < 0 { // last part
				nameAndRe = segments[i+1:]
				i = len(segments) - 1
			} else {
				nameAndRe = segments[i+1 : i+m]
				i = i + m - 1 // for i++
			}

			reSep := strings.IndexByte(nameAndRe, '=') // Search for a name/regexp separator.
			if reSep < 0 {                             // only name
				fields = append(fields, nameAndRe)
			} else {
				fields = append(fields, nameAndRe[:reSep])
				expr := nameAndRe[reSep+1:]
				if expr == "" {
					err = fmt.Errorf("pattern has empty regular expression - %q", segments)
					return
				}
				rec := -1 // regular expression keychar
				for j, re := range *regexps {
					if re.String() == expr {
						rec = j
						break
					}
				}
				if rec == -1 { // regular expression not exist
					var re *regexp.Regexp
					if re, err = regexp.Compile(expr); err != nil {
						err = fmt.Errorf("pattern has invalid regular expression - %q", segments)
						return
					}
					rec = len(*regexps)
					*regexps = append(*regexps, re)
				}

				kbuilder = append(kbuilder, '=', byte(rec))
			}
		} else if c == '*' { // wildcard parameter
			m := strings.IndexByte(segments[i:], '/')
			if m > 0 {
				err = fmt.Errorf("'*' in pattern must is last segment - %q", segments)
				return
			}
			fields = append(fields, segments[i+1:])
			i = len(segments) - 1
		}
	}

	return Pattern{
		key:     *(*string)(unsafe.Pointer(&kbuilder)),
		fields:  fields,
		pattern: pattern,
	}, nil
}

// MustPattern is a helper function which makes it easier to call NewPattern in variable initialization.
func MustPattern(p Pattern, err error) Pattern {
	if err != nil {
		panic(fmt.Sprintf("Pattern initialization failed: %v", err))
	}
	return p
}

// Key returns the key value of the pattern.
func (p Pattern) Key() string { return p.key }

// NumField returns a pattern's field count.
func (p Pattern) NumField() int { return len(p.fields) }

// Field returns a pattern's i'th field name.
func (p Pattern) Field(i int) string { return p.fields[i] }

// Verb returns the VERB part of the path pattern. It is empty if the pattern does not have VERB part.
func (p Pattern) Verb() string { return p.verb }

// Pattern returns the original pattern (example: /v1/users/{id})
func (p Pattern) Pattern() string { return p.pattern }

func splitURLPath(path string) (segments, verb string) {
	for i := len(path) - 1; i >= 0 && path[i] != '/'; i-- {
		if path[i] == ':' {
			return path[:i], path[i:]
		}
	}
	return path, ""
}
