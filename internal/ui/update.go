package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"kvex/internal/azure"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if m.width > 0 && m.height > 0 {
			m.ready = true
		}
		m.layout()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case secretNamesLoadedMsg:
		return m.onSecretNamesLoaded(msg)

	case secretValueLoadedMsg:
		return m.onSecretValueLoaded(msg)

	case secretVersionsLoadedMsg:
		return m.onSecretVersionsLoaded(msg)

	case secretSavedMsg:
		return m.onSecretSaved(msg)

	case secretVersionValuesLoadedMsg:
		return m.onSecretVersionValuesLoaded(msg)
	}

	// Non-key, non-app messages (e.g. cursor blink ticks) go to whichever
	// sub-component currently owns input focus.
	var cmd tea.Cmd
	switch {
	case m.focus == focusDetail && m.editMode:
		m.editArea, cmd = m.editArea.Update(msg)
	case m.focus == focusVaults:
		m.vaultList, cmd = m.vaultList.Update(msg)
	case m.focus == focusSecrets:
		m.secretList, cmd = m.secretList.Update(msg)
	case m.focus == focusVersions:
		m.versionList, cmd = m.versionList.Update(msg)
	}
	return m, cmd
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	switch m.focus {
	case focusVaults:
		return m.handleVaultsKey(msg)
	case focusSecrets:
		return m.handleSecretsKey(msg)
	case focusDetail:
		if m.editMode {
			return m.handleEditKey(msg)
		}
		return m.handleDetailKey(msg)
	case focusVersions:
		return m.handleVersionsKey(msg)
	}
	return m, nil
}

func (m Model) handleVaultsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.vaultList.SettingFilter() {
		var cmd tea.Cmd
		m.vaultList, cmd = m.vaultList.Update(msg)
		return m, cmd
	}
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "right":
		m.focus = focusSecrets
		return m, nil
	case "left":
		m.focus = focusDetail
		return m, nil
	case "enter":
		item, ok := m.vaultList.SelectedItem().(vaultItem)
		if !ok {
			return m, nil
		}
		return m.selectVault(item.vault.Name)
	}
	var cmd tea.Cmd
	m.vaultList, cmd = m.vaultList.Update(msg)
	return m, cmd
}

func (m Model) handleSecretsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.secretList.SettingFilter() {
		var cmd tea.Cmd
		m.secretList, cmd = m.secretList.Update(msg)
		return m, cmd
	}
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "right":
		if m.showVersions {
			m.focus = focusVersions
		} else {
			m.focus = focusDetail
		}
		return m, nil
	case "left":
		m.focus = focusVaults
		return m, nil
	case "enter":
		item, ok := m.secretList.SelectedItem().(secretItem)
		if !ok {
			return m, nil
		}
		return m.selectSecret(item.name)
	case "e":
		return m.enterEditMode()
	case "v":
		return m.toggleVersions()
	}
	var cmd tea.Cmd
	m.secretList, cmd = m.secretList.Update(msg)
	return m, cmd
}

func (m Model) handleDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "right":
		m.focus = focusVaults
		return m, nil
	case "left":
		if m.showVersions {
			m.focus = focusVersions
		} else {
			m.focus = focusSecrets
		}
		return m, nil
	case "e":
		return m.enterEditMode()
	case "v":
		return m.toggleVersions()
	}
	var cmd tea.Cmd
	m.detail, cmd = m.detail.Update(msg)
	return m, cmd
}

// enterEditMode is shared by handleDetailKey and handleSecretsKey so 'e'
// works whether or not the user has explicitly moved focus into the detail
// pane — you shouldn't have to navigate panes just to edit what you're
// already previewing.
func (m Model) enterEditMode() (Model, tea.Cmd) {
	if m.currentSecretName == "" {
		return m, nil
	}
	if m.currentVersion != "" || m.comparingCount > 0 {
		m.status = "cannot edit here — select the latest version first"
		return m, nil
	}
	m.editMode = true
	m.editArea.SetValue(m.currentValue)
	m.editArea.Focus()
	m.focus = focusDetail
	m.status = "EDIT MODE — ctrl+s to save, esc to cancel"
	return m, nil
}

