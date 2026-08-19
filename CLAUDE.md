# kvex — Project Notes for Claude Code

Terminal UI (bubbletea) for browsing/filtering secrets across multiple Azure
Key Vaults. See [README.md](README.md) for the full feature description,
keybindings, and usage.

## GitHub

- Repo: `kubawan/kvex` (origin, default branch `main`)
- The repo is linked to a GitHub Project with an **auto-add workflow**
  enabled — any issue created in this repo is automatically added to the
  project board. No extra step needed to get an issue onto the board.
- **Write all issue titles and bodies in English**, regardless of the
  language the conversation is happening in.

## Development

```bash
go build ./...
go vet ./...
gofmt -l .
```

Run against the mock vault (no Azure access needed) instead of installing:

```bash
go run ./cmd/kvex --config examples/config.yaml
```

## Project layout

```
cmd/kvex/         entrypoint: config load, credential setup, bubbletea run
internal/azure/    thin wrapper around azsecrets.Client
internal/config/   YAML config loading
internal/ui/       bubbletea model/update/view, split by pane
```

Inside `internal/ui/`:
- `model.go` — `Model` struct, focus states (`focusVaults`, `focusSecrets`,
  `focusDetail`, `focusVersions`), layout sizing
- `update.go` — key handling per focus pane, message handling
- `view.go` — rendering, pane layout, banners
- `items.go` — list item adapters (`versionItem`, etc.)
- `messages.go` — bubbletea messages and the `tea.Cmd`s that fetch data via
  `azure.SecretsClient`

## Conventions

- No local persistence of secret values, ever — in-memory cache of names
  only (see README "Non-goals"). Don't add caching/persistence for secret
  values without checking with the user first.
- Default mode is read-only; edit mode is explicit (`e`) and always shows a
  persistent banner. Preserve this when touching edit-related code.
- The mock vault (`mock://` URI, `internal/azure`) is the primary way to
  develop/test the UI without real Azure access — prefer it over requiring
  live Key Vault credentials when testing changes locally.
