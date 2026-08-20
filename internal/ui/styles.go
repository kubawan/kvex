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

	// selectedRowBg tints a selected list item's whole row, not just its
	// text, so the highlight reads as a filled row (closer to the design
	// mockups) instead of a bare colored border.
	selectedRowBg = lipgloss.Color("17")

	// paneLabelColor is the neutral grey used for pane title labels
	// ("VAULTS", "SECRETS · 5") — deliberately not accentColor, so the
	// blue accent stays reserved for focus/selection instead of coloring
	// everything.
	paneLabelColor = lipgloss.Color("245")

	headerBg = lipgloss.Color("235")

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentColor).
			Background(headerBg)

	// headerCrumbStyle renders the vault/secret breadcrumb trail next to
	// the bold "kvex" brand text in the header, once something is selected.
	headerCrumbStyle = lipgloss.NewStyle().
				Foreground(paneLabelColor).
				Background(headerBg)

	statusBg = lipgloss.Color("236")

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")).
			Background(statusBg)

	// statusKeyStyle bolds and accents just the key portion of a status-line
	// hint ("enter" in "enter: select"), so the keys you can press stand out
	// from what they do.
	statusKeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentColor).
			Background(statusBg)

	// metaStyle renders the small "created: ..." footer under a secret's
	// value in the detail pane.
	metaStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

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
// flattened from a colored pill (the library default, and an earlier purple
// one in kvex itself) to a plain small grey label — pane titles set their
// own uppercase text (e.g. "SECRETS · 5") to read as a label rather than a
// badge, matching the design mockups — and its filter cursor recolored to
// kvex's accent instead of the library's default pink.
func listStyles() list.Styles {
	s := list.DefaultStyles()
	s.Title = lipgloss.NewStyle().Foreground(paneLabelColor).Padding(0, 0, 1, 0)
	s.FilterCursor = s.FilterCursor.Foreground(accentColor)
	return s
}

// listItemStyles returns bubbles/list's default item styles with the
// selected-item highlight recolored from its library default (pink/magenta
// — #EE6FF8, #AD58B4) to kvex's accent colors, and a tinted row background
// added so the selection reads as a filled row instead of a bare border.
func listItemStyles() list.DefaultItemStyles {
	s := list.NewDefaultItemStyles()
	s.SelectedTitle = s.SelectedTitle.
		BorderForeground(accentColor).
		Foreground(accentColor).
		Background(selectedRowBg)
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
