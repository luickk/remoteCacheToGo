package util

import (
	"unicode"
)

// checks wether input consists only of characters
func CharacterWhiteList(input string) bool {
    for _, r := range input {
        if unicode.IsLetter(r) || unicode.IsNumber(r) {
            return true
        }
    }
    return false
}

// slice operation
// from https://stackoverflow.com/questions/8307478/how-to-find-out-element-position-in-slice
func Index(limit int, predicate func(i int) bool) int {
    for i := 0; i < limit; i++ {
        if predicate(i) {
            return i
        }
    }
    return -1
}
