package scrapers

import (
	"strings"
)

type CleanFunc func(string) string

func NewPipe(cleanFuncs ...CleanFunc) func(string) string {
	return func(str string) string {
		return CleanString(str, cleanFuncs...)
	}
}

func CleanString(str string, cleanFuncs ...CleanFunc) string {
	cleaned := str
	for _, clean := range cleanFuncs {
		cleaned = clean(cleaned)
	}

	return cleaned
}

func RemovePrefix(prefix string) CleanFunc {
	return func(str string) string {
		if strings.HasPrefix(str, prefix) {
			return str[len(prefix):]
		}
		return str
	}
}

func OneLine(str string) string {
	return strings.Replace(str, "\n", " ", -1)
}
