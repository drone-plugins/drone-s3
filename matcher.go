package main

import (
	"regexp"
)

func matcher(match string, stringMap map[string]string) string {
	for pattern := range stringMap {
		matched, err := regexp.MatchString(pattern, match)

		if err != nil {
			panic(err)
		}

		if matched {
			return stringMap[pattern]
		}
	}

	return ""
}
