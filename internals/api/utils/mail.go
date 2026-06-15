package utils

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/luqmanshaban/gomont/internals/config"
	"gopkg.in/mail.v2"
)

func SendEmail(cfg *config.Config, to string, code int) error {
	from := cfg.EMAIL_USER
	password := cfg.EMAIL_PASS // Use an App Password for Gmail
	smtpHost := cfg.EMAIL_HOST
	smtpPort := cfg.EMAIL_PORT
	// receivers := []string{to}

	msg := mail.NewMessage()
	msg.SetHeader("From", from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", "OTP-verification code")

	// Set email body
	msg.SetBody("text/plain", fmt.Sprintf("%d",code))

	// Set up the SMTP dialer
	dialer := mail.NewDialer(smtpHost, smtpPort, from, password)
	dialer.StartTLSPolicy = mail.MandatoryStartTLS

	// Send the email
	if err := dialer.DialAndSend(msg); err != nil {
		slog.Error("Error sending email", "error", err)
		return err
	}
	slog.Info("Email sent successfully!", "email", to)
	return nil
}

func SendNotificationEmail(cfg *config.Config, to, url, err string, time time.Time) error {
	from := cfg.EMAIL_USER
	password := cfg.EMAIL_PASS // Use an App Password for Gmail
	smtpHost := cfg.EMAIL_HOST
	smtpPort := cfg.EMAIL_PORT
	// receivers := []string{to}

	msg := mail.NewMessage()
	msg.SetHeader("From", from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", "Your application has issues")

	// Set email body
	msg.SetBody("text/plain", fmt.Sprintf(`
	URL: 	%s
	Error: %v 
	Time: %v
	`,url, err, time))

	// Set up the SMTP dialer
	dialer := mail.NewDialer(smtpHost, smtpPort, from, password)
	dialer.StartTLSPolicy = mail.MandatoryStartTLS

	// Send the email
	if err := dialer.DialAndSend(msg); err != nil {
		slog.Error("Error sending Notification email", "error", err)
		return err
	}
	slog.Info("Notification Email sent successfully!", "email", to)
	return nil
}