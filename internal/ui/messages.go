package ui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"kvex/internal/azure"
)

const apiTimeout = 30 * time.Second

type secretNamesLoadedMsg struct {
	vault string
	names []string
	err   error
}

type secretValueLoadedMsg struct {
	vault   string
	name    string
	version string
	value   string
	err     error
}

type secretVersionsLoadedMsg struct {
	vault    string
	name     string
	versions []azure.Version
	err      error
}

type secretSavedMsg struct {
	vault string
	name  string
	err   error
}

// versionValueEntry pairs a fetched version's value with its metadata, for
// side-by-side display in the detail pane.
type versionValueEntry struct {
	Version string
	Created string
	Value   string
}

type secretVersionValuesLoadedMsg struct {
	vault   string
	name    string
	entries []versionValueEntry
	err     error
}

func fetchSecretNamesCmd(client azure.SecretsClient, vaultName string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
		defer cancel()
		names, err := client.ListSecretNames(ctx)
		return secretNamesLoadedMsg{vault: vaultName, names: names, err: err}
	}
}

func fetchSecretValueCmd(client azure.SecretsClient, vaultName, secretName, version string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
		defer cancel()
		value, err := client.GetSecret(ctx, secretName, version)
		return secretValueLoadedMsg{vault: vaultName, name: secretName, version: version, value: value, err: err}
	}
}

func fetchSecretVersionsCmd(client azure.SecretsClient, vaultName, secretName string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
		defer cancel()
		versions, err := client.ListSecretVersions(ctx, secretName)
		return secretVersionsLoadedMsg{vault: vaultName, name: secretName, versions: versions, err: err}
	}
}

// fetchSecretVersionValuesCmd fetches the value of each given version,
// sequentially, and bundles them into one message for side-by-side display.
func fetchSecretVersionValuesCmd(client azure.SecretsClient, vaultName, secretName string, versions []azure.Version) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
		defer cancel()
		entries := make([]versionValueEntry, 0, len(versions))
		for _, v := range versions {
			value, err := client.GetSecret(ctx, secretName, v.Version)
			if err != nil {
				return secretVersionValuesLoadedMsg{vault: vaultName, name: secretName, err: err}
			}
			entries = append(entries, versionValueEntry{Version: v.Version, Created: v.Created, Value: value})
		}
		return secretVersionValuesLoadedMsg{vault: vaultName, name: secretName, entries: entries}
	}
}

func saveSecretCmd(client azure.SecretsClient, vaultName, secretName, value string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
		defer cancel()
		err := client.SetSecret(ctx, secretName, value)
		return secretSavedMsg{vault: vaultName, name: secretName, err: err}
	}
}
