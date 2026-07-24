# kvex — Key Vault Explorer

A terminal UI for browsing, filtering, and cross-checking secrets across
multiple Azure Key Vaults (dev/test/prod), without going through the Azure
Portal.

`kvex` is explicitly **not** a secrets manager or a Vault/Infisical
replacement — that's what Key Vault already is. It's a read-heavy
productivity tool, closer in spirit to `lazygit`/`lazydocker` than to a
secrets engine.

## Status

MVP in progress. Feature 1 (browse & filter) is implemented end-to-end.
Feature 2 (cross-vault compare) is not built yet.

### Feature 1 — Browse & filter

- Lists secret **names only** per vault (never bulk-fetches values)
- Live substring/fuzzy filter over the cached name list (no API roundtrip
  per keystroke)
- Enter on a secret lazily fetches its current value into the detail pane
- Default mode is **read-only**; `e` toggles edit mode with a persistent,
  hard-to-miss banner
- Version history panel (`v`) — always read-only, even while edit mode is
  active elsewhere

### Feature 2 — Cross-vault compare (not implemented yet)

Will let you pick a secret, select target vaults, and see a MATCH / DIFFER /
NOT_FOUND / NO_ACCESS verdict per vault — values hidden by default, revealed
only with an extra explicit keypress.

## Non-goals

- No local persistence of secret values, ever (in-memory cache of names
  only)
- No secret rotation, no policy engine, no multi-cloud support
- No vault/RBAC management, no bulk write workflows

## Install

Requires Go >= 1.25 and `az login` (auth goes through
`azidentity.NewDefaultAzureCredential`, so anything that credential chain
supports works — Azure CLI, managed identity, env vars, etc.).

```bash
git clone https://github.com/kubawan/kvex.git
cd kvex
./install.sh
```

This builds `kvex` via `go install ./cmd/kvex`, writes a starter config to
`~/.config/kvex/config.yaml` if one doesn't already exist, and warns you if
`$(go env GOPATH)/bin` isn't on your `PATH`.

Or do it by hand:

```bash
go install ./cmd/kvex
```

## Configure

`kvex` reads `~/.config/kvex/config.yaml` by default, or pass `--config
/path/to/config.yaml`:

```yaml
vaults:
  - name: dev
    uri: https://dev-kv.vault.azure.net/
  - name: test
    uri: https://test-kv.vault.azure.net/
    subscription: my-test-subscription # optional, display only
  - name: prod
    uri: https://prod-kv.vault.azure.net/
```

See [examples/config.yaml](examples/config.yaml).

## Usage

```bash
kvex --config examples/config.yaml
```

| Key | Action |
|---|---|
| `tab` / `shift+tab` | cycle focus between panes |
| `↑`/`↓`, `j`/`k` | move selection within a pane |
| `/` | filter the focused list |
| `enter` | select vault / secret / version |
| `e` | toggle edit mode (only on the latest version) |
| `ctrl+s` | save value while in edit mode |
| `esc` | cancel edit / close version panel |
| `v` | toggle version history panel |
| `q` / `ctrl+c` | quit |

## Development

```bash
go build ./...
go vet ./...
gofmt -l .
```

Project layout:

```
cmd/kvex/         entrypoint: config load, credential setup, bubbletea run
internal/azure/    thin wrapper around azsecrets.Client
internal/config/   YAML config loading
internal/ui/       bubbletea model/update/view, split by pane
```

## Stack

- Go
- TUI: [bubbletea](https://github.com/charmbracelet/bubbletea) +
  [bubbles](https://github.com/charmbracelet/bubbles) +
  [lipgloss](https://github.com/charmbracelet/lipgloss)
- Azure SDK: `azidentity` + `azsecrets`

## License

MIT — see [LICENSE](LICENSE).
