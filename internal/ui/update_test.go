package ui

import (
	"strings"
	"testing"
)

// loadSecret drives the real selectVault/selectSecret flow (including
// executing the tea.Cmds they return through the mock backend) so tests
// exercise the same path the running app does, ending with a secret's
// value and version history loaded.
func loadSecret(t *testing.T, m Model, vault, secret string) Model {
	t.Helper()

	mm, cmd := m.selectVault(vault)
	mm = applyAll(mm, collectMsgs(cmd))

	mm, cmd = mm.selectSecret(secret)
	mm = applyAll(mm, collectMsgs(cmd))

	if mm.currentSecretName != secret {
		t.Fatalf("loadSecret(%q, %q): currentSecretName = %q", vault, secret, mm.currentSecretName)
	}
	if mm.currentValue == "" {
		t.Fatalf("loadSecret(%q, %q): currentValue not populated", vault, secret)
	}
	if len(mm.versionList.Items()) == 0 {
		t.Fatalf("loadSecret(%q, %q): versionList not populated", vault, secret)
	}
	return mm
}

func TestPaneNavigation(t *testing.T) {
	tests := []struct {
		name       string
		from       focus
		key        string
		wantFocus  focus
		handleFunc func(Model, string) (Model, bool)
	}{
		{"vaults right -> secrets", focusVaults, "right", focusSecrets, callVaults},
		{"vaults left -> detail", focusVaults, "left", focusDetail, callVaults},
		{"secrets right -> versions", focusSecrets, "right", focusVersions, callSecrets},
		{"secrets left -> vaults", focusSecrets, "left", focusVaults, callSecrets},
		{"versions right -> detail", focusVersions, "right", focusDetail, callVersions},
		{"versions left -> secrets", focusVersions, "left", focusSecrets, callVersions},
		{"detail right -> vaults", focusDetail, "right", focusVaults, callDetail},
		{"detail left -> versions", focusDetail, "left", focusVersions, callDetail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel()
			m.focus = tt.from
			got, _ := tt.handleFunc(m, tt.key)
			if got.focus != tt.wantFocus {
				t.Fatalf("focus = %v, want %v", got.focus, tt.wantFocus)
			}
		})
	}
}

func callVaults(m Model, key string) (Model, bool) {
	tm, _ := m.handleVaultsKey(keyMsg(key))
	return tm.(Model), true
}
func callSecrets(m Model, key string) (Model, bool) {
	tm, _ := m.handleSecretsKey(keyMsg(key))
	return tm.(Model), true
}
func callVersions(m Model, key string) (Model, bool) {
	tm, _ := m.handleVersionsKey(keyMsg(key))
	return tm.(Model), true
}
func callDetail(m Model, key string) (Model, bool) {
	tm, _ := m.handleDetailKey(keyMsg(key))
	return tm.(Model), true
}

func TestSelectVault(t *testing.T) {
	m := newTestModel()

	mm, cmd := m.selectVault("mock")
	if mm.focus != focusSecrets {
		t.Fatalf("focus after selectVault = %v, want focusSecrets", mm.focus)
	}
	if mm.currentVaultName != "mock" {
		t.Fatalf("currentVaultName = %q, want %q", mm.currentVaultName, "mock")
	}
	if cmd == nil {
		t.Fatal("selectVault() returned nil cmd, want a fetch command on first (uncached) selection")
	}

	msgs := collectMsgs(cmd)
	mm = applyAll(mm, msgs)
	if len(mm.secretList.Items()) == 0 {
		t.Fatal("secretList not populated after fetch completes")
	}
	if !strings.Contains(mm.status, "5 secrets") {
		t.Fatalf("status = %q, want it to mention 5 secrets", mm.status)
	}

	// Re-selecting the same vault should hit the name cache and skip the
	// fetch entirely.
	mm2, cmd2 := mm.selectVault("mock")
	if cmd2 != nil {
		t.Fatal("selectVault() returned non-nil cmd on cached selection, want nil")
	}
	if !strings.Contains(mm2.status, "cached") {
		t.Fatalf("status = %q, want it to mention the cache", mm2.status)
	}
}

func TestSelectSecret_LoadsValueAndVersionsWithoutMovingFocus(t *testing.T) {
	m := newTestModel()
	m.focus = focusSecrets // simulate arriving here via selectVault

	mm := loadSecret(t, m, "mock", "api-signing-key")

	if mm.focus != focusSecrets {
		t.Fatalf("focus after selectSecret = %v, want focusSecrets (selecting a secret must not move focus)", mm.focus)
	}
	if !strings.Contains(mm.currentValue, "MOCK-API-SIGNING-KEY") {
		t.Fatalf("currentValue = %q, want the seeded api-signing-key value", mm.currentValue)
	}
	if len(mm.versionList.Items()) != 1 {
		t.Fatalf("versionList has %d items, want 1 (api-signing-key has a single seeded version)", len(mm.versionList.Items()))
	}
}

