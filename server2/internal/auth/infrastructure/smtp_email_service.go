package infrastructure

import (
	"fmt"
	"net/smtp"
	"os"
)

// SMTPEmailService implements EmailSender using SMTP
type SMTPEmailService struct {
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
}

// NewSMTPEmailService creates a new SMTP email service
func NewSMTPEmailService() *SMTPEmailService {
	return &SMTPEmailService{
		smtpHost:     os.Getenv("SMTP_HOST"),
		smtpPort:     os.Getenv("SMTP_PORT"),
		smtpUsername: os.Getenv("SMTP_USERNAME"),
		smtpPassword: os.Getenv("SMTP_PASSWORD"),
	}
}

// SendVerificationEmail sends an email verification link
func (s *SMTPEmailService) SendVerificationEmail(to, username, verificationToken string) error {
	frontendURL := os.Getenv("FE_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	verificationLink := fmt.Sprintf("%s/verify?token=%s", frontendURL, verificationToken)

	subject := "Verify Your Email - AI Clipper"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4F46E5; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .button { display: inline-block; padding: 12px 30px; background-color: #ffffff; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Welcome to AI Clipper!</h1>
        </div>
        <div class="content">
            <p>Hi %s,</p>
            <p>Thank you for registering with AI Clipper. Please verify your email address to activate your account.</p>
            <p style="text-align: center;">
                <a href="%s" class="button">Verify Email</a>
            </p>
            <p>Or copy and paste this link into your browser:</p>
            <p style="word-break: break-all; color: #ffffff;">%s</p>
            <p>This verification link will expire in 24 hours.</p>
            <p>If you didn't create an account with AI Clipper, please ignore this email.</p>
        </div>
        <div class="footer">
            <p>&copy; 2025 AI Clipper. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, username, verificationLink, verificationLink)

	return s.sendEmail(to, subject, body)
}

// SendPasswordResetEmail sends a password reset link
func (s *SMTPEmailService) SendPasswordResetEmail(to, username, resetToken string) error {
	frontendURL := os.Getenv("FE_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, resetToken)

	subject := "Reset Your Password - AI Clipper"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4F46E5; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .button { display: inline-block; padding: 12px 30px; background-color: #ffffff; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
        .warning { background-color: #FEF3C7; border-left: 4px solid #F59E0B; padding: 10px; margin: 15px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Password Reset Request</h1>
        </div>
        <div class="content">
            <p>Hi %s,</p>
            <p>We received a request to reset your password for your AI Clipper account.</p>
            <p style="text-align: center;">
                <a href="%s" class="button">Reset Password</a>
            </p>
            <p>Or copy and paste this link into your browser:</p>
            <p style="word-break: break-all; color: #ffffff;">%s</p>
            <div class="warning">
                <strong>Security Notice:</strong> This link will expire in 1 hour for security reasons.
            </div>
            <p>If you didn't request a password reset, please ignore this email or contact support if you have concerns.</p>
        </div>
        <div class="footer">
            <p>&copy; 2025 AI Clipper. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, username, resetLink, resetLink)

	return s.sendEmail(to, subject, body)
}

// sendEmail sends an email using SMTP
func (s *SMTPEmailService) sendEmail(to, subject, body string) error {
	// Get from email or use username as default
	fromEmail := s.smtpUsername

	// Build email message
	message := fmt.Sprintf("From: %s\r\n", fromEmail)
	message += fmt.Sprintf("To: %s\r\n", to)
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	message += "MIME-Version: 1.0\r\n"
	message += "Content-Type: text/html; charset=UTF-8\r\n"
	message += "\r\n"
	message += body

	// Setup authentication
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)

	// Send email
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)
	err := smtp.SendMail(addr, auth, fromEmail, []string{to}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
