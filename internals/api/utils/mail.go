package utils

import (
	"bytes"
	"fmt"
	htmlTemplate "html/template"
	"log/slog"
	textTemplate "text/template"
	"time"

	"github.com/luqmanshaban/gomont/internals/config"
	"gopkg.in/mail.v2"
)

func renderHTML(name, tmplStr string, data any) (string, error) {
	tmpl, err := htmlTemplate.New(name).Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("parse html template %s: %w", name, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute html template %s: %w", name, err)
	}
	return buf.String(), nil
}

func renderText(name, tmplStr string, data any) (string, error) {
	tmpl, err := textTemplate.New(name).Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("parse text template %s: %w", name, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute text template %s: %w", name, err)
	}
	return buf.String(), nil
}

func dashboardURL(cfg *config.Config) string {
	return fmt.Sprintf("%s/dashboard", cfg.APP_URL)
}

// send builds and dispatches a multipart (plaintext + HTML) message
func send(cfg *config.Config, to []string, subject, plainBody, htmlBody string) error {
	msg := mail.NewMessage()
	msg.SetHeader("From", cfg.EMAIL_USER)
	
	// FIXED: Use the variadic spread operator (...) to expand the slice for the header safely
	msg.SetHeader("To", to...) 
	msg.SetHeader("Subject", subject)

	msg.SetBody("text/plain", plainBody)
	msg.AddAlternative("text/html", htmlBody)

	dialer := mail.NewDialer(cfg.EMAIL_HOST, cfg.EMAIL_PORT, cfg.EMAIL_USER, cfg.EMAIL_PASS)
	dialer.StartTLSPolicy = mail.MandatoryStartTLS

	if err := dialer.DialAndSend(msg); err != nil {
		slog.Error("error sending email", "to", to, "subject", subject, "error", err)
		return err
	}
	slog.Info("email sent successfully", "to", to, "subject", subject)
	return nil
}

// SendEmail handles single recipient OTP delivery
func SendEmail(cfg *config.Config, to string, code int) error {
	data := struct{ Code int }{Code: code}

	plainBody, err := renderText("otp_text", otpEmailText, data)
	if err != nil {
		return err
	}
	htmlBody, err := renderHTML("otp_html", otpEmailHTML, data)
	if err != nil {
		return err
	}

	// FIXED: Wrap the single string recipient into an inline string slice literal
	return send(cfg, []string{to}, "Your Gomont verification code", plainBody, htmlBody)
}

// SendNotificationEmail alerts multiple destination addresses that a monitor went down.
func SendNotificationEmail(cfg *config.Config, to []string, url, errMsg string, checkedAt time.Time) error {
	data := struct {
		URL          string
		Err          string
		Time         string
		DashboardURL string
	}{
		URL:          url,
		Err:          errMsg,
		Time:         checkedAt.Format("January 2, 2006, 3:04 PM MST"),
		DashboardURL: dashboardURL(cfg),
	}

	plainBody, err := renderText("notification_text", notificationEmailText, data)
	if err != nil {
		return err
	}
	htmlBody, err := renderHTML("notification_html", notificationEmailHTML, data)
	if err != nil {
		return err
	}

	return send(cfg, to, fmt.Sprintf("Gomont alert: %s is down", url), plainBody, htmlBody)
}