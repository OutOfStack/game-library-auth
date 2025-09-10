package mailersend

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/mailersend/mailersend-go"
	"go.opentelemetry.io/otel"
)

//go:embed templates/*
var templateFS embed.FS

var (
	// ErrDailyQuotaExceeded is returned when the daily email quota is exceeded.
	ErrDailyQuotaExceeded = errors.New("daily quota exceeded")

	tracer = otel.Tracer("mailersend")
)

// Client represents MailerSend client
type Client struct {
	client               *mailersend.Mailersend
	fromEmail            string
	fromName             string
	timeout              time.Duration
	verificationHTMLTmpl *template.Template
	verificationTextTmpl *template.Template
}

// Config represents MailerSend client configuration
type Config struct {
	APIToken  string
	FromEmail string
	Timeout   time.Duration
}

// NewClient creates a new MailerSend client
func NewClient(cfg Config) (*Client, error) {
	ms := mailersend.NewMailersend(cfg.APIToken)

	htmlTemplateContent, err := templateFS.ReadFile("templates/email_verification.html")
	if err != nil {
		return nil, fmt.Errorf("load HTML template: %w", err)
	}

	textTemplateContent, err := templateFS.ReadFile("templates/email_verification.txt")
	if err != nil {
		return nil, fmt.Errorf("load text template: %w", err)
	}

	htmlTmpl, err := template.New("email").Parse(string(htmlTemplateContent))
	if err != nil {
		return nil, fmt.Errorf("parse HTML template: %w", err)
	}

	textTmpl, err := template.New("email").Parse(string(textTemplateContent))
	if err != nil {
		return nil, fmt.Errorf("parse text template: %w", err)
	}

	return &Client{
		client:               ms,
		fromEmail:            cfg.FromEmail,
		fromName:             "Game Library",
		timeout:              cfg.Timeout,
		verificationHTMLTmpl: htmlTmpl,
		verificationTextTmpl: textTmpl,
	}, nil
}

// SendEmailVerificationRequest represents email verification request
type SendEmailVerificationRequest struct {
	Email            string
	Username         string
	VerificationCode string
}

// SendEmailVerification sends email verification email with verification code and returns message id
func (c *Client) SendEmailVerification(ctx context.Context, req SendEmailVerificationRequest) (string, error) {
	ctx, span := tracer.Start(ctx, "sendEmailVerification")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	htmlContent, err := c.fillTemplate(c.verificationHTMLTmpl, req)
	if err != nil {
		return "", fmt.Errorf("fill HTML template: %w", err)
	}
	textContent, err := c.fillTemplate(c.verificationTextTmpl, req)
	if err != nil {
		return "", fmt.Errorf("fill text template: %w", err)
	}

	from := mailersend.From{
		Name:  c.fromName,
		Email: c.fromEmail,
	}

	recipients := []mailersend.Recipient{
		{
			Name:  req.Username,
			Email: req.Email,
		},
	}

	message := c.client.Email.NewMessage()
	message.SetFrom(from)
	message.SetRecipients(recipients)
	message.SetSubject("Verify Your Email Address - Game Library")
	message.SetHTML(htmlContent)
	message.SetText(textContent)

	res, err := c.client.Email.Send(ctx, message)
	if err != nil {
		return "", fmt.Errorf("send email: %w", err)
	}

	// Check for rate limiting (429 status)
	if res.StatusCode == http.StatusTooManyRequests {
		return "", ErrDailyQuotaExceeded
	}

	if res.StatusCode >= 400 {
		return "", fmt.Errorf("mailersend API error: status %d", res.StatusCode)
	}

	messageID := res.Header.Get("x-message-id")

	return messageID, nil
}

// fillTemplate fills template placeholders
func (c *Client) fillTemplate(tmpl *template.Template, req SendEmailVerificationRequest) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, req); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}
