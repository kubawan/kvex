package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestKeyHints_ContextualPerPane(t *testing.T) {
	tests := []struct {
		name           string
		focus          focus
		secretLoaded   bool
		wantContains   []string
		wantNotContain []string
	}{
		{
			name:           "vaults",
			focus:          focusVaults,
			wantContains:   []string{"enter: select", "q: quit"},
			wantNotContain: []string{"space", "e: edit", "←/→"},
		},
		{
			name:           "secrets without a loaded secret",
			focus:          focusSecrets,
			secretLoaded:   false,
			wantContains:   []string{"←/→: switch pane", "enter: select"},
			wantNotContain: []string{"e: edit", "space"},
		},
		{
			name:           "secrets with a loaded secret",
			focus:          focusSecrets,
			secretLoaded:   true,
			wantContains:   []string{"e: edit"},
			wantNotContain: []string{"space"},
		},
		{
			name:         "versions",
			focus:        focusVersions,
			wantContains: []string{"space/x: mark", "enter: preview", "e: edit"},
		},
		{
			name:           "detail",
			focus:          focusDetail,
			wantContains:   []string{"←/→: switch pane", "e: edit", "q: quit"},
			wantNotContain: []string{"space", "enter"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel()
			m.focus = tt.focus
			if tt.secretLoaded {
				m.currentSecretName = "some-secret"
			}

			hints := m.keyHints()

			for _, want := range tt.wantContains {
				if !strings.Contains(hints, want) {
					t.Errorf("keyHints() = %q, want it to contain %q", hints, want)
				}
			}
			for _, notWant := range tt.wantNotContain {
				if strings.Contains(hints, notWant) {
					t.Errorf("keyHints() = %q, want it NOT to contain %q", hints, notWant)
				}
			}
		})
	}
}

func TestKeyHints_CopyHintOnlyWhenThereIsSomethingToCopy(t *testing.T) {
	m := newTestModel()
	if strings.Contains(m.keyHints(), "c: copy") {
		t.Fatal("keyHints() advertises c: copy with no value loaded")
	}

	m.currentValue = "some secret value"
	if !strings.Contains(m.keyHints(), "c: copy") {
		t.Fatal("keyHints() doesn't advertise c: copy once a value is loaded")
	}

	m.comparingCount = 2
	if strings.Contains(m.keyHints(), "c: copy") {
		t.Fatal("keyHints() still advertises c: copy while comparing versions, where copy is blocked")
	}
}

func TestHeaderLine_Breadcrumb(t *testing.T) {
	m := newTestModel()
	if strings.Contains(m.headerLine(), "›") {
		t.Fatal("headerLine() shows a breadcrumb separator with nothing selected")
	}

	m.currentVaultName = "mock"
	hdr := m.headerLine()
	if !strings.Contains(hdr, "kvex") || !strings.Contains(hdr, "mock") || !strings.Contains(hdr, "›") {
		t.Fatalf("headerLine() = %q, want it to contain kvex › mock", hdr)
	}
	if strings.Contains(hdr, "Azure Key Vault Explorer") {
		t.Fatal("headerLine() still shows the plain subtitle once a vault is selected")
	}

	m.currentSecretName = "api-signing-key"
	hdr = m.headerLine()
	if !strings.Contains(hdr, "api-signing-key") {
		t.Fatalf("headerLine() = %q, want it to also contain the selected secret name", hdr)
	}
}

func TestCurrentVersionCreated(t *testing.T) {
	m := newTestModel()
	if _, ok := m.currentVersionCreated(); ok {
		t.Fatal("currentVersionCreated() ok with an empty versionList")
	}

	m = loadSecret(t, m, "mock", "database-connection-string") // seeded with 2 versions

	// currentVersion == "" means "latest", i.e. versionList's first item —
	// the list is always sorted newest-first.
	created, ok := m.currentVersionCreated()
	if !ok || created == "" {
		t.Fatal("currentVersionCreated() not ok for the latest version")
	}
	items := m.versionList.Items()
	want := items[0].(versionItem).version.Created
	if created != want {
		t.Fatalf("currentVersionCreated() = %q, want the newest item's Created %q", created, want)
	}

	// Pin to the oldest version explicitly and confirm it's looked up by ID,
	// not just always returning the first item.
	oldest := items[len(items)-1].(versionItem).version
	m.currentVersion = oldest.Version
	created, ok = m.currentVersionCreated()
	if !ok || created != oldest.Created {
		t.Fatalf("currentVersionCreated() = %q, ok=%v, want %q for the pinned oldest version", created, ok, oldest.Created)
	}

	m.currentVersion = "not-a-real-version-id"
	if _, ok := m.currentVersionCreated(); ok {
		t.Fatal("currentVersionCreated() ok for a version ID not in versionList")
	}
}

func TestDetailBorderColor(t *testing.T) {
	tests := []struct {
		name     string
		editMode bool
		focus    focus
		want     lipgloss.Color
	}{
		{"edit mode overrides everything", true, focusVaults, lipgloss.Color("196")},
		{"focused, not editing", false, focusDetail, focusedBorderColor},
		{"unfocused, not editing", false, focusSecrets, blurredBorderColor},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel()
			m.editMode = tt.editMode
			m.focus = tt.focus

			if got := m.detailBorderColor(); got != tt.want {
				t.Errorf("detailBorderColor() = %v, want %v", got, tt.want)
			}
		})
	}
}
