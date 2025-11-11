package application

// EmailSender defines the interface for sending emails
type EmailSender interface {
	// SendVerificationEmail sends an email verification link to the user
	SendVerificationEmail(to, username, verificationToken string) error

	// SendPasswordResetEmail sends a password reset link to the user
	SendPasswordResetEmail(to, username, resetToken string) error
}
