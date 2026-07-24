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
	detailBox := paneStyle(m.focus == focusDetail).
		Width(m.detail.Width).Height(1 + detailContentHeight).
		Render(detailInner)

	panes := []string{vaultsBox, secretsBox}
	if m.showVersions {
		versionsBox := paneStyle(m.focus == focusVersions).
			Width(m.versionList.Width()).Height(m.versionList.Height()).
			Render(m.versionList.View())
		panes = append(panes, versionsBox)
	}
	panes = append(panes, detailBox)

	body := lipgloss.JoinHorizontal(lipgloss.Top, panes...)

	return lipgloss.JoinVertical(lipgloss.Left, header, body, m.statusLine())
}

func (m Model) bannerLine() string {
	switch {
	case m.editMode:
		return editBannerStyle.Render(" EDIT MODE — ctrl+s save · esc cancel ")
	case m.currentVersion != "":
		short := m.currentVersion
		if len(short) > 12 {
			short = short[:12]
		}
		return versionBannerStyle.Render(fmt.Sprintf(" viewing version %s — read-only ", short))
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
	help := "tab: switch pane · enter: select · e: edit · v: versions · q: quit"
	return style.Width(m.width).Render(fmt.Sprintf(" %s   [%s]", text, help))
}
