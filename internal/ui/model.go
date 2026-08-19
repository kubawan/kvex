package ui

import (
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"kvex/internal/azure"
	"kvex/internal/config"
)

// focus identifies which pane currently receives key input.
type focus int

const (
	focusVaults focus = iota
	focusSecrets
	focusDetail
	focusVersions
)

// Model is the top-level bubbletea model for kvex.
type Model struct {
	vaults []config.Vault
	cred   azcore.TokenCredential

	clients          map[string]azure.SecretsClient // vault name -> client, lazy
	secretNamesCache map[string][]string            // vault name -> cached secret names

	vaultList   list.Model
	secretList  list.Model
	versionList list.Model
	detail      viewport.Model
	editArea    textarea.Model

	focus    focus
	editMode bool
	loading  bool
	status   string
	err      error

	currentVaultName  string
	currentSecretName string
	currentVersion    string // "" means latest/current
	currentValue      string

	markedVersions   map[string]bool // version ID -> marked for comparison, scoped to currentSecretName
	comparingCount   int             // number of versions in the currently displayed comparison, 0 if none
	comparingSummary string          // short description shown in the banner while comparing

	width, height int
	ready         bool
}

// New builds the initial Model from the loaded config and credential.
func New(cfg *config.Config, cred azcore.TokenCredential) Model {
	vaultItems := make([]list.Item, len(cfg.Vaults))
	for i, v := range cfg.Vaults {
		vaultItems[i] = vaultItem{vault: v}
	}

	vaultList := list.New(vaultItems, list.NewDefaultDelegate(), 0, 0)
	vaultList.Title = "Vaults"
	vaultList.SetShowHelp(false)

	secretList := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	secretList.Title = "Secrets"
	secretList.SetShowHelp(false)

	versionList := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	versionList.Title = "Versions"
	versionList.SetShowHelp(false)

	detail := viewport.New(0, 0)

	editArea := textarea.New()
	editArea.Placeholder = "secret value"
	editArea.ShowLineNumbers = false

	return Model{
		vaults:           cfg.Vaults,
		cred:             cred,
		clients:          make(map[string]azure.SecretsClient),
		secretNamesCache: make(map[string][]string),
		markedVersions:   make(map[string]bool),
		vaultList:        vaultList,
		secretList:       secretList,
		versionList:      versionList,
		detail:           detail,
		editArea:         editArea,
		focus:            focusVaults,
		status:           "select a vault and press enter",
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

// clientFor returns the (lazily created) client for a vault name. Vaults
// configured with a "mock://" URI get an in-memory MockClient instead of a
// real azsecrets-backed Client, so they work without any Azure credential.
func (m *Model) clientFor(vaultName string) (azure.SecretsClient, error) {
	if c, ok := m.clients[vaultName]; ok {
		return c, nil
	}
	var uri string
	for _, v := range m.vaults {
		if v.Name == vaultName {
			uri = v.URI
			break
		}
	}
	if uri == "" {
		return nil, fmt.Errorf("unknown vault %q", vaultName)
	}

	if strings.HasPrefix(uri, azure.MockScheme) {
		c := azure.NewMockClient()
		m.clients[vaultName] = c
		return c, nil
	}

	c, err := azure.NewClient(uri, m.cred)
	if err != nil {
		return nil, err
	}
	m.clients[vaultName] = c
	return c, nil
}

// layout recomputes pane sizes from the current terminal dimensions.
func (m *Model) layout() {
	if m.width == 0 || m.height == 0 {
		return
	}

	headerHeight := 1
	statusHeight := 1
	bodyHeight := m.height - headerHeight - statusHeight
	if bodyHeight < 3 {
		bodyHeight = 3
	}

	vaultsWidth := m.width * 18 / 100
	secretsWidth := m.width * 34 / 100
	detailWidth := m.width - vaultsWidth - secretsWidth
	versionsWidth := detailWidth * 40 / 100
	detailWidth -= versionsWidth
	m.versionList.SetSize(versionsWidth-2, bodyHeight-2)

	m.vaultList.SetSize(vaultsWidth-2, bodyHeight-2)
	m.secretList.SetSize(secretsWidth-2, bodyHeight-2)
	m.detail.Width = detailWidth - 2
	m.detail.Height = bodyHeight - 4 // leave room for the mode banner
	m.editArea.SetWidth(detailWidth - 2)
	m.editArea.SetHeight(bodyHeight - 4)
}
