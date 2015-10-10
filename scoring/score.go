package scoring

import "strings"

func Score(suggestion, suggestionLower, partial, partialLower string) int {
	for i := len(partial); i > 0; i-- {
		c := 0
		if strings.Contains(suggestion, partial[:i]) {
			c = i*i + 1
		} else if strings.Contains(suggestionLower, partialLower[:i]) {
			c = i * i
		}
		if c > 0 {
			return c + Score(suggestion, suggestionLower, partial[i:], partialLower[i:])
		}
	}
	return 0
}
