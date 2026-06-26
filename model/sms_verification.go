package model

import (
	"time"
)

// SMSVerification stores phone verification codes.
type SMSVerification struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Phone     string    `json:"phone" gorm:"type:varchar(20);index;not null"`
	Code      string    `json:"code" gorm:"type:varchar(10);not null"`
	Purpose   string    `json:"purpose" gorm:"type:varchar(32);index;not null"`
	Used      bool      `json:"used" gorm:"default:false"`
	ExpiresAt time.Time `json:"expires_at" gorm:"index"`
	CreatedAt time.Time `json:"created_at"`
}

const (
	SMSVerificationPurpose    = "sms_verification"
	SMSPasswordResetPurpose   = "sms_password_reset"
	SMSVerificationCodeLength = 4
)

// RecordSMSVerification saves a new SMS verification code.
func RecordSMSVerification(phone string, code string, purpose string, validMinutes int) error {
	verification := &SMSVerification{
		Phone:     phone,
		Code:      code,
		Purpose:   purpose,
		ExpiresAt: time.Now().Add(time.Duration(validMinutes) * time.Minute),
	}
	return DB.Create(verification).Error
}

// VerifySMSCode checks if the SMS code is valid and not expired.
func VerifySMSCode(phone string, code string, purpose string) bool {
	var verification SMSVerification
	result := DB.Where("phone = ? AND code = ? AND purpose = ? AND used = ? AND expires_at > ?",
		phone, code, purpose, false, time.Now()).First(&verification)
	if result.Error != nil {
		return false
	}
	// Mark as used
	DB.Model(&verification).Update("used", true)
	return true
}

// CleanupExpiredSMSVerifications removes expired codes.
func CleanupExpiredSMSVerifications() error {
	return DB.Where("expires_at < ?", time.Now().Add(-24*time.Hour)).Delete(&SMSVerification{}).Error
}
