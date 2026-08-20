package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

var (
	// accentColor is kvex's single brand accent — Azure blue, close to
	// Microsoft's #0078D4 — used everywhere something needs to draw the eye
	// (header, focused pane border, selected list item). accentMutedColor is
	// the same hue dialed down, for secondary emphasis (pane title bars,
	// selected item description text) so nothing competes with accentColor
	// for attention.
	accentColor      = lipgloss.Color("33")
	accentMutedColor = lipgloss.Color("24")

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentColor).
			Background(lipgloss.Color("235"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")).
			Background(lipgloss.Color("236"))

	errorStatusStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("124"))

	readOnlyBannerStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("2"))

	editBannerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("196"))

	versionBannerStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("214"))

	focusedBorderColor = accentColor
	blurredBorderColor = lipgloss.Color("240")
)

func paneStyle(focused bool) lipgloss.Style {
	s := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	if focused {
		return s.BorderForeground(focusedBorderColor)
	}
	return s.BorderForeground(blurredBorderColor)
}

// listStyles returns bubbles/list's default Styles with its pane title bar
// and filter cursor recolored from their library defaults (a purple title
// bar, a pink filter cursor) to kvex's accent colors, so every pane shares
// one consistent palette instead of three different ones.
func listStyles() list.Styles {
	s := list.DefaultStyles()
	s.Title = s.Title.Background(accentMutedColor)
	s.FilterCursor = s.FilterCursor.Foreground(accentColor)
	return s
}

// listItemStyles returns bubbles/list's default item styles with the
// selected-item highlight recolored from its library default (pink/magenta
// — #EE6FF8, #AD58B4) to kvex's accent colors, matching the header and
// focused-pane border instead of clashing with them.
func listItemStyles() list.DefaultItemStyles {
	s := list.NewDefaultItemStyles()
	s.SelectedTitle = s.SelectedTitle.
		BorderForeground(accentColor).
		Foreground(accentColor)
	s.SelectedDesc = s.SelectedTitle.Foreground(accentMutedColor)
	return s
}

// newListDelegate returns a list delegate pre-themed with listItemStyles,
// for use by every list.Model in the UI (vaults, secrets, versions).
func newListDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.Styles = listItemStyles()
	return d
}

// themeList applies listStyles to l and recolors its filter-mode text
// cursor to match. list.Model bakes its FilterInput's cursor style in at
// construction time from its own (unthemed) default Styles, so assigning
// l.Styles alone doesn't reach it — the cursor has to be recolored
// directly or it stays the library's default pink.
func themeList(l *list.Model) {
	l.Styles = listStyles()
	l.FilterInput.Cursor.Style = l.FilterInput.Cursor.Style.Foreground(accentColor)
}
