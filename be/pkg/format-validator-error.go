package pkg

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

// FormatValidationErrors maps validator errors to `{ json_field: VALIDATION_<TAG> }`.
// The frontend translates `VALIDATION_<TAG>` codes, e.g. VALIDATION_REQUIRED, VALIDATION_MAX.
func FormatValidationErrors(err error) map[string]string {
	fields := map[string]string{}

	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, fe := range ve {
			fields[toSnake(fe.Field())] = "VALIDATION_" + strings.ToUpper(fe.Tag())
		}
		return fields
	}

	fields["body"] = "VALIDATION_INVALID_BODY"
	return fields
}

func toSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(r + ('a' - 'A'))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