func TestSelectSecret_ResetsStateFromPreviousSecret(t *testing.T) {
	m := newTestModel()
	mm := loadSecret(t, m, "mock", "smtp-password")
	mm.markedVersions["some-fake-version"] = true
	mm.comparingCount = 2
	mm.editMode = true

	mm = loadSecret(t, mm, "mock", "api-signing-key")

	if len(mm.markedVersions) != 0 {
		t.Fatalf("markedVersions = %v, want empty after switching secrets", mm.markedVersions)
	}
	if mm.comparingCount != 0 {
		t.Fatalf("comparingCount = %d, want 0 after switching secrets", mm.comparingCount)
	}
	if mm.editMode {
		t.Fatal("editMode still true after switching secrets, want false")
	}
}

func TestEnterEditMode_NoSecretLoaded(t *testing.T) {
	m := newTestModel()
	got, cmd := m.enterEditMode()
	if got.editMode {
		t.Fatal("editMode = true with no secret loaded, want false (no-op)")
	}
	if cmd != nil {
		t.Fatal("enterEditMode() returned non-nil cmd with no secret loaded")
	}
}

func TestEnterEditMode_BlockedWhileComparing(t *testing.T) {
	m := newTestModel()
	m = loadSecret(t, m, "mock", "smtp-password")
	m.comparingCount = 2

	got, _ := m.enterEditMode()
	if got.editMode {
		t.Fatal("editMode = true while comparingCount > 0, want blocked")
	}
	if !strings.Contains(got.status, "cannot edit while comparing") {
		t.Fatalf("status = %q, want a message explaining why edit was blocked", got.status)
	}
}

func TestEnterEditMode_RemembersOriginFocus(t *testing.T) {
	for _, origin := range []focus{focusSecrets, focusVersions, focusDetail} {
		m := newTestModel()
		m = loadSecret(t, m, "mock", "smtp-password")
		m.focus = origin

		got, _ := m.enterEditMode()
		if !got.editMode {
			t.Fatalf("origin %v: editMode = false, want true", origin)
		}
		if got.focus != focusDetail {
			t.Fatalf("origin %v: focus = %v while editing, want focusDetail", origin, got.focus)
		}
		if got.editOriginFocus != origin {
			t.Fatalf("origin %v: editOriginFocus = %v, want %v", origin, got.editOriginFocus, origin)
		}
	}
}

func TestHandleEditKey_EscRestoresOriginFocus(t *testing.T) {
	m := newTestModel()
	m = loadSecret(t, m, "mock", "smtp-password")
	m.focus = focusVersions

	edit, _ := m.enterEditMode()
	tm, _ := edit.handleEditKey(keyMsg("esc"))
	after := tm.(Model)

	if after.editMode {
		t.Fatal("editMode still true after esc, want false")
	}
	if after.focus != focusVersions {
		t.Fatalf("focus after esc = %v, want focusVersions (the pane edit was entered from)", after.focus)
	}
}

func TestHandleEditKey_CtrlSSavesAndRestoresOriginFocus(t *testing.T) {
	m := newTestModel()
	m = loadSecret(t, m, "mock", "smtp-password")
	m.focus = focusSecrets

	edit, _ := m.enterEditMode()
	edit.editArea.SetValue("new value from test")

	tm, cmd := edit.handleEditKey(keyMsg("ctrl+s"))
	saving := tm.(Model)
	if !saving.editMode {
		t.Fatal("editMode = false immediately after ctrl+s fires the save command, want still true until the save completes")
	}
	if cmd == nil {
		t.Fatal("handleEditKey(ctrl+s) returned nil cmd, want the save command")
	}

	msgs := collectMsgs(cmd)
	after := applyAll(saving, msgs)

	if after.editMode {
		t.Fatal("editMode still true after save completes, want false")
	}
	if after.focus != focusSecrets {
		t.Fatalf("focus after save = %v, want focusSecrets (the pane edit was entered from)", after.focus)
	}
	if after.currentValue != "new value from test" {
		t.Fatalf("currentValue = %q, want the saved value", after.currentValue)
	}
}

