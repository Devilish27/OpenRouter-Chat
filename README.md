# OpenRouter Chat

A tiny, fast, portable chat client for [OpenRouter](https://openrouter.ai) models on Windows.
One small exe, no installer, no account, no telemetry. It is not an Electron app, so it
stays light: the exe is about 6 MB.

<img width="1117" height="902" alt="screnshot" src="https://github.com/user-attachments/assets/b9d1826f-6c7e-43cb-83e7-7bd15ac4717e" />

## Download

1. Go to the [**Releases**](https://github.com/Devilish27/OpenRouter-Chat/releases) page and download the latest zip.
2. Unzip it anywhere (a USB stick works too).
3. Double-click `OpenRouterChat.exe`. Your browser opens the chat.

Windows SmartScreen may warn about an unsigned app the first time. Click **More info**, then
**Run anyway**. If you would rather not trust a binary, [build it yourself](#build-from-source)
in one command.

## Features

- **Fully portable.** Settings, favorites and chats live in a `data` folder next to the exe.
  Copy the folder to move everything to another PC.
- **Saved chats.** Every chat is saved automatically. Reopen, rename (double-click) or delete
  them from the sidebar.
- **Any OpenRouter model.** The full model list loads automatically, with a filter box and a
  free-model label.
- **Favorite models.** Click the star next to the model picker. Favorites are listed first.
- **Thinking and effort control.** For models that support reasoning, pick Default, Off or an
  effort level (minimal up to xhigh, whatever the model allows). Thinking streams into a
  collapsible "Thought for N s" section and is saved with the chat.
- **Streaming replies** with a Stop button.
- **Smart scrolling.** The view follows new text only while you are at the bottom. Scroll up
  and it stays put, with a "Latest" button to jump back.
- **Code blocks with Copy buttons**, and a **Download .html** button when the AI writes a
  complete HTML document.
- Follows your system light or dark mode.

## Quick start

1. Create an API key at [openrouter.ai/keys](https://openrouter.ai/keys).
2. In the app, open **Settings** (bottom left) and paste the key. It is saved for next time.
3. Pick a model, type a message, press Enter (Shift+Enter for a new line).

To quit, use **Settings, Quit app**. It also closes itself about two and a half minutes after
you close the browser tab. Starting it again while it runs just reopens the page.

## Privacy and security

- Your browser talks **directly to openrouter.ai**. The local program never sees your prompts
  or replies; it only stores your saved chats and settings on disk.
- The local server listens on `127.0.0.1` only, and rejects requests from other websites.
- Your **API key is stored as plain text** in `data/settings.json`. Do not share or commit
  that folder. If you build inside a clone of this repo, make sure `data/` is in your `.gitignore`.
- There is no analytics, update check or other network traffic besides OpenRouter.

## Build from source

You need only the [Go toolchain](https://go.dev/dl/) 1.19 or newer (tested with 1.22).
There are no third-party modules, no Node.js and no C compiler, and no internet is needed to build.

```
winget install GoLang.Go
go build -trimpath -ldflags="-s -w -H=windowsgui" -o OpenRouterChat.exe .
```

- `-H=windowsgui` hides the console window. Leave it out while debugging.
- `-s -w` shrinks the binary, and `-trimpath` removes local paths from it.
- Quick test without building: `go run .` (set `NOOPEN=1` to stop it opening a browser tab).

Cross-compile from Linux or macOS:

```
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -H=windowsgui" -o OpenRouterChat.exe .
```

The code also builds for Linux and macOS (it opens the default browser with `xdg-open` or
`open`), but only Windows has been tested end to end.

`index.html` is embedded into the exe at compile time, so rebuild after editing it.

## How it works

| File | Role |
| --- | --- |
| `main.go` | Local server on `127.0.0.1` (ports 47821 to 47830). Serves the page and saves settings and chats as JSON. |
| `index.html` | The whole UI: HTML, CSS and JavaScript, with no libraries and no build step. |

The page pings `/api/ping` every 10 seconds, and the program exits when pings stop for 150
seconds. Files are written atomically (temp file, then rename), so a crash will not corrupt them.

Endpoints: `GET/PUT /api/settings`, `GET /api/chats`, `GET/PUT/DELETE /api/chats/<id>`,
`GET /api/ping`, `POST /api/quit`.

### Data files

- `data/settings.json`: `{ "key", "favorites": [model ids], "model", "effort" }`
- `data/chats/<id>.json`: `{ "id", "title", "model", "updated", "messages": [ { "role", "content", "reasoning"?, "secs"? } ] }`

## Troubleshooting

- **SmartScreen or antivirus flags the exe.** It is unsigned. Building it yourself from this
  source is the cleanest answer.
- **Nothing opens.** Browse to <http://127.0.0.1:47821> manually.
- **The page can't reach the server.** The program exited (it quits shortly after the last tab
  closes). Start the exe again.
- **No thinking shown for a model.** Some models (for example OpenAI's o-series) reason
  internally but do not return their reasoning. The effort setting still affects the answer.
- **Build error mentioning a Go version or toolchain.** Update Go to 1.19 or newer.

## Notes

Reasoning tokens are billed as output tokens, so higher effort costs more and is slower.
This project is not affiliated with OpenRouter.

## License

MIT (see `LICENSE`).
