package email

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const sendTimeout = 30 * time.Second

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string // e.g. Pingwise <alerts@example.com>
	AppUrl   string // the web app's address, for links in emails
}

var cfg *Config

// Init sets up the SMTP sender. With an empty host, emails are turned off and Enabled returns false.
func Init(c Config) error {
	if c.Host == "" {
		cfg = nil
		return nil
	}
	if _, err := mail.ParseAddress(c.From); err != nil {
		return errors.Wrap(err, "ALERT_FROM_EMAIL is not a valid address")
	}
	cfg = &c
	return nil
}

func Enabled() bool {
	return cfg != nil
}

// AppUrl returns the web app's address without a trailing slash
func AppUrl() string {
	if cfg == nil {
		return ""
	}
	return strings.TrimRight(cfg.AppUrl, "/")
}

// Send sends a plain-text email to one recipient. It upgrades the connection with STARTTLS
// (as Gmail on port 587 requires) and refuses to send the password over an unencrypted connection.
func Send(to, subject, body string) error {
	if cfg == nil {
		return errors.New("email is not configured")
	}
	from, err := mail.ParseAddress(cfg.From)
	if err != nil {
		return errors.Wrap(err, "invalid from address")
	}
	toAddr, err := mail.ParseAddress(to)
	if err != nil {
		return errors.Wrap(err, "invalid recipient")
	}

	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return errors.Wrap(err, "failed to connect to the mail server")
	}
	_ = conn.SetDeadline(time.Now().Add(sendTimeout))
	client, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		_ = conn.Close()
		return errors.Wrap(err, "failed to start the SMTP session")
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: cfg.Host, MinVersion: tls.VersionTLS12}); err != nil {
			return errors.Wrap(err, "STARTTLS failed")
		}
	}
	if cfg.Username != "" {
		// PlainAuth only sends the password over TLS, or to localhost
		if err := client.Auth(smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)); err != nil {
			return errors.Wrap(err, "SMTP login failed")
		}
	}
	if err := client.Mail(from.Address); err != nil {
		return errors.Wrap(err, "MAIL FROM rejected")
	}
	if err := client.Rcpt(toAddr.Address); err != nil {
		return errors.Wrap(err, "recipient rejected")
	}
	w, err := client.Data()
	if err != nil {
		return errors.Wrap(err, "DATA rejected")
	}
	if _, err := w.Write(buildMessage(from, toAddr, subject, body)); err != nil {
		return errors.Wrap(err, "failed to write the message")
	}
	if err := w.Close(); err != nil {
		return errors.Wrap(err, "the mail server did not accept the message")
	}
	return client.Quit()
}

func buildMessage(from, to *mail.Address, subject, body string) []byte {
	// header values come from user input (service names), so line breaks are removed
	subject = strings.NewReplacer("\r", " ", "\n", " ").Replace(subject)
	domain := from.Address[strings.LastIndex(from.Address, "@")+1:]

	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", from.String())
	fmt.Fprintf(&b, "To: %s\r\n", to.String())
	fmt.Fprintf(&b, "Subject: %s\r\n", mime.QEncoding.Encode("utf-8", subject))
	fmt.Fprintf(&b, "Date: %s\r\n", time.Now().Format(time.RFC1123Z))
	fmt.Fprintf(&b, "Message-ID: <%s@%s>\r\n", randomId(), domain)
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	b.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	b.WriteString("\r\n")
	b.WriteString(strings.ReplaceAll(strings.ReplaceAll(body, "\r\n", "\n"), "\n", "\r\n"))
	return []byte(b.String())
}

func randomId() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
