package valueobject

import (
	"regexp"
	"strings"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Email is an immutable value object that guarantees a valid email format.
type Email struct {
	value string
}

// NewEmail creates a validated Email value object.
func NewEmail(raw string) (Email, error) {
	normalized := strings.TrimSpace(strings.ToLower(raw))
	if normalized == "" || !emailRegex.MatchString(normalized) {
		return Email{}, domain.ErrInvalidEmail
	}
	return Email{value: normalized}, nil
}

// String returns the email as a string.
func (e Email) String() string {
	return e.value
}
