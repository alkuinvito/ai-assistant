package mailer

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
)

type Mailer interface {
	Send(ctx context.Context, to, subject, body string) error
}

type SmtpConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type SmtpMail struct {
	client *smtp.Client
	config *SmtpConfig
	logger *logrus.Logger
}

func NewSmtpMail(logger *logrus.Logger) (Mailer, func()) {
	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		logger.WithError(err).Fatal("invalid SMTP_PORT")
	}

	config := &SmtpConfig{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     port,
		Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     os.Getenv("SMTP_FROM"),
	}

	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         config.Host,
	}

	addr := net.JoinHostPort(config.Host, strconv.Itoa(config.Port))
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		logger.WithError(err).WithField("addr", addr).Fatal("failed to connect to SMTP server")
	}

	// 2. Create the SMTP client
	client, err := smtp.NewClient(conn, config.Host)
	if err != nil {
		logger.WithError(err).Fatal("failed to create SMTP client")
	}

	if err := client.StartTLS(tlsconfig); err != nil {
		logger.WithError(err).Fatal("failed to upgrade connection via STARTTLS")
	}

	if config.Username != "" {
		auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)
		if err := client.Auth(auth); err != nil {
			logger.WithError(err).Fatal("failed to authenticate with SMTP server")
		}
	}

	cleanup := func() {
		client.Quit()
	}

	return &SmtpMail{
		client: client,
		config: config,
		logger: logger,
	}, cleanup
}

func (g *SmtpMail) Send(ctx context.Context, to, subject, body string) error {
	if err := g.client.Mail(g.config.From); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	if err := g.client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	w, err := g.client.Data()
	if err != nil {
		return fmt.Errorf("failed to open data writer: %w", err)
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n%s",
		g.config.From, to, subject, body)

	if _, err := w.Write([]byte(msg)); err != nil {
		return fmt.Errorf("failed to write message body: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return nil
}
