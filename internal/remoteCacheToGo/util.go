package util

import (
	"unicode"
)

func CharacterWhiteList(input string) bool {
    for _, r := range input {
        if unicode.IsLetter(r) || unicode.IsNumber(r) {
            return true
        }
    }
    return false
}
