package util

import (
	"strings"

	"github.com/ventu-io/go-shortid"
)

var (
	tagGenerator  *shortid.Shortid
)

func init() {
	tagGenerator = shortid.MustNew(1, shortid.DefaultABC, 2342)
}

func GenerateTag() string {
	// TODO - verify tag not already in use!!!
	return tagGenerator.MustGenerate()
}

func IsTag(value string) bool {
	if len(value) != 9 {
		return false
	}

	for _, chr := range value {
		if strings.IndexRune(shortid.DefaultABC, chr) < 0 {
			return false
		}
	}
	return true
}