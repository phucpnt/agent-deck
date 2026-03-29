package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteNotification(t *testing.T) {
	// Override home dir so we write to a temp location
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	err := WriteNotification("test-conductor", "telegram", "-100123456", "hello from test")
	if err != nil {
		t.Fatalf("WriteNotification failed: %v", err)
	}

	outboxDir := filepath.Join(tmp, ".agent-deck", "conductor", "test-conductor", "outbox")
	entries, err := os.ReadDir(outboxDir)
	if err != nil {
		t.Fatalf("read outbox dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 file in outbox, got %d", len(entries))
	}

	filename := entries[0].Name()
	if !strings.HasSuffix(filename, ".json") {
		t.Errorf("expected .json suffix, got %s", filename)
	}
	// No .tmp files should remain
	if strings.HasSuffix(filename, ".tmp") {
		t.Errorf("temp file should not remain: %s", filename)
	}

	data, err := os.ReadFile(filepath.Join(outboxDir, filename))
	if err != nil {
		t.Fatalf("read notification file: %v", err)
	}

	var notif OutboxNotification
	if err := json.Unmarshal(data, &notif); err != nil {
		t.Fatalf("unmarshal notification: %v", err)
	}

	if notif.Profile != "test-conductor" {
		t.Errorf("profile = %q, want %q", notif.Profile, "test-conductor")
	}
	if notif.Platform != "telegram" {
		t.Errorf("platform = %q, want %q", notif.Platform, "telegram")
	}
	if notif.ChatID != "-100123456" {
		t.Errorf("chat_id = %q, want %q", notif.ChatID, "-100123456")
	}
	if notif.Text != "hello from test" {
		t.Errorf("text = %q, want %q", notif.Text, "hello from test")
	}
	if notif.ID == "" {
		t.Error("id should not be empty")
	}
	if notif.CreatedAt == "" {
		t.Error("created_at should not be empty")
	}
}

func TestWriteNotification_Validation(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	tests := []struct {
		name     string
		cond     string
		platform string
		chatID   string
		text     string
		wantErr  string
	}{
		{"empty conductor", "", "telegram", "123", "msg", "conductor name is required"},
		{"empty platform", "cond", "", "123", "msg", "platform is required"},
		{"empty chat_id", "cond", "telegram", "", "msg", "chat_id is required"},
		{"empty text", "cond", "telegram", "123", "", "text is required"},
		{"bad platform", "cond", "whatsapp", "123", "msg", "unsupported platform"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WriteNotification(tt.cond, tt.platform, tt.chatID, tt.text)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want containing %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestWriteNotification_AutoCreatesDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	outboxDir := filepath.Join(tmp, ".agent-deck", "conductor", "auto-create", "outbox")
	// Verify dir doesn't exist yet
	if _, err := os.Stat(outboxDir); !os.IsNotExist(err) {
		t.Fatal("outbox dir should not exist before test")
	}

	err := WriteNotification("auto-create", "slack", "C12345", "test auto-create")
	if err != nil {
		t.Fatalf("WriteNotification failed: %v", err)
	}

	if _, err := os.Stat(outboxDir); err != nil {
		t.Fatalf("outbox dir should exist after WriteNotification: %v", err)
	}
}
