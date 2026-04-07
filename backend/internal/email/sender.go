package email

import (
	"fmt"
	"strconv"

	"gopkg.in/gomail.v2"
)

type Sender struct {
	dialer  *gomail.Dialer
	from    string
	baseURL string
}

func NewSender(host, portStr, user, pass, from, baseURL string) *Sender {
	port, _ := strconv.Atoi(portStr)
	dialer := gomail.NewDialer(host, port, user, pass)
	return &Sender{
		dialer:  dialer,
		from:    from,
		baseURL: baseURL,
	}
}

func (s *Sender) SendVerificationEmail(toEmail, token string) error {
	verifyURL := fmt.Sprintf("%s/verify-email?token=%s", s.baseURL, token)

	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "Confirm your MesaScore account")
	m.SetBody("text/plain", fmt.Sprintf("Click to verify your account: %s\n\nThis link expires in 24 hours.", verifyURL))
	m.AddAlternative("text/html", buildVerificationHTML(verifyURL))
	return s.dialer.DialAndSend(m)
}

func buildVerificationHTML(url string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family: sans-serif; padding: 20px;">
  <h2>Welcome to MesaScore!</h2>
  <p>Click the button below to verify your email address:</p>
  <a href="%s" style="display: inline-block; padding: 12px 24px; background-color: #4f46e5; color: white; text-decoration: none; border-radius: 6px; font-weight: bold;">Verify Email</a>
  <p style="margin-top: 20px; color: #666;">This link expires in 24 hours.</p>
  <p style="color: #999; font-size: 12px;">If you didn't create an account, you can ignore this email.</p>
</body>
</html>`, url)
}
