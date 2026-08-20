package ui

import (
	"strings"
	"testing"
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
