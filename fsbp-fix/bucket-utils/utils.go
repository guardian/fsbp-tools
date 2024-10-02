package utils

import (
	"strings"

	"github.com/guardian/fsbp-tools/common"
)

func SplitAndTrim(str string) []string {
	split := strings.Split(str, ",")
	var trimmed []string
	for _, s := range split {
		s := strings.Trim(s, " ")
		trimmed = append(trimmed, s)
	}

	return common.Complement(trimmed, []string{""})
}
