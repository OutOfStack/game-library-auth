package resendapi

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/resend/resend-go/v2"
	"go.opentelemetry.io/otel"
)

//go:embed templates/*
var templateFS embed.FS

var (
	// ErrDailyQuotaExceeded is returned when the daily email quota is exceeded
	ErrDailyQuotaExceeded = errors.New("daily quota exceeded")

	tracer = otel.Tracer("resendapi")
)

// Client represents Resend client
type Client struct {
	client               *resend.Client
	baseURL              string
	fromEmail            string
	fromName             string
	contactEmail         string
	unsubscribeURL       string
	verificationHTMLTmpl *template.Template
	verificationTextTmpl *template.Template
}

// Config represents Resend client configuration
type Config struct {
	APIToken       string
	FromEmail      string
	ContactEmail   string
	BaseURL        string
	UnsubscribeURL string
	Timeout        time.Duration
}

// NewClient creates a new Resend client
func NewClient(cfg Config) (*Client, error) {
	httpClient := &http.Client{
		Timeout: cfg.Timeout,
	}
	client := resend.NewCustomClient(httpClient, cfg.APIToken)

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
		client:               client,
		baseURL:              cfg.BaseURL,
		fromEmail:            cfg.FromEmail,
		fromName:             "Game Library",
		contactEmail:         cfg.ContactEmail,
		unsubscribeURL:       cfg.UnsubscribeURL,
		verificationHTMLTmpl: htmlTmpl,
		verificationTextTmpl: textTmpl,
	}, nil
}

// SendEmailVerification sends email verification email with verification code and returns message id
func (c *Client) SendEmailVerification(ctx context.Context, req SendEmailVerificationRequest) (string, error) {
	ctx, span := tracer.Start(ctx, "sendEmailVerification")
	defer span.End()

	htmlContent, err := c.fillTemplate(c.verificationHTMLTmpl, req)
	if err != nil {
		return "", fmt.Errorf("fill HTML template: %w", err)
	}
	textContent, err := c.fillTemplate(c.verificationTextTmpl, req)
	if err != nil {
		return "", fmt.Errorf("fill text template: %w", err)
	}

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", c.fromName, c.fromEmail),
		To:      []string{req.Email},
		Subject: "Verify Your Email Address - Game Library",
		Html:    htmlContent,
		Text:    textContent,
	}

	sent, err := c.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		if isQuotaExceededError(err) {
			return "", ErrDailyQuotaExceeded
		}
		return "", fmt.Errorf("send email: %w", err)
	}

	return sent.Id, nil
}

// fillTemplate fills template placeholders
func (c *Client) fillTemplate(tmpl *template.Template, req SendEmailVerificationRequest) (string, error) {
	data := templateData{
		Email:             req.Email,
		Username:          req.Username,
		VerificationCode:  req.VerificationCode,
		UnsubscribeToken:  req.UnsubscribeToken,
		BaseURL:           c.baseURL,
		UnsubscribeURL:    c.unsubscribeURL,
		ContactEmail:      c.contactEmail,
		PrivacyPolicyURL:  c.baseURL + "/privacy-policy.html",
		TermsOfServiceURL: c.baseURL + "/terms-of-service.html",
		CurrentYear:       time.Now().Year(),
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}

func isQuotaExceededError(err error) bool {
	errStr := err.Error()
	return strings.Contains(errStr, "429") || strings.Contains(errStr, "Too Many Requests")
}
