package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// IsValidUUID returns true if s is a valid UUID (RFC 4122 format).
func IsValidUUID(s string) bool {
	return uuidRegex.MatchString(s)
}

// GetErrorMessage create a list with all wrong fields
func GetErrorMessage(err error) string {
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return err.Error()
	}
	var errFields = ""
	for _, ve := range validationErrors {
		errFields += fmt.Sprintf(" %s", strings.ToLower(ve.Field()))
	}
	return errFields
}
