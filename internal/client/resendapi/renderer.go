package resendapi

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	texttemplate "text/template"
	"time"

	"github.com/OutOfStack/game-library-auth/internal/notification"
)

//go:embed templates/*
var templateFS embed.FS

// Config contains email template settings
type Config struct {
	FromEmail      string
	ContactEmail   string
	BaseURL        string
	UnsubscribeURL string
}

// Renderer prepares email content without network access
type Renderer struct {
	cfg  Config
	html *template.Template
	text *texttemplate.Template
}

// NewRenderer parses the embedded verification templates
func NewRenderer(cfg Config) (*Renderer, error) {
	html, err := template.ParseFS(templateFS, "templates/email_verification.html")
	if err != nil {
		return nil, fmt.Errorf("parse verification HTML: %w", err)
	}
	text, err := texttemplate.ParseFS(templateFS, "templates/email_verification.txt")
	if err != nil {
		return nil, fmt.Errorf("parse verification text: %w", err)
	}
	return &Renderer{cfg: cfg, html: html, text: text}, nil
}

// PrepareVerification freezes all template and configuration values for later delivery
func (r *Renderer) PrepareVerification(req SendEmailVerificationRequest) (notification.Email, error) {
	data := templateData{
		Email: req.Email, Username: req.Username, VerificationCode: req.VerificationCode, UnsubscribeToken: req.UnsubscribeToken,
		BaseURL: r.cfg.BaseURL, UnsubscribeURL: r.cfg.UnsubscribeURL, ContactEmail: r.cfg.ContactEmail,
		PrivacyPolicyURL: r.cfg.BaseURL + "/privacy-policy.html", TermsOfServiceURL: r.cfg.BaseURL + "/terms-of-service.html",
		CurrentYear: time.Now().UTC().Year(),
	}
	var html, text bytes.Buffer
	if err := r.html.Execute(&html, data); err != nil {
		return notification.Email{}, err
	}
	if err := r.text.Execute(&text, data); err != nil {
		return notification.Email{}, err
	}
	return notification.Email{From: fmt.Sprintf("Game Library <%s>", r.cfg.FromEmail), To: []string{req.Email},
		Subject: "Verify Your Email Address - Game Library", HTML: html.String(), Text: text.String()}, nil
}