func TestHandleVersionsKey_SpaceTogglesMarking(t *testing.T) {
	m := newTestModel()
	m = loadSecret(t, m, "mock", "database-connection-string") // seeded with 2 versions
	if len(m.versionList.Items()) != 2 {
		t.Fatalf("versionList has %d items, want 2", len(m.versionList.Items()))
	}

	tm, cmd := m.handleVersionsKey(keyMsg(" "))
	m = applyAll(tm.(Model), collectMsgs(cmd))

	if len(m.markedVersions) != 1 {
		t.Fatalf("markedVersions = %v, want exactly 1 entry after marking", m.markedVersions)
	}
	item, ok := m.versionList.SelectedItem().(versionItem)
	if !ok || !item.marked {
		t.Fatal("selected versionItem.marked = false after space, want true")
	}

	// Toggling the same item again should unmark it.
	tm, cmd = m.handleVersionsKey(keyMsg(" "))
	m = applyAll(tm.(Model), collectMsgs(cmd))
	if len(m.markedVersions) != 0 {
		t.Fatalf("markedVersions = %v, want empty after unmarking", m.markedVersions)
	}
}

func TestHandleVersionsKey_EnterPreviewsWithoutMarking(t *testing.T) {
	m := newTestModel()
	m = loadSecret(t, m, "mock", "database-connection-string")

	tm, cmd := m.handleVersionsKey(keyMsg("enter"))
	m = applyAll(tm.(Model), collectMsgs(cmd))

	if len(m.markedVersions) != 0 {
		t.Fatalf("markedVersions = %v, want empty — enter must not mark", m.markedVersions)
	}
	if m.comparingCount != 0 {
		t.Fatalf("comparingCount = %d, want 0 after a plain preview", m.comparingCount)
	}
	if m.currentVersion == "" {
		t.Fatal("currentVersion empty after enter, want the previewed version's ID")
	}
}

func TestPreviewMarkedVersions_TwoOrMoreCompares(t *testing.T) {
	m := newTestModel()
	m = loadSecret(t, m, "mock", "database-connection-string")

	// Mark both seeded versions for comparison.
	for range m.versionList.Items() {
		tm, cmd := m.handleVersionsKey(keyMsg(" "))
		m = applyAll(tm.(Model), collectMsgs(cmd))
		tm, _ = m.handleVersionsKey(keyMsg("j"))
		m = tm.(Model)
	}

	if len(m.markedVersions) != 2 {
		t.Fatalf("markedVersions = %v, want 2 entries", m.markedVersions)
	}
	if m.comparingCount != 2 {
		t.Fatalf("comparingCount = %d, want 2 after marking both versions", m.comparingCount)
	}
	if !strings.Contains(m.comparingSummary, "2 versions") {
		t.Fatalf("comparingSummary = %q, want it to mention 2 versions", m.comparingSummary)
	}
}

func TestOnSecretValueLoaded_IgnoresStaleResponse(t *testing.T) {
	m := newTestModel()
	m = loadSecret(t, m, "mock", "smtp-password")
	staleValue := m.currentValue

	// Simulate a slow response for a secret the user has since navigated
	// away from.
	stale := secretValueLoadedMsg{vault: "mock", name: "api-signing-key", value: "should not apply"}
	tm, _ := m.Update(stale)
	after := tm.(Model)

	if after.currentValue != staleValue {
		t.Fatalf("currentValue = %q, want unchanged %q (stale response for a different secret must be ignored)", after.currentValue, staleValue)
	}
}

func TestOnSecretVersionsLoaded_IgnoresStaleResponse(t *testing.T) {
	m := newTestModel()
	m = loadSecret(t, m, "mock", "smtp-password")
	itemCount := len(m.versionList.Items())

	stale := secretVersionsLoadedMsg{vault: "mock", name: "api-signing-key", versions: nil}
	tm, _ := m.Update(stale)
	after := tm.(Model)

	if len(after.versionList.Items()) != itemCount {
		t.Fatalf("versionList has %d items, want unchanged %d (stale response for a different secret must be ignored)", len(after.versionList.Items()), itemCount)
	}
}

func TestOnSecretSaved_Error(t *testing.T) {
	m := newTestModel()
	m = loadSecret(t, m, "mock", "smtp-password")
	edit, _ := m.enterEditMode()

	after, _ := edit.onSecretSaved(secretSavedMsg{vault: "mock", name: "smtp-password", err: strErr("boom")})

	if !after.editMode {
		t.Fatal("editMode = false after a failed save, want to stay in edit mode so the user doesn't lose their input")
	}
	if !strings.Contains(after.status, "save failed") {
		t.Fatalf("status = %q, want it to mention the save failure", after.status)
	}
}

type strErr string

func (e strErr) Error() string { return string(e) }
