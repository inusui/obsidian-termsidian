# System Terminal

[![CI and release](https://github.com/inusui/obsidian-termsidian/actions/workflows/release.yml/badge.svg)](https://github.com/inusui/obsidian-termsidian/actions/workflows/release.yml)

A real terminal in Obsidian's right sidebar, running your system shell (PowerShell, cmd, bash, zsh…).

No Python and no node-pty: a small Go program, `termsidian-pty`, opens a real system terminal (ConPTY on Windows, PTY on macOS and Linux), and the plugin renders it with [xterm.js](https://xtermjs.org/).

*System Terminal is an independent project, not affiliated with or endorsed by Obsidian.*

(Léeme en español: [README.es.md](README.es.md))

## Tested systems

| System | Obsidian install | Status |
|---|---|---|
| Windows 11 | Official installer | ✅ Works |
| Linux (Fedora KDE) | Flatpak | ✅ Works (needs [a permission](#linux-on-flatpak)) |
| macOS | — | ⚪ Builds, untested |

## What it does

- Opens a terminal in the right sidebar, starting in the vault folder.
- A real terminal: colors, Tab completion, Ctrl+C, history, and interactive programs.
- Resizes when you resize the panel.
- Uses your Obsidian theme's monospace font and colors, and follows theme changes.
- Leaves no orphan processes when you close the panel.

## Usage

Open the terminal with the terminal icon in the left ribbon, or `Ctrl+P` → **System Terminal: Open terminal**.

The shell is PowerShell on Windows, and on macOS and Linux whatever `$SHELL` points to. If Obsidian is installed via Flatpak, there's one extra step: see [Linux on Flatpak](#linux-on-flatpak).

| Key | What it does |
|---|---|
| `Ctrl+C` | With text selected, copies it. Without a selection, interrupts the running program, as usual. |
| `Ctrl+V` | Pastes the clipboard. |
| `Enter` after `exit` | Opens a new shell. |

Closing the panel kills the shell and everything running inside it. Programs with their own window that you launched from the terminal (for example `code .`) stay open.

## Installing in your vault

**Requirements:** Obsidian desktop 1.7.2 or later, with community plugins enabled (restricted mode off).

### From Community Plugins

Go to **Settings → Community plugins → Browse**, search for **System Terminal**, and click **Install**, then **Enable**.

The helper for your system (Windows, macOS or Linux, Intel or ARM) is written to disk the first time you open a terminal — it ships embedded inside `main.js`, since Obsidian only downloads `main.js`, `manifest.json` and `styles.css` for plugins installed this way.

### Manual install

`<your-vault>` is the folder that contains `.obsidian`.

1. On the [Releases](https://github.com/inusui/obsidian-termsidian/releases) page, download `termsidian-<version>.zip`.
2. Unzip it inside `<your-vault>/.obsidian/plugins/`. On Windows, right-click → **Extract All**. You should end up with a `termsidian` folder inside `plugins`.

   With Git Bash, macOS, or Linux (swap in the version you want):

   ```bash
   cd "<your-vault>/.obsidian/plugins"
   curl -LO https://github.com/inusui/obsidian-termsidian/releases/download/0.1.0/termsidian-0.1.0.zip
   unzip -o termsidian-0.1.0.zip && rm termsidian-0.1.0.zip
   ```

3. In Obsidian, go to **Settings → Community plugins**, click the reload button next to "Installed plugins", and enable **System Terminal**.

The plugin folder looks like this once you've opened a terminal at least once (`bin/` only has the one binary for the system you opened it on):

```
<your-vault>/.obsidian/plugins/termsidian/
 ├─ main.js
 ├─ manifest.json
 ├─ styles.css
 ├─ LICENSE
 ├─ THIRD_PARTY_NOTICES.md
 └─ bin/
     └─ termsidian-pty-<your system>-<your architecture>
```

Prefer to build it yourself instead of downloading the release? See [DEVELOPMENT.md](DEVELOPMENT.md).

### Updating

Through Community Plugins, updates install like any other plugin, from **Settings → Community plugins**. Installed manually, replace the files with a new `termsidian-<version>.zip` (steps above) — disable the plugin or close Obsidian first, since Windows won't let you replace a `.exe` that's in use.

> **Synced vault (Nextcloud, Dropbox, OneDrive…):** `bin/` syncs too, so once you've opened a terminal on each of your machines, the plugin works on all of them, whether Windows, macOS, or Linux.

### Linux on Flatpak

If you installed Obsidian from Flathub, Discover, or GNOME Software, it runs inside a Flatpak sandbox with its own filesystem. Without this step, the terminal opens inside that sandbox: it shows a prompt like `sh-5.3$` and can't find your programs.

System Terminal detects Flatpak and opens your real system shell with `flatpak-spawn --host`. For that, Obsidian needs a permission it doesn't have by default. Run this once in a regular terminal, outside Obsidian:

```bash
flatpak override --user --talk-name=org.freedesktop.Flatpak md.obsidian.Obsidian
```

Then fully quit Obsidian and reopen it. Until the permission is granted, the terminal runs inside the sandbox and shows a yellow warning with this same command.

> **What this permission means:** it lets Obsidian and all its plugins run programs outside the sandbox, same as a non-Flatpak install. That's what a terminal needs to see your system. To remove it:
>
> ```bash
> flatpak override --user --no-talk-name=org.freedesktop.Flatpak md.obsidian.Obsidian
> ```

The terminal starts in the vault folder. If that folder doesn't exist outside the sandbox (for example, a vault opened through the document portal), it starts in your home folder instead.

## Troubleshooting

**The terminal shows "No hay un helper para tu sistema (…)"**

Your system or CPU architecture isn't one of the 6 supported ones (see [Why a Go helper?](#why-a-go-helper)), or `main.js` got corrupted. Reinstall from a release ([Installing in your vault](#installing-in-your-vault)).

**On macOS or Linux, the helper won't start due to a permissions error**

Some sync services don't preserve the execute permission. Restore it:

```bash
chmod +x "<your-vault>/.obsidian/plugins/termsidian/bin/"termsidian-pty-*
```

**On Linux, the prompt is `sh-5.3$` and it can't find my programs**

Obsidian is installed via Flatpak and is missing the permission. See [Linux on Flatpak](#linux-on-flatpak).

**Checking error messages**

Open Obsidian's console with `Ctrl+Shift+I`, **Console** tab. Messages from the helper start with `[termsidian-pty]`.

## Why a Go helper?

- `node-pty` needs to be recompiled for every Electron version, and breaks when Obsidian updates.
- A Python helper needs Python installed, and the `pty` module doesn't exist on Windows.
- Go produces a single, dependency-free binary for each system.

More detail on how the plugin and helper talk to each other in [DEVELOPMENT.md](DEVELOPMENT.md#cómo-funciona).

## Contributing

System Terminal is maintained by a single person. Suggestions and bug reports are welcome as an [issue](https://github.com/inusui/obsidian-termsidian/issues), but **pull requests aren't accepted**: any PR gets closed automatically with a comment linking here. If you want to propose a change, open an issue describing it.

## Security

If you find a vulnerability, don't open a public issue: report it privately as explained in [SECURITY.md](SECURITY.md).

## License

[MIT](LICENSE).
