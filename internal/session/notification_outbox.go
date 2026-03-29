package session

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// OutboxNotification is the JSON structure written to the outbox directory.
type OutboxNotification struct {
	ID        string `json:"id"`
	Profile   string `json:"profile"`
	Platform  string `json:"platform"`
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
}

// OutboxDir returns the outbox directory for a named conductor.
func OutboxDir(conductorName string) (string, error) {
	dir, err := ConductorNameDir(conductorName)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "outbox"), nil
}

// WriteNotification writes a notification file to the conductor outbox.
// The file is written atomically (temp file + rename).
func WriteNotification(conductorName, platform, chatID, text string) error {
	if conductorName == "" {
		return fmt.Errorf("conductor name is required")
	}
	if platform == "" {
		return fmt.Errorf("platform is required")
	}
	if chatID == "" {
		return fmt.Errorf("chat_id is required")
	}
	if text == "" {
		return fmt.Errorf("text is required")
	}

	switch platform {
	case "telegram", "slack", "discord":
	default:
		return fmt.Errorf("unsupported platform %q (must be telegram, slack, or discord)", platform)
	}

	dir, err := OutboxDir(conductorName)
	if err != nil {
		return fmt.Errorf("outbox dir: %w", err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create outbox dir: %w", err)
	}

	randomBytes := make([]byte, 4)
	if _, err := rand.Read(randomBytes); err != nil {
		return fmt.Errorf("generate random id: %w", err)
	}
	randomHex := hex.EncodeToString(randomBytes)

	now := time.Now()
	id := fmt.Sprintf("%d-%s", now.UnixMilli(), randomHex)
	filename := id + ".json"

	notification := OutboxNotification{
		ID:        id,
		Profile:   conductorName,
		Platform:  platform,
		ChatID:    chatID,
		Text:      text,
		CreatedAt: now.UTC().Format(time.RFC3339),
	}

	data, err := json.MarshalIndent(notification, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal notification: %w", err)
	}

	tmpFile := filepath.Join(dir, filename+".tmp")
	finalFile := filepath.Join(dir, filename)

	if err := os.WriteFile(tmpFile, data, 0o644); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := os.Rename(tmpFile, finalFile); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("rename to final: %w", err)
	}

	return nil
}
