Version 1.0 Features:
100% portable, your settings, favorites and chats are saved in data folder
100 times smaller than other chat clients
Model filters, favorite models, thinking effort selection
Copy or download generated html documents


# OpenRouter Chat: building from source

A tiny local chat client for OpenRouter. One Go program (`main.go`) that embeds
one web page (`index.html`) and stores everything as JSON files next to the exe.

    source/
      main.go      local server: serves the page, saves settings and chats to disk
      index.html   the entire UI (HTML + CSS + JS, no libraries, no build step)
      go.mod       module file (no third-party dependencies)

## Dependencies

Only one: the **Go toolchain, version 1.19 or newer** (developed and tested with 1.22).

- No third-party Go modules (standard library only), so no internet is needed to build.
- No Node.js, npm, Python, C compiler or CGO.
- No external JS/CSS/fonts in the page. It works fully offline except for the calls to
  `openrouter.ai`, which the browser makes directly.
- At runtime the exe needs a web browser (it opens your default one) and Windows 10 or newer.

### Install Go on Windows

Either run `winget install GoLang.Go` in a terminal, or download the installer from
https://go.dev/dl/. Then open a **new** terminal and check with `go version`.

## Build the Windows exe

In a terminal inside this `source` folder:

    go build -trimpath -ldflags="-s -w -H=windowsgui" -o OpenRouterChat.exe .

What the flags do:

- `-H=windowsgui` makes it a GUI program, so no black console window appears.
  Leave it out while debugging to see errors in a console.
- `-s -w` strips debug info (smaller exe, about 6 MB).
- `-trimpath` removes your local paths from the binary.

Copy `OpenRouterChat.exe` wherever you like. The `data` folder is created next to it
on first run.

### Cross-compiling (from Linux or macOS)

    GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -H=windowsgui" -o OpenRouterChat.exe .

PowerShell equivalent when building for another OS: `$env:GOOS="linux"; go build -o OpenRouterChat .`
Use `GOARCH=arm64` for Windows on ARM.

### Run without building

    go run .

Set the environment variable `NOOPEN=1` to stop it from opening a browser tab (then
browse to http://127.0.0.1:47821 yourself).

## Editing the UI

`index.html` is embedded into the exe at compile time (`//go:embed`), so **rebuild after
every change** (`go run .` is the quickest way to try a change).

## How it works

- **Server:** listens on `127.0.0.1` only, trying ports 47821 to 47830. If another copy is
  already running, a second launch just opens the browser at it and exits.
- **Safety:** requests with an unexpected `Host` or `Origin` header are rejected (403), so
  other websites can't read your saved key through your browser.
- **Lifetime:** the page pings `/api/ping` every 10 s; the program exits when no ping has
  arrived for 150 s, or when `POST /api/quit` is called (Settings > Quit app).
- **API calls:** the browser talks to OpenRouter directly (`/models` and `/chat/completions`
  with streaming). The local server never sees your prompts, only the saved chats.

Endpoints: `GET/PUT /api/settings`, `GET /api/chats`, `GET/PUT/DELETE /api/chats/<id>`,
`GET /api/ping`, `POST /api/quit`.

## Data files (`data/` next to the exe)

- `settings.json`: `{ "key", "favorites": [model ids], "model", "effort" }`
  (the API key is stored as plain text)
- `chats/<id>.json`: `{ "id", "title", "model", "updated", "messages": [ { "role", "content",
  "reasoning"?, "secs"? } ] }`

Files are written atomically (temp file, then rename), so a crash won't corrupt them.

## Troubleshooting

- **`go: ... requires go >= ...` or a toolchain download error:** your Go is older than the
  version in `go.mod`. Update Go, or lower the `go` line (minimum is 1.19).
- **Windows SmartScreen or antivirus flags the exe:** it is unsigned. Building it yourself
  from this source is the cleanest answer. Choose "More info" > "Run anyway".
- **Nothing opens:** the browser launch failed. Open http://127.0.0.1:47821 manually.
- **Page says it can't reach the server:** the program exited (it quits ~2.5 minutes after
  the last tab closed). Start the exe again.
