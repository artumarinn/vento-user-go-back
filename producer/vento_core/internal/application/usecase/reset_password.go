package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/vento-ai/shared/logger"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/dto"
	"github.com/vento-ai/vento-user-go-back/producer/vento_core/internal/application/port"
)

type ResetPasswordUsecases struct {
	userRepo port.UserRepository
}

func NewResetPasswordUsecases(userRepo port.UserRepository) *ResetPasswordUsecases {
	return &ResetPasswordUsecases{userRepo: userRepo}
}

// ForgotPassword generates a reset token and "sends" the email (logs it).
func (uc *ResetPasswordUsecases) ForgotPassword(ctx context.Context, req dto.ForgotPasswordRequest) error {
	user, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		// We return nil to avoid leaking user existence, but we log the attempt.
		logger.L().Warn("Forgot password attempt for non-existent email", "email", req.Email)
		return nil
	}

	token := generateRandomToken()
	expiresAt := time.Now().Add(1 * time.Hour)

	if err := uc.userRepo.SaveResetToken(ctx, user.Email.String(), token, expiresAt); err != nil {
		return err
	}

	// MOCK EMAIL: In a real system, we would call an EmailPort.
	// For Arturo, we log it clearly so he can use it.
	resetLink := fmt.Sprintf("http://localhost:3000/auth/reset-password?token=%s", token)
	
	fmt.Println("\n\n" + strings.Repeat("=", 60))
	fmt.Println("📧 VENTO RECOVERY EMAIL (MOCK)")
	fmt.Println("To:   ", req.Email)
	fmt.Println("Link: ", resetLink)
	fmt.Println(strings.Repeat("=", 60) + "\n\n")

	logger.L().Info("Password reset token generated", "email", req.Email, "link", resetLink)

	return nil
}

// ResetPassword validates the token and updates the user's password.
func (uc *ResetPasswordUsecases) ResetPassword(ctx context.Context, req dto.ResetPasswordRequest) error {
	email, err := uc.userRepo.GetEmailByResetToken(ctx, req.Token)
	if err != nil {
		return err
	}

	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}

	// Update password in domain entity to get hashing and validation
	if err := user.Password.Update(req.Password); err != nil {
		return err
	}

	if err := uc.userRepo.UpdatePassword(ctx, email, user.Password.Hash()); err != nil {
		return err
	}

	// Cleanup token
	return uc.userRepo.DeleteResetToken(ctx, req.Token)
}

func generateRandomToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
