package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"kvex/internal/azure"
	"kvex/internal/config"
)

// newTestModel builds a ready Model backed by two independent mock vaults
// (no real Azure credential needed), sized so list/detail panes have room
// to hold items.
func newTestModel() Model {
	cfg := &config.Config{Vaults: []config.Vault{
		{Name: "mock", URI: azure.MockScheme + "local"},
		{Name: "mock2", URI: azure.MockScheme + "local2"},
	}}
	m := New(cfg, nil)
	m.width, m.height = 120, 40
	m.layout()
	return m
}

// keyMsg builds the tea.KeyMsg that msg.String() would report as s, for the
// small set of keys kvex actually handles.
func keyMsg(s string) tea.KeyMsg {
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case " ":
		return tea.KeyMsg{Type: tea.KeySpace}
	case "ctrl+s":
		return tea.KeyMsg{Type: tea.KeyCtrlS}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

// collectMsgs runs cmd (and, recursively, every sub-command of any
// tea.BatchMsg it produces) and returns the flattened list of resulting
// messages, in the order bubbletea would eventually deliver them.
func collectMsgs(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		var out []tea.Msg
		for _, c := range batch {
			out = append(out, collectMsgs(c)...)
		}
		return out
	}
	return []tea.Msg{msg}
}

// applyAll feeds each message in msgs through Model.Update in order,
// threading the resulting model forward.
func applyAll(m Model, msgs []tea.Msg) Model {
	for _, msg := range msgs {
		tm, _ := m.Update(msg)
		m = tm.(Model)
	}
	return m
}
