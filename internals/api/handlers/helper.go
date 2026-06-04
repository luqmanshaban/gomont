package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/luqmanshaban/gomont/internals/config"
	"gopkg.in/mail.v2"
)

func WriteJson(w http.ResponseWriter, status int, data any) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func SendEmail(cfg *config.Config, to string, code int) error {
	from := cfg.EMAIL_USER
	password := cfg.EMAIL_PASS // Use an App Password for Gmail
	smtpHost := cfg.EMAIL_HOST
	smtpPort := cfg.EMAIL_PORT
	println(smtpPort)
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
