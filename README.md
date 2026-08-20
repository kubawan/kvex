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
- Version history panel — always visible alongside the secrets list and
  detail pane, always read-only even while edit mode is active elsewhere

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

### Mock vault (no Azure access needed)

A vault configured with a `mock://` URI is backed by an in-memory fake
instead of a real Key Vault — no credential, no network, pre-seeded with a
handful of sample secrets (including one with multiple versions, to exercise
the version-history panel). Useful for trying kvex out or developing the UI
without Azure access:

```yaml
vaults:
  - name: mock
    uri: mock://local
```

This is already the first entry in [examples/config.yaml](examples/config.yaml),
so a fresh `./install.sh` gives you something to click around in immediately.

## Usage

```bash
kvex --config examples/config.yaml
```

| Key | Action |
|---|---|
| `→`/`←` | cycle focus between panes (vaults ↔ secrets ↔ versions ↔ detail) |
| `↑`/`↓`, `j`/`k` | move selection within a pane |
| `/` | filter the focused list |
| `enter` | select a vault (moves into the secrets pane); in the secrets pane, load the highlighted secret; in the versions pane, preview the highlighted version — none of these move focus except selecting a vault |
| `space` / `x` | mark a version for comparison — immediately previews it (or, with 2+ marked, the comparison) in the detail pane |
| `e` | edit mode, including on a historical version — works from the secrets list and versions panel too |
| `ctrl+s` | save value while in edit mode; returns focus to whichever pane `e` was pressed from |
| `esc` | cancel edit; returns focus to whichever pane `e` was pressed from |
| `q` / `ctrl+c` | quit |

Selecting a secret (`enter` in the secrets pane) loads its value and version
history into the detail and versions panes but keeps focus on the secrets
list, so you can preview several secrets in a row with just `↑`/`↓` + `enter`
— no need to navigate back and forth between panes for a quick look. The
versions pane works the same way: `enter` steps through versions one at a
time without leaving the versions list, and without marking anything —
marking (for a multi-version comparison) is `space`/`x`'s job alone.
`e` also works directly from the secrets list on whatever secret is
currently loaded. Editing moves focus into the detail pane so you can type,
but canceling (`esc`) or saving (`ctrl+s`) sends focus back to wherever `e`
was pressed from — the secrets list, the versions pane, or the detail pane
itself — rather than always leaving you on detail.

Editing is allowed on any single version, not just the latest — Key Vault
has no "edit in place" for an old version, so saving always creates a new
current version seeded from whatever value you were looking at. Comparing
2+ versions has no single value to edit, so that stays blocked.

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
