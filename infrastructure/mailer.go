package infrastructure

import (
	"fmt"
	"net/smtp"
	"os"
)

type Mailer interface {
	SendVerificationEmail(toEmail, token string) error
	SendResetPasswordEmail(toEmail, otp string) error
	SendApplicationAcceptedEmail(toEmail, fullname, unitName, mentorName, attendanceDate string) error
	SendApplicationRejectedEmail(toEmail, fullname, adminNotes string) error
	SendLoginOTPEmail(toEmail, otp string) error
}

type gmailMailer struct {
	sender   string
	username string
	password string
	host     string
	port     string
	baseURL  string
}

func NewGmailMailer(sender, password string) Mailer {
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		host = "smtp.gmail.com"
	}

	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "587"
	}

	baseURL := os.Getenv("APP_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:9090"
	}

	username := os.Getenv("SMTP_USERNAME")
    if username == "" {
        username = sender
    }

	return &gmailMailer{
		sender:   sender,
		username: username,
		password: password,
		host:     host,
		port:     port,
		baseURL:  baseURL,
	}
}

func (m *gmailMailer) SendVerificationEmail(toEmail, token string) error {
	subject := "Verifikasi Akun"
	verifyLink := fmt.Sprintf("%s/api/v1/verify?token=%s", m.baseURL, token)

	htmlBody := GetVerificationEmailTemplate(verifyLink)

	return m.send(toEmail, subject, htmlBody)
}

func (m *gmailMailer) SendResetPasswordEmail(toEmail, otp string) error {
	subject := "Password Reset Code"
	htmlBody := GetResetPasswordTemplate(otp)

	return m.send(toEmail, subject, htmlBody)
}

func (m *gmailMailer) SendApplicationAcceptedEmail(toEmail, fullname, unitName, mentorName, attendanceDate string) error {
	subject := "Pendaftaran PKL Anda Diterima"
	htmlBody := GetApplicationAcceptedTemplate(fullname, unitName, mentorName, attendanceDate)
	return m.send(toEmail, subject, htmlBody)
}

func (m *gmailMailer) SendApplicationRejectedEmail(toEmail, fullname, adminNotes string) error {
	subject := "Informasi Status Pendaftaran PKL"
	htmlBody := GetApplicationRejectedTemplate(fullname, adminNotes)
	return m.send(toEmail, subject, htmlBody)
}

func (m *gmailMailer) send(to, subject, htmlBody string) error {
	if err := m.validate(); err != nil {
		return err
	}

	msg := buildMessage(m.sender, to, subject, htmlBody)
	auth := smtp.PlainAuth("", m.username, m.password, m.host)

	if err := smtp.SendMail(m.host+":"+m.port, auth, m.sender, []string{to}, msg); err != nil {
		return fmt.Errorf("gagal mengirim email ke %s: %w", to, err)
	}
	return nil
}

func (m *gmailMailer) validate() error {
	if m.sender == "" || m.password == "" || m.host == "" || m.port == "" {
		return fmt.Errorf("konfigurasi SMTP tidak lengkap")
	}
	return nil
}

func buildMessage(from, to, subject, htmlBody string) []byte {
	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n%s",
		from, to, subject, htmlBody,
	)
	return []byte(msg)
}

func (m *gmailMailer) SendLoginOTPEmail(toEmail, otp string) error {
	return m.send(toEmail, "Kode OTP Login", GetLoginOTPTemplate(otp))
}