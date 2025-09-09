package text

import (
	"regexp"
	"strings"
)

var tokenRe = regexp.MustCompile(`\pL+\p{M}*|\pN+`)

var stop = map[string]struct{}{
	"the": {}, "a": {}, "an": {}, "and": {}, "or": {}, "to": {}, "of": {},
	"in": {}, "on": {}, "for": {}, "with": {}, "is": {}, "it": {}, "this": {}, "that": {},
}

func Tokenize(s string) []string {
	s = strings.ToLower(s)
	raw := tokenRe.FindAllString(s, -1)
	out := make([]string, 0, len(raw))
	for _, t := range raw {
		if len(t) <= 1 {
			continue
		}
		if _, bad := stop[t]; bad {
			continue
		}
		out = append(out, t)
	}
	return out
}
