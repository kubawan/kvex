package ui

import (
	"fmt"

	"kvex/internal/azure"
	"kvex/internal/config"
)

// vaultItem adapts a config.Vault for display in a bubbles/list.
type vaultItem struct {
	vault config.Vault
}

func (i vaultItem) Title() string { return i.vault.Name }

func (i vaultItem) Description() string {
	if i.vault.Subscription == "" {
		return i.vault.URI
	}
	return fmt.Sprintf("%s (%s)", i.vault.URI, i.vault.Subscription)
}

func (i vaultItem) FilterValue() string { return i.vault.Name }

// secretItem adapts a secret name for display in a bubbles/list.
type secretItem struct {
	name string
}

func (i secretItem) Title() string       { return i.name }
func (i secretItem) Description() string { return "" }
func (i secretItem) FilterValue() string { return i.name }

// versionItem adapts azure.Version for display in a bubbles/list. marked
// tracks whether this version is selected for side-by-side comparison.
type versionItem struct {
	version azure.Version
	marked  bool
}

func (i versionItem) Title() string {
	v := i.version.Version
	if len(v) > 12 {
		v = v[:12]
	}
	box := "☐"
	if i.marked {
		box = "☑"
	}
	title := box + " " + v
	if !i.version.Enabled {
		title += " (disabled)"
	}
	return title
}

func (i versionItem) Description() string {
	if i.version.Created == "" {
		return "created: unknown"
	}
	return "created: " + i.version.Created
}

func (i versionItem) FilterValue() string { return i.version.Version }
