package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/asheshgoplani/agent-deck/internal/session"
)

// handleNotify sends a proactive notification to the bridge outbox.
func handleNotify(profile string, args []string) {
	fs := flag.NewFlagSet("notify", flag.ExitOnError)

	platform := fs.String("platform", "", "Target platform: telegram, slack, or discord")
	chatID := fs.String("chat-id", "", "Platform-specific chat/channel ID")
	tg := fs.String("tg", "", "Shorthand: --platform telegram --chat-id <value>")
	slack := fs.String("slack", "", "Shorthand: --platform slack --chat-id <value>")
	discord := fs.String("discord", "", "Shorthand: --platform discord --chat-id <value>")

	fs.Usage = func() {
		fmt.Println("Usage: agent-deck [-p profile] notify [flags] <message>")
		fmt.Println()
		fmt.Println("Send a proactive notification to the bridge (Telegram/Slack/Discord).")
		fmt.Println()
		fmt.Println("Flags:")
		fs.PrintDefaults()
		fmt.Println()
		fmt.Println("Shorthand flags (alternative to --platform + --chat-id):")
		fmt.Println("  --tg <chat_id>       Telegram chat/group ID")
		fmt.Println("  --slack <channel_id>  Slack channel ID")
		fmt.Println("  --discord <channel_id> Discord channel ID")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  agent-deck -p default notify --tg -100123456 \"Delegating to session 'auth-flow'\"")
		fmt.Println("  agent-deck notify --platform slack --chat-id C67890 \"Session completed\"")
	}

	if err := fs.Parse(normalizeArgs(fs, args)); err != nil {
		os.Exit(1)
	}

	// Resolve platform + chatID from shorthand flags
	resolvedPlatform := *platform
	resolvedChatID := *chatID

	shorthands := 0
	if *tg != "" {
		shorthands++
		resolvedPlatform = "telegram"
		resolvedChatID = *tg
	}
	if *slack != "" {
		shorthands++
		resolvedPlatform = "slack"
		resolvedChatID = *slack
	}
	if *discord != "" {
		shorthands++
		resolvedPlatform = "discord"
		resolvedChatID = *discord
	}

	if shorthands > 1 {
		fmt.Fprintln(os.Stderr, "Error: specify only one of --tg, --slack, --discord")
		os.Exit(1)
	}

	if shorthands > 0 && (*platform != "" || *chatID != "") {
		fmt.Fprintln(os.Stderr, "Error: cannot combine shorthand flags (--tg/--slack/--discord) with --platform/--chat-id")
		os.Exit(1)
	}

	if resolvedPlatform == "" || resolvedChatID == "" {
		fmt.Fprintln(os.Stderr, "Error: platform and chat-id are required. Use --tg/--slack/--discord or --platform + --chat-id")
		fs.Usage()
		os.Exit(1)
	}

	// Message is the remaining positional args joined
	text := strings.TrimSpace(strings.Join(fs.Args(), " "))
	if text == "" {
		fmt.Fprintln(os.Stderr, "Error: notification message text is required")
		os.Exit(1)
	}

	// Resolve conductor name from profile
	conductorName := profile
	if conductorName == "" {
		conductorName = "default"
	}

	if err := session.WriteNotification(conductorName, resolvedPlatform, resolvedChatID, text); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
