package i18n

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
)

//go:embed email_templates
var emailTemplatesFS embed.FS

// EmailVerificationData holds data for the email verification template.
type EmailVerificationData struct {
	SystemName   string
	Code         string
	ValidMinutes int
	Year         int
}

// EmailNotificationData holds data for the notification template.
type EmailNotificationData struct {
	SystemName string
	Title      string
	Content    string
	Year       int
}

// PasswordResetData holds data for the password reset template.
type PasswordResetData struct {
	SystemName   string
	Link         string
	ValidMinutes int
	Year         int
}

// languageToTemplateDir maps i18n language codes to template directory names.
var languageToTemplateDir = map[string]string{
	"ru":    "ru",
	"zh-CN": "zh",
	"zh-TW": "zh",
	"en":    "en",
}

// RenderEmailTemplate renders an email template for the given language.
// templateName: base name without .html extension (e.g. "email_verification", "notification")
// lang: i18n language code (e.g. "ru", "en", "zh-CN")
// data: template data struct
func RenderEmailTemplate(templateName string, c *gin.Context, data any) (string, error) {
	lang := GetLangFromContext(c)
	langDir := "en"
	if dir, ok := languageToTemplateDir[lang]; ok {
		langDir = dir
	}

	templatePath := fmt.Sprintf("email_templates/%s/%s.html", langDir, templateName)

	tmpl, err := template.New(templateName + ".html").ParseFS(emailTemplatesFS, templatePath)
	if err != nil {
		// Fallback to English if the template is missing for the detected language
		if langDir != "en" {
			fallbackPath := fmt.Sprintf("email_templates/en/%s.html", templateName)
			tmpl, err = template.New(templateName + ".html").ParseFS(emailTemplatesFS, fallbackPath)
			if err != nil {
				return "", fmt.Errorf("failed to parse email template %s: %w", templateName, err)
			}
		} else {
			return "", fmt.Errorf("failed to parse email template %s: %w", templateName, err)
		}
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute email template %s: %w", templateName, err)
	}

	return buf.String(), nil
}

// NewEmailVerificationData builds template data for email verification.
func NewEmailVerificationData(systemName string, code string, validMinutes int) EmailVerificationData {
	return EmailVerificationData{
		SystemName:   systemName,
		Code:         code,
		ValidMinutes: validMinutes,
		Year:         time.Now().Year(),
	}
}

// NewEmailNotificationData builds template data for notifications.
func NewEmailNotificationData(systemName string, title string, content string) EmailNotificationData {
	return EmailNotificationData{
		SystemName: systemName,
		Title:      title,
		Content:    content,
		Year:       time.Now().Year(),
	}
}

// SendTemplateEmail renders a template and sends it via SMTP.
func SendTemplateEmail(templateName string, c *gin.Context, subject string, receiver string, data any) error {
	content, err := RenderEmailTemplate(templateName, c, data)
	if err != nil {
		return err
	}
	return common.SendEmail(subject, receiver, content)
}