// toggleVersions is shared by handleDetailKey and handleSecretsKey, same
// reasoning as enterEditMode.
func (m Model) toggleVersions() (Model, tea.Cmd) {
	if m.currentSecretName == "" {
		return m, nil
	}
	if m.showVersions {
		m.showVersions = false
		m.focus = focusDetail
		m.layout()
		return m, nil
	}
	m.showVersions = true
	m.focus = focusVersions
	m.layout()
	client, err := m.clientFor(m.currentVaultName)
	if err != nil {
		m.err = err
		m.status = err.Error()
		return m, nil
	}
	m.loading = true
	m.status = "loading versions..."
	return m, fetchSecretVersionsCmd(client, m.currentVaultName, m.currentSecretName)
}

func (m Model) handleEditKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.editMode = false
		m.editArea.Blur()
		m.status = "edit cancelled"
		return m, nil
	case "ctrl+s":
		client, err := m.clientFor(m.currentVaultName)
		if err != nil {
			m.err = err
			m.status = err.Error()
			return m, nil
		}
		value := m.editArea.Value()
		m.loading = true
		m.status = "saving..."
		return m, saveSecretCmd(client, m.currentVaultName, m.currentSecretName, value)
	}
	var cmd tea.Cmd
	m.editArea, cmd = m.editArea.Update(msg)
	return m, cmd
}

func (m Model) handleVersionsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.versionList.SettingFilter() {
		var cmd tea.Cmd
		m.versionList, cmd = m.versionList.Update(msg)
		return m, cmd
	}
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "esc", "v":
		m.showVersions = false
		m.focus = focusDetail
		m.layout()
		return m, nil
	case "right":
		m.focus = focusDetail
		return m, nil
	case "left":
		m.focus = focusSecrets
		return m, nil
	case " ", "x":
		idx := m.versionList.Index()
		item, ok := m.versionList.SelectedItem().(versionItem)
		if !ok {
			return m, nil
		}
		item.marked = !item.marked
		if item.marked {
			m.markedVersions[item.version.Version] = true
		} else {
			delete(m.markedVersions, item.version.Version)
		}
		cmd := m.versionList.SetItem(idx, item)
		if len(m.markedVersions) > 0 {
			m.status = fmt.Sprintf("%d version(s) marked — enter to compare", len(m.markedVersions))
		}
		return m, cmd
	case "enter":
		client, err := m.clientFor(m.currentVaultName)
		if err != nil {
			m.err = err
			m.status = err.Error()
			return m, nil
		}

		if len(m.markedVersions) >= 2 {
			var marked []azure.Version
			for _, it := range m.versionList.Items() {
				vi, ok := it.(versionItem)
				if ok && m.markedVersions[vi.version.Version] {
					marked = append(marked, vi.version)
				}
			}
			m.loading = true
			m.focus = focusDetail
			m.status = fmt.Sprintf("loading %d versions to compare...", len(marked))
			return m, fetchSecretVersionValuesCmd(client, m.currentVaultName, m.currentSecretName, marked)
		}

		item, ok := m.versionList.SelectedItem().(versionItem)
		if !ok {
			return m, nil
		}
		m.currentVersion = item.version.Version
		m.comparingCount = 0
		m.comparingSummary = ""
		m.loading = true
		m.focus = focusDetail
		short := item.version.Version
		if len(short) > 12 {
			short = short[:12]
		}
		m.status = "loading version " + short + "..."
		return m, fetchSecretValueCmd(client, m.currentVaultName, m.currentSecretName, item.version.Version)
	}
	var cmd tea.Cmd
	m.versionList, cmd = m.versionList.Update(msg)
	return m, cmd
}

func (m *Model) setSecretItems(names []string) {
	items := make([]list.Item, len(names))
	for i, n := range names {
		items[i] = secretItem{name: n}
	}
	m.secretList.SetItems(items)
}

func (m Model) selectVault(name string) (Model, tea.Cmd) {
	m.currentVaultName = name
	m.currentSecretName = ""
	m.currentVersion = ""
	m.currentValue = ""
	m.editMode = false
	m.showVersions = false
	m.markedVersions = make(map[string]bool)
	m.comparingCount = 0
	m.comparingSummary = ""
	m.focus = focusSecrets

	if names, ok := m.secretNamesCache[name]; ok {
		m.setSecretItems(names)
		m.status = fmt.Sprintf("%s: %d secrets (cached)", name, len(names))
		return m, nil
	}

	client, err := m.clientFor(name)
	if err != nil {
		m.err = err
		m.status = err.Error()
		return m, nil
	}
	m.setSecretItems(nil)
	m.loading = true
	m.status = "loading secrets for " + name + "..."
	return m, fetchSecretNamesCmd(client, name)
}

