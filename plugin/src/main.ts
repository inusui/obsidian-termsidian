import { chmodSync, existsSync, mkdirSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { FileSystemAdapter, Plugin, type WorkspaceLeaf } from "obsidian";
import { HELPER_BINARIES } from "./helper-binaries.generated.ts";
import { TerminalView, VIEW_TYPE_TERMINAL } from "./view.ts";

// Nombres de sistema y arquitectura de Node → los de Go (GOOS/GOARCH), que
// son los que llevan los binarios del helper.
const GO_OS: Record<string, string> = { win32: "windows", darwin: "darwin", linux: "linux" };
const GO_ARCH: Record<string, string> = { x64: "amd64", arm64: "arm64" };

export default class TermsidianPlugin extends Plugin {
  async onload(): Promise<void> {
    this.registerView(VIEW_TYPE_TERMINAL, (leaf) => new TerminalView(leaf, this));
    this.addRibbonIcon("terminal", "Abrir terminal", () => this.openTerminal());
    this.addCommand({
      id: "open-terminal",
      name: "Abrir terminal",
      callback: () => this.openTerminal(),
    });
  }

  // Abre la terminal en la barra lateral derecha, o la enfoca si ya existe.
  async openTerminal(): Promise<void> {
    const { workspace } = this.app;
    let leaf: WorkspaceLeaf | null = workspace.getLeavesOfType(VIEW_TYPE_TERMINAL)[0] ?? null;
    if (!leaf) {
      leaf = workspace.getRightLeaf(false);
      if (!leaf) return;
      await leaf.setViewState({ type: VIEW_TYPE_TERMINAL, active: true });
    }
    await workspace.revealLeaf(leaf);
    if (leaf.view instanceof TerminalView) leaf.view.focus();
  }

  // Carpeta del vault en disco: el directorio inicial del shell.
  vaultPath(): string {
    const adapter = this.app.vault.adapter;
    return adapter instanceof FileSystemAdapter ? adapter.getBasePath() : process.cwd();
  }

  // Ruta del helper para este sistema, dentro de la carpeta del plugin:
  // <vault>/.obsidian/plugins/termsidian/bin/termsidian-pty-windows-amd64.exe
  helperPath(): string {
    const os = GO_OS[process.platform] ?? process.platform;
    const arch = GO_ARCH[process.arch] ?? process.arch;
    const ext = process.platform === "win32" ? ".exe" : "";
    return join(this.vaultPath(), this.manifest.dir ?? "", "bin", `termsidian-pty-${os}-${arch}${ext}`);
  }

  // El listado de Community Plugins de Obsidian solo descarga main.js,
  // manifest.json y styles.css: bin/ no llega instalado. main.js lleva los 6
  // binarios del helper embebidos (en base64, generados por
  // generate-helper-binaries.mjs); la primera vez que hace falta el helper,
  // lo escribimos a disco en la ruta de siempre y lo reusamos de ahí en
  // adelante. Si ya existe (instalación manual desde el .zip, o una llamada
  // anterior), no lo tocamos: en Windows no se puede sobrescribir un .exe en
  // uso.
  ensureHelper(): string | null {
    const path = this.helperPath();
    if (existsSync(path)) return path;

    const os = GO_OS[process.platform] ?? process.platform;
    const arch = GO_ARCH[process.arch] ?? process.arch;
    const base64 = HELPER_BINARIES[`${os}-${arch}`];
    if (!base64) return null;

    mkdirSync(dirname(path), { recursive: true });
    writeFileSync(path, Buffer.from(base64, "base64"));
    if (process.platform !== "win32") chmodSync(path, 0o755);
    return path;
  }
}
