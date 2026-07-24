// Command kvex (Key Vault Explorer) is a terminal UI for browsing and
// comparing secrets across multiple Azure Key Vaults.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	tea "github.com/charmbracelet/bubbletea"

	"kvex/internal/config"
	"kvex/internal/ui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "kvex:", err)
		os.Exit(1)
	}
}

func run() error {
	defaultPath, err := config.DefaultPath()
	if err != nil {
		return err
	}

	configPath := flag.String("config", defaultPath, "path to kvex config.yaml")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("%w\n\nExample config:\n\nvaults:\n  - name: dev\n    uri: https://dev-kv.vault.azure.net/\n  - name: prod\n    uri: https://prod-kv.vault.azure.net/", err)
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return fmt.Errorf("creating Azure credential (run `az login`?): %w", err)
	}

	model := ui.New(cfg, cred)
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
