package ui

import (
	"strings"
	"testing"
	"time"

	"kvex/internal/azure"
)

func TestVersionItem_Title(t *testing.T) {
	tests := []struct {
		name         string
		item         versionItem
		wantContains []string
	}{
		{
			name:         "current version",
			item:         versionItem{version: azure.Version{Version: "abc123"}, current: true},
			wantContains: []string{"☐", "abc123", "current"},
		},
		{
			name:         "marked version shows a checked box",
			item:         versionItem{version: azure.Version{Version: "abc123"}, marked: true, current: true},
			wantContains: []string{"☑"},
		},
		{
			name:         "disabled version is flagged",
			item:         versionItem{version: azure.Version{Version: "abc123", Enabled: false}, current: true},
			wantContains: []string{"(disabled)"},
		},
		{
			name: "older version shows a relative age, not the raw timestamp",
			item: versionItem{version: azure.Version{
				Version: "def456",
				Created: time.Now().Add(-3 * 24 * time.Hour).Format("2006-01-02 15:04:05"),
			}},
			wantContains: []string{"3d ago"},
		},
		{
			name:         "unknown created time",
			item:         versionItem{version: azure.Version{Version: "def456", Created: ""}},
			wantContains: []string{"unknown"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title := tt.item.Title()
			for _, want := range tt.wantContains {
				if !strings.Contains(title, want) {
					t.Errorf("Title() = %q, want it to contain %q", title, want)
				}
			}
		})
	}
}

func TestVersionItem_Title_UsesLocalTimeForRelativeAge(t *testing.T) {
	// Regression test: azure.Version.Created is formatted from a local-zone
	// time.Time without ever converting to UTC (see internal/azure), so
	// parsing it back must assume the same zone. Parsing with plain
	// time.Parse (which defaults to UTC) would skew the computed age by
	// the machine's UTC offset — on a non-UTC machine, a version created
	// exactly 3 days ago could misreport as "2d ago" or "4d ago".
	created := time.Now().Add(-3 * 24 * time.Hour)
	item := versionItem{version: azure.Version{
		Version: "abc123",
		Created: created.Format("2006-01-02 15:04:05"),
	}}

	if got := item.Title(); !strings.Contains(got, "3d ago") {
		t.Fatalf("Title() = %q, want it to contain %q (local-time parsing regression)", got, "3d ago")
	}
}

func TestVersionItem_Description_IsAlwaysEmpty(t *testing.T) {
	// Version rows fold their status into Title() and render single-line
	// (see newCompactListDelegate) — Description must stay empty.
	item := versionItem{version: azure.Version{Version: "abc123", Created: "2026-01-01 00:00:00"}}
	if got := item.Description(); got != "" {
		t.Fatalf("Description() = %q, want empty", got)
	}
}
