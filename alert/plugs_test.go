package alert

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestSlackWebhookUsesInjectedClient(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		got = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	plug := NewSlack(srv.URL, srv.Client())
	ev := Event{SKU: "sku-1", Engine: "claude", Reason: "price-delta", Summary: "|Δprice|=2"}
	if err := plug.Notify(ev); err != nil {
		t.Fatal(err)
	}
	var payload map[string]string
	if err := json.Unmarshal([]byte(got), &payload); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(payload["text"], "sku-1") || !strings.Contains(payload["text"], "price-delta") {
		t.Fatalf("payload = %s", got)
	}
	if len(plug.Sent()) != 0 {
		t.Fatal("webhook path should not record on the stub")
	}
}

func TestSlackWebhookRequiresClient(t *testing.T) {
	plug := NewSlack("https://example.invalid/hook", nil)
	err := plug.Notify(Event{SKU: "sku-1", Reason: "stock-flip"})
	if err == nil || !strings.Contains(err.Error(), "no implicit network") {
		t.Fatalf("err = %v", err)
	}
}

func TestSlackWebhookStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusBadRequest)
	}))
	defer srv.Close()
	plug := NewSlack(srv.URL, srv.Client())
	if err := plug.Notify(Event{SKU: "sku-1", Reason: "price-delta"}); err == nil {
		t.Fatal("expected status error")
	}
}

func TestEmailSendFunc(t *testing.T) {
	var mu sync.Mutex
	type mail struct{ from, subject, body string }
	var sent mail
	plug := NewEmail(EmailOptions{
		From: "otm@example.com",
		To:   []string{"ops@example.com"},
		Send: func(from string, to []string, subject, body string) error {
			mu.Lock()
			defer mu.Unlock()
			sent = mail{from: from, subject: subject, body: body}
			if len(to) != 1 || to[0] != "ops@example.com" {
				t.Errorf("to = %v", to)
			}
			return nil
		},
	})
	if err := plug.Notify(Event{SKU: "sku-9", Reason: "stock-flip", Summary: "flipped"}); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	if sent.from != "otm@example.com" || !strings.Contains(sent.subject, "sku-9") {
		t.Fatalf("sent = %+v", sent)
	}
	if !strings.Contains(sent.body, "stock-flip") {
		t.Fatalf("body = %s", sent.body)
	}
	if len(plug.Sent()) != 0 {
		t.Fatal("send path should not record on the stub")
	}
}

func TestSMTPSenderRequiresHostAtSendTime(t *testing.T) {
	fn := SMTPSender(EmailOptions{})
	if err := fn("a@b.c", []string{"d@e.f"}, "s", "b"); err == nil {
		t.Fatal("expected host error without dialing")
	}
}

func TestFormatRFC822(t *testing.T) {
	msg := string(formatRFC822("from@x", []string{"a@x", "b@x"}, "hello", "body"))
	for _, want := range []string{"From: from@x", "To: a@x, b@x", "Subject: hello", "body"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("missing %q in %q", want, msg)
		}
	}
}

func TestNewEmailWithHostWiresSMTPSender(t *testing.T) {
	plug := NewEmail(EmailOptions{Host: "smtp.example.invalid", From: "a@b.c", To: []string{"d@e.f"}})
	if plug.send == nil {
		t.Fatal("expected SMTP sender when host is set")
	}
}
