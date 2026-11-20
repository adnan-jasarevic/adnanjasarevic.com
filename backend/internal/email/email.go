package email

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/adnan-jasarevic/adnanjasarevic.com/backend/internal/store"
	"github.com/resend/resend-go/v2"
)

type Resend struct {
	client *resend.Client
	from   string
	to     string
}

func NewResend(apiKey, from, to string) *Resend {
	return &Resend{client: resend.NewClient(apiKey), from: from, to: to}
}

func (r *Resend) NotifyContact(ctx context.Context, id int64, m store.NewContactMessage) error {
	name := strings.Join(strings.Fields(m.Name), " ")
	req := &resend.SendEmailRequest{
		From:    r.from,
		To:      []string{r.to},
		ReplyTo: m.Email,
		Subject: "New contact message from " + name,
		Text:    fmt.Sprintf("From: %s <%s>\n\n%s\n", name, m.Email, m.Message),
	}
	opts := &resend.SendEmailOptions{IdempotencyKey: fmt.Sprintf("contact-message/%d", id)}

	sent, err := r.client.Emails.SendWithOptions(ctx, req, opts)
	if err != nil {
		return fmt.Errorf("send contact notification: %w", err)
	}
	log.Printf("contact notification sent: message %d, email %s", id, sent.Id)
	return nil
}
