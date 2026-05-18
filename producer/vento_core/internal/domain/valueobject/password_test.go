package valueobject_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain/valueobject"
)

func TestNewPassword_Strong(t *testing.T) {
	p, err := valueobject.NewPassword("Strong1Pass")
	require.NoError(t, err)
	assert.NotEmpty(t, p.Hash())
	assert.True(t, p.Verify("Strong1Pass"))
	assert.False(t, p.Verify("wrongpassword"))
}

func TestNewPassword_Weak(t *testing.T) {
	cases := []string{
		"short1A",   // < 8 chars
		"alllowercase1", // no uppercase
		"ALLUPPERCASE",  // no digit
		"NoNumber!",     // no digit
	}
	for _, tc := range cases {
		_, err := valueobject.NewPassword(tc)
		assert.ErrorIs(t, err, domain.ErrWeakPassword, "input: %q", tc)
	}
}

func TestNewPasswordFromHash(t *testing.T) {
	p1, err := valueobject.NewPassword("Valid1Pass")
	require.NoError(t, err)
	p2 := valueobject.NewPasswordFromHash(p1.Hash())
	assert.Equal(t, p1.Hash(), p2.Hash())
	assert.True(t, p2.Verify("Valid1Pass"))
}