func (m Model) selectSecret(name string) (Model, tea.Cmd) {
	client, err := m.clientFor(m.currentVaultName)
	if err != nil {
		m.err = err
		m.status = err.Error()
		return m, nil
	}
	m.currentSecretName = name
	m.currentVersion = ""
	m.editMode = false
	m.showVersions = false
	m.markedVersions = make(map[string]bool)
	m.comparingCount = 0
	m.comparingSummary = ""
	// Deliberately don't move focus to the detail pane here: staying on the
	// secrets list lets you preview values with just up/down + enter,
	// without needing to navigate back after every single secret.
	m.layout()
	m.loading = true
	m.detail.SetContent("loading...")
	m.status = fmt.Sprintf("loading %s...", name)
	return m, fetchSecretValueCmd(client, m.currentVaultName, name, "")
}

func (m Model) onSecretNamesLoaded(msg secretNamesLoadedMsg) (Model, tea.Cmd) {
	m.loading = false
	if msg.err != nil {
		m.err = msg.err
		m.status = "error listing secrets: " + msg.err.Error()
		return m, nil
	}
	m.secretNamesCache[msg.vault] = msg.names
	if m.currentVaultName == msg.vault {
		m.setSecretItems(msg.names)
		m.status = fmt.Sprintf("%s: %d secrets", msg.vault, len(msg.names))
	}
	return m, nil
}

func (m Model) onSecretValueLoaded(msg secretValueLoadedMsg) (Model, tea.Cmd) {
	m.loading = false
	if msg.err != nil {
		m.err = msg.err
		m.status = "error fetching secret: " + msg.err.Error()
		m.detail.SetContent(msg.err.Error())
		return m, nil
	}
	if m.currentVaultName == msg.vault && m.currentSecretName == msg.name {
		m.currentValue = msg.value
		m.comparingCount = 0
		m.comparingSummary = ""
		m.detail.SetContent(msg.value)
		if msg.version == "" {
			m.status = "READ-ONLY"
		} else {
			short := msg.version
			if len(short) > 12 {
				short = short[:12]
			}
			m.status = fmt.Sprintf("viewing version %s (read-only)", short)
		}
	}
	return m, nil
}

func (m Model) onSecretVersionsLoaded(msg secretVersionsLoadedMsg) (Model, tea.Cmd) {
	m.loading = false
	if msg.err != nil {
		m.err = msg.err
		m.status = "error listing versions: " + msg.err.Error()
		return m, nil
	}
	items := make([]list.Item, len(msg.versions))
	for i, v := range msg.versions {
		items[i] = versionItem{version: v, marked: m.markedVersions[v.Version]}
	}
	m.versionList.SetItems(items)
	m.status = fmt.Sprintf("%d versions — space to mark, enter to compare 2+", len(msg.versions))
	return m, nil
}

// onSecretVersionValuesLoaded renders two or more marked versions'
// values stacked in the detail pane for side-by-side comparison. Always
// read-only, regardless of edit mode elsewhere.
func (m Model) onSecretVersionValuesLoaded(msg secretVersionValuesLoadedMsg) (Model, tea.Cmd) {
	m.loading = false
	if msg.err != nil {
		m.err = msg.err
		m.status = "error comparing versions: " + msg.err.Error()
		return m, nil
	}
	if m.currentVaultName != msg.vault || m.currentSecretName != msg.name {
		return m, nil
	}

	blocks := make([]string, len(msg.entries))
	for i, e := range msg.entries {
		short := e.Version
		if len(short) > 12 {
			short = short[:12]
		}
		blocks[i] = fmt.Sprintf("── %s · %s ──\n%s", short, e.Created, e.Value)
	}
	m.detail.SetContent(strings.Join(blocks, "\n\n"))
	m.currentVersion = ""
	m.comparingCount = len(msg.entries)
	m.comparingSummary = fmt.Sprintf("comparing %d versions", len(msg.entries))
	m.status = m.comparingSummary + " — read-only"
	return m, nil
}

func (m Model) onSecretSaved(msg secretSavedMsg) (Model, tea.Cmd) {
	m.loading = false
	if msg.err != nil {
		m.err = msg.err
		m.status = "save failed: " + msg.err.Error()
		return m, nil
	}
	value := m.editArea.Value()
	m.editMode = false
	m.editArea.Blur()
	m.currentValue = value
	m.detail.SetContent(value)
	m.status = "saved new version"

	client, err := m.clientFor(msg.vault)
	if err != nil {
		return m, nil
	}
	return m, fetchSecretVersionsCmd(client, msg.vault, msg.name)
}
