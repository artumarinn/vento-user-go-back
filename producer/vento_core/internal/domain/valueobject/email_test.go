package valueobject_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/valueobject"
)

func TestNewEmail_Valid(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"user@example.com", "user@example.com"},
		{"  USER@EXAMPLE.COM  ", "user@example.com"},
		{"a.b+tag@sub.domain.io", "a.b+tag@sub.domain.io"},
	}
	for _, tc := range cases {
		e, err := valueobject.NewEmail(tc.input)
		require.NoError(t, err, "input: %q", tc.input)
		assert.Equal(t, tc.expected, e.String())
	}
}

func TestNewEmail_Invalid(t *testing.T) {
	cases := []string{"", "   ", "notanemail", "@no-user.com", "no-at-sign", "missing@"}
	for _, tc := range cases {
		_, err := valueobject.NewEmail(tc)
		assert.ErrorIs(t, err, domain.ErrInvalidEmail, "input: %q", tc)
	}
}
