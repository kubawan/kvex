package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if !m.ready {
		return "starting kvex..."
	}

	header := m.headerLine()

	// bubbles/list and bubbles/textarea don't pad short lines out to their
	// set width, so the pane style must force width/height explicitly —
	// otherwise lipgloss's border hugs the actual (shorter) content instead
	// of the allocated column width.
	vaultsBox := paneStyle(m.focus == focusVaults).
		Width(m.vaultList.Width()).Height(m.vaultList.Height()).
		Render(m.vaultList.View())
	secretsBox := paneStyle(m.focus == focusSecrets).
		Width(m.secretList.Width()).Height(m.secretList.Height()).
		Render(m.secretList.View())

	var detailContent string
	var detailContentHeight int
	var metaLine string
	if m.editMode {
		detailContent = m.editArea.View()
		detailContentHeight = m.editArea.Height()
	} else {
		detailContent = m.detail.View()
		detailContentHeight = m.detail.Height
		if m.comparingCount == 0 {
			if created, ok := m.currentVersionCreated(); ok {
				metaLine = "created: " + created
			}
		}
	}
	// The meta line's row is always reserved, even when empty, so the
	// detail pane's height doesn't jump around as you move between a plain
	// value, a historical version, and a comparison.
	detailInner := lipgloss.JoinVertical(lipgloss.Left, m.bannerLine(), detailContent, metaStyle.Render(" "+metaLine))
	detailBox := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).
		BorderForeground(m.detailBorderColor()).
		Width(m.detail.Width).Height(2 + detailContentHeight).
		Render(detailInner)

	versionsBox := paneStyle(m.focus == focusVersions).
		Width(m.versionList.Width()).Height(m.versionList.Height()).
		Render(m.versionList.View())

	panes := []string{vaultsBox, secretsBox, versionsBox, detailBox}

	body := lipgloss.JoinHorizontal(lipgloss.Top, panes...)

	return lipgloss.JoinVertical(lipgloss.Left, header, body, m.statusLine())
}

// headerLine renders the app header as a breadcrumb once the user has
// drilled into a vault/secret ("kvex › dev › db-password"), instead of a
// static title — so it's always clear which vault and secret is loaded
// without checking the secrets pane. Falls back to a plain subtitle before
// anything's selected, for orientation on first run.
func (m Model) headerLine() string {
	brand := headerStyle.Render(" kvex")
	if m.currentVaultName == "" {
		return lipgloss.NewStyle().Background(headerBg).Width(m.width).
			Render(brand + headerCrumbStyle.Render(" — Azure Key Vault Explorer "))
	}
	crumb := headerCrumbStyle.Render("  ›  " + m.currentVaultName)
	if m.currentSecretName != "" {
		crumb += headerCrumbStyle.Render("  ›  " + m.currentSecretName)
	}
	return lipgloss.NewStyle().Background(headerBg).Width(m.width).Render(brand + crumb + " ")
}

// currentVersionCreated returns the created timestamp of whatever version
// is currently loaded in the detail pane, for the small meta footer under
// the value. versionList is always sorted newest-first (see
// azure.SecretsClient.ListSecretVersions), so an empty m.currentVersion
// (meaning "latest") is simply its first item.
func (m Model) currentVersionCreated() (string, bool) {
	items := m.versionList.Items()
	if len(items) == 0 {
		return "", false
	}
	if m.currentVersion == "" {
		vi, ok := items[0].(versionItem)
		return vi.version.Created, ok && vi.version.Created != ""
	}
	for _, it := range items {
		vi, ok := it.(versionItem)
		if ok && vi.version.Version == m.currentVersion {
			return vi.version.Created, vi.version.Created != ""
		}
	}
	return "", false
}

// detailBorderColor gives the detail pane its own border color in edit
// mode (matching editBannerStyle) instead of just the banner — visible
// even out of the corner of your eye, since it's the one mode where a
// stray keystroke can change data. Every other state uses the normal
// focused/unfocused border like the other three panes.
func (m Model) detailBorderColor() lipgloss.Color {
	if m.editMode {
		return lipgloss.Color("196")
	}
	if m.focus == focusDetail {
		return focusedBorderColor
	}
	return blurredBorderColor
}

func (m Model) bannerLine() string {
	switch {
	case m.editMode:
		return editBannerStyle.Render(" EDIT MODE — ctrl+s save · esc cancel ")
	case m.comparingCount > 0:
		return versionBannerStyle.Render(" " + m.comparingSummary + " — read-only ")
	case m.currentVersion != "":
		short := m.currentVersion
		if len(short) > 12 {
			short = short[:12]
		}
		return versionBannerStyle.Render(fmt.Sprintf(" viewing version %s — press e to edit ", short))
	default:
		return readOnlyBannerStyle.Render(" READ-ONLY ")
	}
}

func (m Model) statusLine() string {
	if m.err != nil {
		return errorStatusStyle.Width(m.width).Render(fmt.Sprintf(" %s   [%s]", m.err.Error(), m.keyHints()))
	}

	// Bold+accent just the key portion of each "key: action" hint (e.g. the
	// "enter" in "enter: select"), so the keys you can press stand out from
	// what they do — matching the design mockups' bold-key/dim-label
	// convention. Segments without a "key: action" shape (there aren't any
	// today, but keyHints() is free-form) fall back to plain text.
	segments := strings.Split(m.keyHints(), " · ")
	rendered := make([]string, len(segments))
	for i, seg := range segments {
		key, action, ok := strings.Cut(seg, ": ")
		if !ok {
			rendered[i] = statusStyle.Render(seg)
			continue
		}
		rendered[i] = statusKeyStyle.Render(key) + statusStyle.Render(": "+action)
	}
	hints := strings.Join(rendered, statusStyle.Render(" · "))

	line := statusStyle.Render(" "+m.status+"   [") + hints + statusStyle.Render("]")
	return lipgloss.NewStyle().Background(statusBg).Width(m.width).Render(line)
}

// keyHints returns the plain-mode key hint text for the currently focused
// pane, so the status line only advertises keys that actually do something
// there (e.g. space/x only marks versions in the versions pane). Edit mode
// has its own contextual message via bannerLine(), so this is unused while
// m.editMode is true.
func (m Model) keyHints() string {
	copyHint := ""
	if m.currentValue != "" && m.comparingCount == 0 {
		copyHint = " · c: copy"
	}
	switch m.focus {
	case focusVaults:
		return "→: switch pane · ↑/↓: move · /: filter · enter: select" + copyHint + " · q: quit"
	case focusSecrets:
		hints := "←/→: switch pane · ↑/↓: move · /: filter · enter: select"
		if m.currentSecretName != "" {
			hints += " · e: edit"
		}
		return hints + copyHint + " · q: quit"
	case focusVersions:
		return "←/→: switch pane · ↑/↓: move · /: filter · enter: preview · space/x: mark · e: edit" + copyHint + " · q: quit"
	case focusDetail:
		return "←/→: switch pane · e: edit" + copyHint + " · q: quit"
	}
	return "q: quit"
}
