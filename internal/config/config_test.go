package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writing temp config: %v", err)
	}
	return path
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr string // substring expected in the error, "" means no error
	}{
		{
			name: "valid single vault",
			content: `
vaults:
  - name: dev
    uri: https://dev-kv.vault.azure.net/
`,
		},
		{
			name: "valid multiple vaults with optional subscription",
			content: `
vaults:
  - name: dev
    uri: https://dev-kv.vault.azure.net/
  - name: prod
    uri: https://prod-kv.vault.azure.net/
    subscription: my-sub
`,
		},
		{
			name:    "no vaults",
			content: `vaults: []`,
			wantErr: "defines no vaults",
		},
		{
			name:    "empty file",
			content: ``,
			wantErr: "defines no vaults",
		},
		{
			name: "vault missing name",
			content: `
vaults:
  - uri: https://dev-kv.vault.azure.net/
`,
			wantErr: "has no name",
		},
		{
			name: "vault missing uri",
			content: `
vaults:
  - name: dev
`,
			wantErr: "has no uri",
		},
		{
			name: "duplicate vault name",
			content: `
vaults:
  - name: dev
    uri: https://dev-kv.vault.azure.net/
  - name: dev
    uri: https://dev2-kv.vault.azure.net/
`,
			wantErr: "duplicate vault name",
		},
		{
			name:    "invalid yaml",
			content: "vaults: [this is not valid yaml:::",
			wantErr: "parsing config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeTempConfig(t, tt.content)
			cfg, err := Load(path)

			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Load() unexpected error: %v", err)
				}
				if cfg == nil || len(cfg.Vaults) == 0 {
					t.Fatalf("Load() returned no vaults")
				}
				return
			}

			if err == nil {
				t.Fatalf("Load() expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Load() error = %q, want substring %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "does-not-exist.yaml"))
	if err == nil {
		t.Fatal("Load() expected error for missing file, got nil")
	}
	if !strings.Contains(err.Error(), "reading config") {
		t.Fatalf("Load() error = %q, want substring %q", err.Error(), "reading config")
	}
}

func TestDefaultPath(t *testing.T) {
	path, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath() unexpected error: %v", err)
	}
	if !strings.HasSuffix(path, filepath.Join(".config", "kvex", "config.yaml")) {
		t.Fatalf("DefaultPath() = %q, want suffix %q", path, filepath.Join(".config", "kvex", "config.yaml"))
	}
}
