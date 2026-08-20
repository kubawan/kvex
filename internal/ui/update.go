package ui

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
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
	case "c":
		return m.copySecretValue()
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
		m.focus = focusVersions
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
	case "c":
		return m.copySecretValue()
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
		m.focus = focusVersions
		return m, nil
	case "e":
		return m.enterEditMode()
	case "c":
		return m.copySecretValue()
	}
	var cmd tea.Cmd
	m.detail, cmd = m.detail.Update(msg)
	return m, cmd
}

// enterEditMode is shared by handleDetailKey and handleSecretsKey so 'e'
// works whether or not the user has explicitly moved focus into the detail
// pane — you shouldn't have to navigate panes just to edit what you're
// already previewing. Editing a historical version is allowed: Key Vault
// has no "edit in place" for an old version, so saving creates a new
// (now-current) version seeded from that historical value. Comparing 2+
// versions has no single coherent value to edit, so that stays blocked.
func (m Model) enterEditMode() (Model, tea.Cmd) {
	if m.currentSecretName == "" {
		return m, nil
	}
	if m.comparingCount > 0 {
		m.status = "cannot edit while comparing versions — view a single version first"
		return m, nil
	}
	m.editMode = true
	m.editArea.SetValue(m.currentValue)
	m.editArea.Focus()
	m.editOriginFocus = m.focus
	m.focus = focusDetail
	if m.currentVersion != "" {
		m.status = "EDIT MODE (from historical version — saves as a new current version) — ctrl+s to save, esc to cancel"
	} else {
		m.status = "EDIT MODE — ctrl+s to save, esc to cancel"
	}
	return m, nil
}

// copySecretValue copies whatever is currently loaded in the detail pane to
// the system clipboard. Unlike enterEditMode, it works from any pane and
// never moves focus — it's a read-only action, so there's no reason to
// detour through edit mode (or even the detail pane) just to grab a value.
func (m Model) copySecretValue() (Model, tea.Cmd) {
	if m.comparingCount > 0 {
		m.status = "cannot copy while comparing versions — view a single version first"
		return m, nil
	}
	if m.currentValue == "" {
		return m, nil
	}
	if err := clipboard.WriteAll(m.currentValue); err != nil {
		m.err = err
		m.status = "copy failed: " + err.Error()
		return m, nil
	}
	m.err = nil
	m.status = "copied to clipboard"
	return m, nil
}

func (m Model) handleEditKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.editMode = false
		m.editArea.Blur()
		m.focus = m.editOriginFocus
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
	case "right":
		m.focus = focusDetail
		return m, nil
	case "left":
		m.focus = focusSecrets
		return m, nil
	case "e":
		return m.enterEditMode()
	case "c":
		return m.copySecretValue()
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
		listCmd := m.versionList.SetItem(idx, item)
		mm, previewCmd := m.previewMarkedVersions()
		return mm, tea.Batch(listCmd, previewCmd)
	case "enter":
		// enter is a pure preview: it shows the highlighted version without
		// marking it or touching any existing marks. space/x is the only
		// way to mark versions for a multi-version comparison.
		item, ok := m.versionList.SelectedItem().(versionItem)
		if !ok {
			return m, nil
		}
		return m.previewVersion(item.version)
	}
	var cmd tea.Cmd
	m.versionList, cmd = m.versionList.Update(msg)
	return m, cmd
}

// previewMarkedVersions fetches and displays whatever the current
// markedVersions selection implies — nothing marked is a no-op (leaves
// whatever's already shown), exactly one is a single read-only view, two or
// more is a side-by-side comparison. It deliberately doesn't move focus, so
// marking versions while browsing the list immediately updates the detail
// pane without forcing you to leave the list.
func (m Model) previewMarkedVersions() (Model, tea.Cmd) {
	if len(m.markedVersions) == 0 {
		return m, nil
	}

	client, err := m.clientFor(m.currentVaultName)
	if err != nil {
		m.err = err
		m.status = err.Error()
		return m, nil
	}

	var marked []azure.Version
	for _, it := range m.versionList.Items() {
		vi, ok := it.(versionItem)
		if ok && vi.marked {
			marked = append(marked, vi.version)
		}
	}

	if len(marked) == 1 {
		return m.previewVersion(marked[0])
	}

	m.loading = true
	m.status = fmt.Sprintf("loading %d versions to compare...", len(marked))
	return m, fetchSecretVersionValuesCmd(client, m.currentVaultName, m.currentSecretName, marked)
}

// previewVersion loads a single version's value into the detail pane,
// independent of any marking — used by enter (a pure "look at this one",
// distinct from space/x which marks versions for comparison).
func (m Model) previewVersion(v azure.Version) (Model, tea.Cmd) {
	client, err := m.clientFor(m.currentVaultName)
	if err != nil {
		m.err = err
		m.status = err.Error()
		return m, nil
	}
	m.currentVersion = v.Version
	m.comparingCount = 0
	m.comparingSummary = ""
	m.loading = true
	short := v.Version
	if len(short) > 12 {
		short = short[:12]
	}
	m.status = "loading version " + short + "..."
	return m, fetchSecretValueCmd(client, m.currentVaultName, m.currentSecretName, v.Version)
}

func (m *Model) setSecretItems(names []string) {
	items := make([]list.Item, len(names))
	for i, n := range names {
		items[i] = secretItem{name: n}
	}
	m.secretList.SetItems(items)
	if names == nil {
		m.secretList.Title = "SECRETS"
	} else {
		m.secretList.Title = fmt.Sprintf("SECRETS · %d", len(names))
	}
}

func (m Model) selectVault(name string) (Model, tea.Cmd) {
	m.currentVaultName = name
	m.currentSecretName = ""
	m.currentVersion = ""
	m.currentValue = ""
	m.editMode = false
	m.markedVersions = make(map[string]bool)
	m.comparingCount = 0
	m.comparingSummary = ""
	m.versionList.SetItems(nil)
	m.versionList.Title = "VERSIONS"
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
	m.markedVersions = make(map[string]bool)
	m.comparingCount = 0
	m.comparingSummary = ""
	m.versionList.SetItems(nil)
	m.versionList.Title = "VERSIONS"
	// Deliberately don't move focus to the detail pane here: staying on the
	// secrets list lets you preview values with just up/down + enter,
	// without needing to navigate back after every single secret.
	m.layout()
	m.loading = true
	m.detail.SetContent("loading...")
	m.status = fmt.Sprintf("loading %s...", name)
	return m, tea.Batch(
		fetchSecretValueCmd(client, m.currentVaultName, name, ""),
		fetchSecretVersionsCmd(client, m.currentVaultName, name),
	)
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
	if m.currentVaultName != msg.vault || m.currentSecretName != msg.name {
		return m, nil
	}
	items := make([]list.Item, len(msg.versions))
	for i, v := range msg.versions {
		items[i] = versionItem{version: v, marked: m.markedVersions[v.Version]}
	}
	m.versionList.SetItems(items)
	m.versionList.Title = fmt.Sprintf("VERSIONS · %d", len(msg.versions))
	m.status = fmt.Sprintf("%d versions — space to preview, mark 2+ to compare", len(msg.versions))
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
	m.focus = m.editOriginFocus
	m.currentValue = value
	m.currentVersion = "" // the save just created a new current version
	m.detail.SetContent(value)
	m.status = "saved new version"

	client, err := m.clientFor(msg.vault)
	if err != nil {
		return m, nil
	}
	return m, fetchSecretVersionsCmd(client, msg.vault, msg.name)
}
