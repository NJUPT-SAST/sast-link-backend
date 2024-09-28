package validator

import "regexp"

// ValidateURL validates the url
//
// Return true if the url is valid.
func ValidateURL(url string) bool {
	compileRegex := regexp.MustCompile("[0-9]+")
	matchArr := compileRegex.FindAllString(url, -1)
	return matchArr != nil
}
