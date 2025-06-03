package helper

import "regexp"

// MatchPattern memeriksa apakah string cocok dengan pola regex yang diberikan
func MatchPattern(str string, pattern string) bool {
	match, err := regexp.MatchString(pattern, str)
	if err != nil {
		return false
	}
	return match
}
