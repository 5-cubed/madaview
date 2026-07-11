# madaview

A cross-platform (Windows/macOS/Linux) markdown viewer, distributed as a single
self-contained binary. No Node, Python, or JVM install required — download it, run it,
view your markdown in a browser.

Renders GitHub-flavored markdown (tables, task lists, syntax-highlighted code), Mermaid
diagrams, and KaTeX math. Mermaid and KaTeX are hydrated entirely client-side — the
markdown source never leaves your browser for diagram or math rendering.

## Quickstart

1. Download the binary for your OS from the [Releases](../../releases) page:
   - Windows: `madaview-windows-amd64.exe`
   - macOS (Apple Silicon): `madaview-darwin-arm64`
   - macOS (Intel): `madaview-darwin-amd64`
   - Linux: `madaview-linux-amd64`
2. Run it from a terminal in the folder you want to browse:
   ```sh
   ./madaview
   ```
3. Open `http://localhost:4800` in a browser.

By default madaview serves the current working directory. Override the root folder with
`--root <path>`, or change it later from the Settings page in the browser UI. The server
binds to `0.0.0.0`, so other devices on your LAN can also view it at your machine's IP —
content is scoped to the root folder you chose.

### First-run security warnings

The binaries aren't code-signed, so your OS will warn on first run. This is expected —
just a one-time step:

- **macOS**: Gatekeeper will block the first launch. Right-click the binary in Finder and
  choose **Open**, then confirm in the dialog that appears.
- **Windows**: SmartScreen will flag it as unrecognized. Click **More info**, then
  **Run anyway**.

## CLI flags

| Flag | Default | Description |
|---|---|---|
| `--root <path>` | last-used root, or current directory | Folder to serve |
| `--port <n>` | `4800` | Port to listen on |
| `-v`, `--verbose` | off | Debug-level logging |

## Development

Requires Go and Node.js.

```sh
make build   # builds the frontend, then the madaview binary
./madaview
```

`go build`, `go vet`, and `go test ./...` work on a fresh clone without the frontend
build — `internal/assets/dist` ships with a placeholder so the `go:embed` directive
always has something to embed.

See `.context/adr/20260711-173722-madaview-initial-setup.md` for the architecture
decision record covering this project's design.
