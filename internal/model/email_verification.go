package model

import (
	"time"
)

const (
	// VerificationCodeTTL is the time a verification code is valid for
	VerificationCodeTTL = 24 * time.Hour
	// UnsubscribeTokenTTL is the time an unsubscribe token is valid for
	UnsubscribeTokenTTL = 7 * 24 * time.Hour

	// ResendVerificationCodeCooldown is the cooldown period between verification code resends
	ResendVerificationCodeCooldown = 120 * time.Second
)
