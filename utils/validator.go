package utils

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// GetErrorMessage create a list with all wrong fields
func GetErrorMessage(err error) string {
	var errFields = ""
	for _, err := range err.(validator.ValidationErrors) {
		errFields += fmt.Sprintf(" %s", strings.ToLower(err.Field()))
	}
	return errFields
}
