package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Options configures the day-1 Slack + email plugs.
// Empty credentials keep each plug on its in-memory stub.
type Options struct {
	SlackWebhook string
	HTTPClient   *http.Client // required when SlackWebhook is set (no implicit DefaultClient)
	Email        EmailOptions
}

// EmailOptions is the email plug config. A nil Send func with an empty
// Host stays on the stub. SMTPSender may be supplied as Send when a host
// is later configured; tests inject a fake Send.
type EmailOptions struct {
	From string
	To   []string
	Host string
	Port int
	User string
	Pass string
	Send SendFunc
}

// SendFunc delivers one email. Tests inject a recorder; production may
// use SMTPSender. A nil SendFunc means "no credentials yet".
type SendFunc func(from string, to []string, subject, body string) error

// Day1Plugs returns the Slack and email notifier plugs.
func Day1Plugs(opts Options) []Notifier {
	return []Notifier{
		NewSlack(opts.SlackWebhook, opts.HTTPClient),
		NewEmail(opts.Email),
	}
}

// SlackPlug posts incoming-webhook JSON when a webhook URL and client are
// set. Otherwise it records in memory (credential stub).
type SlackPlug struct {
	webhook string
	client  *http.Client
	stub    *Recorder
}

// NewSlack returns the Slack plug. An empty webhook is the stub path.
func NewSlack(webhookURL string, client *http.Client) *SlackPlug {
	return &SlackPlug{
		webhook: strings.TrimSpace(webhookURL),
		client:  client,
		stub:    SlackStub(),
	}
}

// Channel implements Notifier.
func (s *SlackPlug) Channel() Channel { return Slack }

// Sent returns stub-recorded events (empty when the webhook path is used).
func (s *SlackPlug) Sent() []Event {
	if s.stub == nil {
		return nil
	}
	return s.stub.Sent()
}

// Notify implements Notifier. The webhook path requires an explicit client
// so tests never accidentally dial the public internet.
func (s *SlackPlug) Notify(e Event) error {
	if err := validateEvent(e); err != nil {
		return err
	}
	if s.webhook == "" {
		return s.stub.Notify(e)
	}
	if s.client == nil {
		return fmt.Errorf("alert: slack: webhook configured but client is nil (no implicit network)")
	}
	payload, err := json.Marshal(map[string]string{"text": FormatText(e)})
	if err != nil {
		return fmt.Errorf("alert: slack: encode: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, s.webhook, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("alert: slack: request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("alert: slack: post: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 512))
		return fmt.Errorf("alert: slack: status %d: %s", res.StatusCode, bytes.TrimSpace(body))
	}
	return nil
}

// EmailPlug sends via Send when configured; otherwise records in memory.
type EmailPlug struct {
	from string
	to   []string
	send SendFunc
	stub *Recorder
}

// NewEmail returns the email plug. Missing Send (and no SMTP host) is the stub.
func NewEmail(opts EmailOptions) *EmailPlug {
	send := opts.Send
	if send == nil && strings.TrimSpace(opts.Host) != "" {
		send = SMTPSender(opts)
	}
	return &EmailPlug{
		from: opts.From,
		to:   append([]string(nil), opts.To...),
		send: send,
		stub: EmailStub(),
	}
}

// Channel implements Notifier.
func (e *EmailPlug) Channel() Channel { return Email }

// Sent returns stub-recorded events (empty when Send is used).
func (e *EmailPlug) Sent() []Event {
	if e.stub == nil {
		return nil
	}
	return e.stub.Sent()
}

// Notify implements Notifier.
func (e *EmailPlug) Notify(ev Event) error {
	if err := validateEvent(ev); err != nil {
		return err
	}
	if e.send == nil {
		return e.stub.Notify(ev)
	}
	from := e.from
	if from == "" {
		from = "otm@localhost"
	}
	to := e.to
	if len(to) == 0 {
		to = []string{"alerts@localhost"}
	}
	subject := fmt.Sprintf("[otm] %s %s", ev.SKU, ev.Reason)
	return e.send(from, to, subject, FormatText(ev))
}

// SMTPSender returns a SendFunc that uses net/smtp when credentials exist.
// Tests must not call the returned func; inject a fake Send instead.
func SMTPSender(opts EmailOptions) SendFunc {
	host := strings.TrimSpace(opts.Host)
	port := opts.Port
	if port == 0 {
		port = 587
	}
	addr := fmt.Sprintf("%s:%d", host, port)
	from := opts.From
	return func(fromOverride string, to []string, subject, body string) error {
		if fromOverride != "" {
			from = fromOverride
		}
		if host == "" {
			return fmt.Errorf("alert: email: smtp host required")
		}
		if from == "" || len(to) == 0 {
			return fmt.Errorf("alert: email: from and to required")
		}
		msg := formatRFC822(from, to, subject, body)
		return smtpSend(addr, opts.User, opts.Pass, from, to, msg)
	}
}

func formatRFC822(from string, to []string, subject, body string) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "From: %s\r\n", from)
	fmt.Fprintf(&b, "To: %s\r\n", strings.Join(to, ", "))
	fmt.Fprintf(&b, "Subject: %s\r\n", subject)
	fmt.Fprintf(&b, "MIME-Version: 1.0\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n")
	b.WriteString(body)
	if !strings.HasSuffix(body, "\n") {
		b.WriteByte('\n')
	}
	return b.Bytes()
}
