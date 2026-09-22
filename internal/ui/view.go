package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if !m.ready {
		return "starting kvex..."
	}

	header := headerStyle.Width(m.width).Render(" kvex — Azure Key Vault Explorer ")

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
	if m.editMode {
		detailContent = m.editArea.View()
		detailContentHeight = m.editArea.Height()
	} else {
		detailContent = m.detail.View()
		detailContentHeight = m.detail.Height
	}
	detailInner := lipgloss.JoinVertical(lipgloss.Left, m.bannerLine(), detailContent)
	detailBox := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).
		BorderForeground(m.detailBorderColor()).
		Width(m.detail.Width).Height(1 + detailContentHeight).
		Render(detailInner)

	versionsBox := paneStyle(m.focus == focusVersions).
		Width(m.versionList.Width()).Height(m.versionList.Height()).
		Render(m.versionList.View())

	panes := []string{vaultsBox, secretsBox, versionsBox, detailBox}

	body := lipgloss.JoinHorizontal(lipgloss.Top, panes...)

	return lipgloss.JoinVertical(lipgloss.Left, header, body, m.statusLine())
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
	style := statusStyle
	text := m.status
	if m.err != nil {
		style = errorStatusStyle
		text = m.err.Error()
	}
	return style.Width(m.width).Render(fmt.Sprintf(" %s   [%s]", text, m.keyHints()))
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
