package firewall

import (
	"regexp"
	"strings"
)

type InputType string

// decalre const instead of writing "string" in the switch case, might change other switch statements to this for security
const (
	TypeUsername InputType = "username"
	TypePassword InputType = "password"
	TypePath     InputType = "path"
	TypeURL      InputType = "url"
)

// Handy for testing https://regex101.com/
var (
	UsernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)
	PasswordPattern = regexp.MustCompile(`^[a-zA-Z0-9._:/?#@!$&'()*+,;=-]{12,64}$`)
	UrPattern       = regexp.MustCompile(`^([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)
	PathPattern     = regexp.MustCompile(`^[a-zA-Z0-9\s._~:/?#[\]@!$&'()*+,;=-]*$`)
)

func ValidateInput(input string, inputType InputType) bool {
	// Check for null bytes
	if strings.Contains(input, "\x00") {
		return false
	}
	//no escaping out of the static folder
	if strings.Contains(input, "..") {
		return false
	}

	switch inputType {
	case TypeUsername:
		return UsernamePattern.MatchString(input)
	case TypePassword:
		return PasswordPattern.MatchString(input)
	case TypePath:
		return PathPattern.MatchString(input)
	case TypeURL:
		return UrPattern.MatchString(input)
	default:
		return false
	}
}
