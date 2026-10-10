// Package mailer delivers transactional email (verification, password reset).
//
// The default Mailer writes the message to the log so development and tests
// work end to end without a mail server. Set SMTP_HOST to switch to real
// delivery via net/smtp.
package mailer

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

type Message struct {
	To      string
	Subject string
	Body    string
}

type Mailer interface {
	Send(Message) error
}

// New returns an SMTP-backed mailer when host is set, and a logging mailer
// otherwise. from is the envelope/From address.
func New(host string, port int, username, password, from string) Mailer {
	if strings.TrimSpace(host) == "" {
		log.Print("mailer: SMTP not configured, emails will be logged only")
		return LogMailer{}
	}
	return &smtpMailer{host: host, port: port, username: username, password: password, from: from}
}

// LogMailer writes messages to the process log. Useful for local development
// and as a safe fallback when no SMTP server is configured.
type LogMailer struct{}

func (LogMailer) Send(m Message) error {
	log.Printf("mailer: to=%s subject=%q\n%s", m.To, m.Subject, m.Body)
	return nil
}

type smtpMailer struct {
	host     string
	port     int
	username string
	password string
	from     string
}

func (m *smtpMailer) Send(msg Message) error {
	addr := fmt.Sprintf("%s:%d", m.host, m.port)

	var auth smtp.Auth
	if m.username != "" {
		auth = smtp.PlainAuth("", m.username, m.password, m.host)
	}

	body := strings.Join([]string{
		"From: " + m.from,
		"To: " + msg.To,
		"Subject: " + msg.Subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=\"utf-8\"",
		"",
		msg.Body,
	}, "\r\n")

	return smtp.SendMail(addr, auth, m.from, []string{msg.To}, []byte(body))
}
