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

func saveSecretCmd(client azure.SecretsClient, vaultName, secretName, value string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
		defer cancel()
		err := client.SetSecret(ctx, secretName, value)
		return secretSavedMsg{vault: vaultName, name: secretName, err: err}
	}
}
