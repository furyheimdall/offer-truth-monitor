package alert

import (
	"fmt"
	"net"
	"net/smtp"
)

// smtpSend is the production SMTP dial. Tests never call it.
func smtpSend(addr, user, pass, from string, to []string, msg []byte) error {
	var auth smtp.Auth
	if user != "" {
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			return fmt.Errorf("alert: email: smtp addr: %w", err)
		}
		auth = smtp.PlainAuth("", user, pass, host)
	}
	if err := smtp.SendMail(addr, auth, from, to, msg); err != nil {
		return fmt.Errorf("alert: email: smtp: %w", err)
	}
	return nil
}
