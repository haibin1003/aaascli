package service

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"time"
)

// OTPService provides TOTP (Time-based One-Time Password) functionality
// Compatible with Google Authenticator, Microsoft Authenticator, etc.
type OTPService struct{}

// NewOTPService creates a new OTP service instance
func NewOTPService() *OTPService {
	return &OTPService{}
}

// GenerateSecret generates a new random TOTP secret
// Returns base32 encoded secret string (without padding)
func (s *OTPService) GenerateSecret() (string, error) {
	// Generate 20 random bytes (160 bits) for SHA1
	secret := make([]byte, 20)
	if _, err := rand.Read(secret); err != nil {
		return "", fmt.Errorf("failed to generate random secret: %w", err)
	}

	// Encode to base32 without padding
	return base32.StdEncoding.EncodeToString(secret), nil
}

// GenerateQRCodeURL generates a QR code URL for authenticator apps
// Format: otpauth://totp/{account}?secret={secret}&issuer={issuer}
func (s *OTPService) GenerateQRCodeURL(secret, account, issuer string) string {
	// Clean up inputs
	account = strings.TrimSpace(account)
	issuer = strings.TrimSpace(issuer)

	// Build the otpauth URL
	// Label 只使用账号，避免显示乱码
	// Format: otpauth://totp/ACCOUNT?secret=SECRET&issuer=ISSUER
	return fmt.Sprintf(
		"otpauth://totp/%s?secret=%s&issuer=%s&algorithm=SHA1&digits=6&period=30",
		account,
		secret,
		issuer,
	)
}

// GenerateCode generates a TOTP code for the given secret at the specified time
func (s *OTPService) GenerateCode(secret string, t time.Time) (string, error) {
	// Decode base32 secret
	key, err := base32.StdEncoding.DecodeString(strings.ToUpper(secret))
	if err != nil {
		return "", fmt.Errorf("invalid secret: %w", err)
	}

	// Calculate time counter (30-second windows)
	counter := uint64(math.Floor(float64(t.Unix()) / 30))

	// Create HMAC-SHA1
	mac := hmac.New(sha1.New, key)
	binary.Write(mac, binary.BigEndian, counter)
	hash := mac.Sum(nil)

	// Dynamic truncation
	offset := hash[len(hash)-1] & 0x0F
	code := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7FFFFFFF

	// Modulo to get 6 digits
	code = code % 1000000

	// Format with leading zeros
	return fmt.Sprintf("%06d", code), nil
}

// VerifyCode verifies a TOTP code against the secret
// window specifies how many 30-second windows to check before/after current time (default 1)
func (s *OTPService) VerifyCode(secret, code string, window int) (bool, error) {
	if window < 0 {
		window = 1
	}
	if window > 3 {
		window = 3 // Max 3 windows for security
	}

	now := time.Now()

	// Check current window and adjacent windows
	for i := -window; i <= window; i++ {
		t := now.Add(time.Duration(i) * 30 * time.Second)
		expectedCode, err := s.GenerateCode(secret, t)
		if err != nil {
			return false, err
		}

		// Use constant-time comparison to prevent timing attacks
		if constantTimeEquals(expectedCode, code) {
			return true, nil
		}
	}

	return false, nil
}

// GenerateBackupCodes generates a set of backup codes for account recovery
// Returns 10 codes, each 8 characters long
func (s *OTPService) GenerateBackupCodes() ([]string, error) {
	codes := make([]string, 10)
	for i := 0; i < 10; i++ {
		// Generate 4 random bytes
		bytes := make([]byte, 4)
		if _, err := rand.Read(bytes); err != nil {
			return nil, fmt.Errorf("failed to generate backup code: %w", err)
		}
		// Format as 8-character hex string
		codes[i] = fmt.Sprintf("%08x", binary.BigEndian.Uint32(bytes))
	}
	return codes, nil
}

// constantTimeEquals compares two strings in constant time to prevent timing attacks
func constantTimeEquals(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}

// GetRemainingSeconds returns the number of seconds remaining in the current TOTP window
func (s *OTPService) GetRemainingSeconds() int {
	now := time.Now()
	elapsed := now.Unix() % 30
	return int(30 - elapsed)
}

// FormatSecretForDisplay formats the secret for human-readable display (groups of 4)
func (s *OTPService) FormatSecretForDisplay(secret string) string {
	secret = strings.ToUpper(secret)
	var result strings.Builder
	for i, c := range secret {
		if i > 0 && i%4 == 0 {
			result.WriteByte(' ')
		}
		result.WriteRune(c)
	}
	return result.String()
}
