package iplist

import "strings"

const (
	separator = ","
)

func ToString(strs ...string) string {
	return strings.Join(strs, separator)
}

func FromString(str string) []string {
	return strings.Split(str, separator)
}
