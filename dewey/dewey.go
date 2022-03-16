package dewey

import "regexp"

var rxpDewey = regexp.MustCompile(`^\d{1,3}(.\d+)?$`)

func LooksLike(s string) bool {
	return rxpDewey.MatchString(s)
}
