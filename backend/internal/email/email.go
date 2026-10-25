// Package email sends transactional emails over SMTP.
//
// If SMTP is not configured (no Host), Send() logs the full content to the
// process log instead of trying to deliver. This makes the admin panel fully
// usable in dev and on a fresh VPS before SMTP is set up — the magic link is
// visible in `harady.log` or `journalctl -u harady`.
package email

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/smtp"
	"strings"
	"time"
)

type Sender struct {
	Host    string
	Port    int
	User    string
	Pass    string
	From    string
	enabled bool
}

func NewSender(host string, port int, user, pass, from string) *Sender {
	return &Sender{
		Host:    host,
		Port:    port,
		User:    user,
		Pass:    pass,
		From:    from,
		enabled: host != "",
	}
}

// Send delivers a plain-text email. If SMTP is disabled, it logs the
// full message body under the given subject tag so the admin can retrieve
// the magic link from the log.
func (s *Sender) Send(to, subject, body string) error {
	if !s.enabled {
		slog.Info("email (SMTP disabled, logging content)",
			"to", to, "subject", subject, "body", body)
		return nil
	}

	msg := buildMessage(s.From, to, subject, body)

	addr := net.JoinHostPort(s.Host, fmt.Sprint(s.Port))
	auth := smtp.PlainAuth("", s.User, s.Pass, s.Host)

	// Try STARTTLS on ports that support it (587, 25). Ports like 465 use
	// implicit TLS which needs a different dial — treat 465 specially.
	if s.Port == 465 {
		return sendImplicitTLS(addr, s.Host, auth, s.From, to, msg)
	}
	return sendStartTLS(addr, s.Host, auth, s.From, to, msg)
}

func buildMessage(from, to, subject, body string) []byte {
	// Strip CRLF from header fields to prevent header injection.
	sanitize := func(s string) string {
		return strings.NewReplacer("\r", "", "\n", "").Replace(s)
	}
	var b strings.Builder
	b.WriteString("From: " + sanitize(from) + "\r\n")
	b.WriteString("To: " + sanitize(to) + "\r\n")
	b.WriteString("Subject: " + sanitize(subject) + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	b.WriteString("Date: " + time.Now().Format(time.RFC1123Z) + "\r\n")
	b.WriteString("\r\n")
	b.WriteString(body)
	return []byte(b.String())
}

func sendStartTLS(addr, host string, auth smtp.Auth, from, to string, msg []byte) error {
	c, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer c.Close()

	if ok, _ := c.Extension("STARTTLS"); ok {
		if err := c.StartTLS(&tls.Config{ServerName: host}); err != nil {
			return fmt.Errorf("starttls: %w", err)
		}
	}
	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err := c.Auth(auth); err != nil {
				return fmt.Errorf("auth: %w", err)
			}
		}
	}
	return deliver(c, from, to, msg)
}

func sendImplicitTLS(addr, host string, auth smtp.Auth, from, to string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
	if err != nil {
		return fmt.Errorf("tls dial: %w", err)
	}
	c, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("smtp client: %w", err)
	}
	defer c.Close()

	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err := c.Auth(auth); err != nil {
				return fmt.Errorf("auth: %w", err)
			}
		}
	}
	return deliver(c, from, to, msg)
}

func deliver(c *smtp.Client, from, to string, msg []byte) error {
	if err := c.Mail(from); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}
	if err := c.Rcpt(to); err != nil {
		return fmt.Errorf("rcpt to: %w", err)
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close data: %w", err)
	}
	return c.Quit()
}
