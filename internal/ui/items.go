package ui

import (
	"fmt"
	"time"

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
// tracks whether this version is selected for side-by-side comparison;
// current marks the newest version (versionList is always sorted
// newest-first, so this is exactly index 0).
type versionItem struct {
	version azure.Version
	marked  bool
	current bool
}

// Title renders the checkbox, short hash, and an age/current status all on
// one line — versions carry no second line (see Description) so a long
// list of them stays scannable, matching the design mockups' single-line
// version rows.
func (i versionItem) Title() string {
	v := i.version.Version
	if len(v) > 12 {
		v = v[:12]
	}
	box := "☐"
	if i.marked {
		box = "☑"
	}
	title := fmt.Sprintf("%s %s  %s", box, v, i.status())
	if !i.version.Enabled {
		title += " (disabled)"
	}
	return title
}

// status is "current" for the newest version, or a short relative age
// ("3d ago") for older ones — easier to scan at a glance than an absolute
// timestamp when a secret has many versions. The detail pane's meta footer
// still shows the precise absolute date for whichever version is loaded.
func (i versionItem) status() string {
	if i.current {
		return "current"
	}
	if i.version.Created == "" {
		return "unknown"
	}
	// azure.Version.Created is formatted from a time.Time via .Format
	// without first converting to UTC (see internal/azure), so the string
	// is local-zone wall-clock — it must be reparsed with the same
	// assumption. time.Parse alone would default to UTC and, off of
	// anything but a UTC machine, silently skew the "ago" by the zone
	// offset.
	t, err := time.ParseInLocation("2006-01-02 15:04:05", i.version.Created, time.Local)
	if err != nil {
		return i.version.Created
	}
	return relativeAge(t) + " ago"
}

func (i versionItem) Description() string { return "" }

func (i versionItem) FilterValue() string { return i.version.Version }

// relativeAge formats how long ago t was, coarsely (minutes/hours/days) —
// this only feeds a compact list row, not anything precision-sensitive.
func relativeAge(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}
