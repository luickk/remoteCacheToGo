package util

import (
	"unicode"
)

func CharacterWhiteList(input string) bool {
    for _, r := range input {
        if !unicode.IsLetter(r) {
            return false
        }
    }
    return true
}
