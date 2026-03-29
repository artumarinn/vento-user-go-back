package valueobject

import (
	"unicode"

	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// Password is an immutable value object holding a bcrypt-hashed password.
type Password struct {
	hash string
}

// NewPassword creates a hashed Password from a plain-text password.
func NewPassword(plain string) (Password, error) {
	if !isPasswordStrong(plain) {
		return Password{}, domain.ErrWeakPassword
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return Password{}, err
	}
	return Password{hash: string(hashed)}, nil
}

// isPasswordStrong validates complexity: min 8 chars, 1 uppercase, 1 number.
func isPasswordStrong(password string) bool {
	if len(password) < 8 {
		return false
	}
	var hasUpper, hasNumber bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsNumber(char):
			hasNumber = true
		}
	}
	return hasUpper && hasNumber
}

// NewPasswordFromHash creates a Password from an already-hashed string (from DB).
func NewPasswordFromHash(hash string) Password {
	return Password{hash: hash}
}

// Verify checks if a plain-text password matches the stored hash.
func (p Password) Verify(plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(p.hash), []byte(plain)) == nil
}

// Hash returns the bcrypt hash string (for persistence).
func (p Password) Hash() string {
	return p.hash
}
