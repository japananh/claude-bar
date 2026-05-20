# Claude Bar

A macOS menu-bar app for managing multiple [Claude Code](https://claude.ai/code) accounts. Switch between accounts instantly, auto-swap when quota runs out, and keep your IDE in sync.

![macOS 14+](https://img.shields.io/badge/macOS-14%2B-blue) ![License MIT](https://img.shields.io/badge/license-MIT-green)

---

## Install via Homebrew

```bash
brew tap ncthanhngo/tap
brew install --cask claude-bar
```

---

## Features

- **Multi-account management** — add, rename, and switch Claude Code accounts from the menu bar
- **Usage display** — 5-hour and 7-day quota bars with % and time-until-reset for each account
- **Auto-swap** — automatically switches to a lower-usage account when the active one hits your threshold; waits for `claude` to exit first, then notifies you
- **IDE reload** — after a swap, reloads VSCode / Cursor / Windsurf windows so the extension picks up new credentials (requires Accessibility permission)
- **CLI auto-restart** — sends SIGINT to running `claude` sessions; use with the bundled `claude-watch` wrapper to auto-restart in your terminal
- **Session guard** — warns you if Claude is running before a manual switch; option to force-switch anyway
- **Web fallback** — embedded WKWebView for fetching usage data when the Anthropic API is rate-limited
- **Themes** — Light, Dark, and Rainbow

---

## Manual install

1. Download `ClaudeBar.zip` from [Releases](https://github.com/ncthanhngo/claude-bar/releases/latest)
2. Unzip and drag `ClaudeBar.app` to `/Applications`
3. Launch — the icon appears in your menu bar

---

## Setup

### Add accounts

Open **Settings → Accounts → Add account**. Each account needs a separate Claude Code login (`claude /login`).

### Auto-restart terminal sessions

Enable **Auto-kill CLI sessions** in Settings → General, then run once in your terminal:

```bash
# Install claude-watch (written by Claude Bar on first launch)
chmod +x "$HOME/Library/Application Support/claude-bar/claude-watch.sh" \
  && ln -sf "$HOME/Library/Application Support/claude-bar/claude-watch.sh" \
     /opt/homebrew/bin/claude-watch

# Make every `claude` call auto-restart after a swap
echo 'alias claude="claude-watch"' >> ~/.zshrc && source ~/.zshrc
```

After this, when a swap happens your terminal session (including GoLand's integrated terminal) restarts automatically with the new account credentials.

### IDE reload — VSCode / Cursor / Windsurf

Enable **Auto-reload IDE after swap** in Settings → General, then click **Grant Access** when the Accessibility prompt appears. Claude Bar will send `⌘⇧P → Developer: Reload Window` after each swap.

---

## Build from source

Requirements: macOS 14+, Xcode 15+, Go 1.23+

```bash
git clone https://github.com/ncthanhngo/claude-bar.git
cd claude-bar
make install      # builds and copies to /Applications/ClaudeBar.app
```

---

## How auto-swap works

1. **Polls usage** every N seconds with adaptive frequency — faster as you approach the threshold
2. Active 5h usage ≥ threshold → notification **"Auto-swap pending (X% used)"** — you are warned before anything happens
3. **Waits** until no `claude` sessions are busy (`safeToSwap = true`) — Claude Bar never interrupts a running session
4. **Swaps** to the inactive account with the lowest 5-hour usage
5. Notification **"Switched to [account]"** — confirmation after the swap completes
6. Triggers **IDE reload** (VSCode/Cursor/Windsurf) and **CLI restart** (`claude-watch`) if those options are enabled in Settings

If all inactive accounts are also above the threshold → notification **"All accounts above threshold"**, retry in 10 minutes.

> Auto-swap is driven by the **5-hour** window exclusively. The 7-day window is displayed for reference but does not affect swap decisions.

---

## License

MIT
