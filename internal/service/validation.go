package service

import (
	"regexp"
	"strings"
)

var plateRegexp = regexp.MustCompile(`^[ABEKMHOPCTYXАВЕКМНОРСТУХ]{1}\d{3}[ABEKMHOPCTYXАВЕКМНОРСТУХ]{2}\d{2,3}$`)

func NormalizePlate(input string) string {
	upper := strings.ToUpper(strings.TrimSpace(input))
	return upper
}

func ValidatePlate(input string) error {
	value := NormalizePlate(input)
	if !plateRegexp.MatchString(value) {
		return ErrInvalidPlate
	}
	return nil
}

func ValidateRole(role string) error {
	switch role {
	case "admin", "guard", "resident":
		return nil
	default:
		return ErrInvalidInput
	}
}
