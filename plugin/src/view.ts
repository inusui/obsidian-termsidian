import { existsSync } from "node:fs";
import { ItemView, type WorkspaceLeaf } from "obsidian";
import { FitAddon } from "@xterm/addon-fit";
import { Terminal, type ITheme } from "@xterm/xterm";
import type TermsidianPlugin from "./main.ts";
import { HelperSession } from "./session.ts";

export const VIEW_TYPE_TERMINAL = "termsidian-terminal";

// Control de flujo: con tantas escrituras pendientes en xterm dejamos de leer
// al helper, y volvemos a leer cuando baja a LOW_WATERMARK.
const HIGH_WATERMARK = 10;
const LOW_WATERMARK = 2;

// TerminalView es el panel de la barra lateral: un xterm.js conectado a una
// HelperSession.
export class TerminalView extends ItemView {
  private readonly plugin: TermsidianPlugin;
  private readonly fitAddon = new FitAddon();
  private term: Terminal | null = null;
  private session: HelperSession | null = null;
  private resizeObserver: ResizeObserver | null = null;
  private pendingWrites = 0;

  constructor(leaf: WorkspaceLeaf, plugin: TermsidianPlugin) {
    super(leaf);
    this.plugin = plugin;
  }

  getViewType(): string {
    return VIEW_TYPE_TERMINAL;
  }

  getDisplayText(): string {
    return "Terminal";
  }

  getIcon(): string {
    return "terminal";
  }

  focus(): void {
    this.term?.focus();
  }

  async onOpen(): Promise<void> {
    const el = this.contentEl;
    el.empty();
    el.addClass("termsidian-view");

    const term = new Terminal({
      cursorBlink: true,
      fontFamily: monospaceFont(),
      fontSize: 13,
      scrollback: 5000,
      theme: this.theme(),
    });
    this.term = term;
    term.loadAddon(this.fitAddon);
    term.open(el);
    this.fit();

    term.attachCustomKeyEventHandler((event) => this.handleKey(event));
    term.onData((data) => {
      if (this.session) this.session.write(data);
      else if (data === "\r") this.startSession(); // Enter tras salir: shell nuevo
    });
    // Cuando fit() cambia columnas o filas, se lo contamos al shell (RF-04).
    term.onResize(({ cols, rows }) => this.session?.resize(cols, rows));

    this.resizeObserver = new ResizeObserver(() => this.fit());
    this.resizeObserver.observe(el);
    // Si cambias de tema en Obsidian, la terminal lo sigue.
    this.registerEvent(
      this.app.workspace.on("css-change", () => {
        if (this.term) this.term.options.theme = this.theme();
      }),
    );

    this.startSession();
  }

  async onClose(): Promise<void> {
    this.resizeObserver?.disconnect();
    this.session?.close();
    this.session = null;
    this.term?.dispose();
    this.term = null;
  }

  private startSession(): void {
    const term = this.term;
    if (!term) return;

    const helperPath = this.plugin.helperPath();
    if (!existsSync(helperPath)) {
      this.printError(`No encuentro el helper en ${helperPath}`);
      return;
    }

    this.session = new HelperSession(
      { helperPath, cwd: this.plugin.vaultPath(), cols: term.cols, rows: term.rows },
      {
        onData: (data) => this.writeOutput(data),
        onExit: (code) => {
          this.session = null;
          term.write(`\r\n\x1b[2m[El shell terminó con código ${code}. Pulsa Enter para abrir otro.]\x1b[0m\r\n`);
        },
        onError: (message) => {
          this.session = null;
          this.printError(message);
        },
      },
    );
  }

  private writeOutput(data: Uint8Array): void {
    const term = this.term;
    if (!term) return;
    this.pendingWrites++;
    if (this.pendingWrites >= HIGH_WATERMARK) this.session?.pause();
    // xterm recibe bytes UTF-8 tal cual: si un carácter llega partido entre
    // dos trozos, xterm lo junta.
    term.write(data, () => {
      this.pendingWrites--;
      if (this.pendingWrites <= LOW_WATERMARK) this.session?.resume();
    });
  }

  private printError(message: string): void {
    // \x1b[31m ... \x1b[0m: secuencia ANSI para texto rojo.
    this.term?.write(`\r\n\x1b[31m[Termsidian] ${message}\x1b[0m\r\n`);
  }

  private fit(): void {
    // Con el panel oculto o colapsado mide 0: no hay nada que ajustar.
    if (!this.term || this.contentEl.clientWidth === 0 || this.contentEl.clientHeight === 0) return;
    this.fitAddon.fit();
  }

  // Copiar y pegar como en Windows Terminal: Ctrl+C copia si hay texto
  // seleccionado (si no, es la interrupción de siempre) y Ctrl+V pega.
  private handleKey(event: KeyboardEvent): boolean {
    if (event.type !== "keydown" || !event.ctrlKey || event.altKey || event.metaKey) return true;
    const key = event.key.toLowerCase();
    if (key === "c" && this.term?.hasSelection()) {
      void navigator.clipboard.writeText(this.term.getSelection());
      this.term.clearSelection();
      return false;
    }
    // false = xterm ignora la tecla; el navegador pega el portapapeles y
    // xterm lo recibe como texto pegado.
    if (key === "v") return false;
    return true;
  }

  private theme(): ITheme {
    return {
      background: backgroundOf(this.contentEl),
      foreground: cssVar("--text-normal"),
      cursor: cssVar("--text-accent"),
      selectionBackground: cssVar("--text-selection"),
    };
  }
}

function cssVar(name: string): string {
  return getComputedStyle(document.body).getPropertyValue(name).trim();
}

// La fuente monoespaciada de Obsidian. La variable puede traer huecos vacíos
// (", , Menlo"), que xterm no entiende.
function monospaceFont(): string {
  const fonts = cssVar("--font-monospace")
    .split(",")
    .map((font) => font.trim())
    .filter(Boolean);
  return fonts.length > 0 ? fonts.join(", ") : "monospace";
}

// El fondo real del panel: se sube por los elementos padre hasta encontrar
// uno que no sea transparente. Así coincide en la barra lateral y fuera de ella.
function backgroundOf(el: HTMLElement): string {
  for (let node: HTMLElement | null = el; node; node = node.parentElement) {
    const color = getComputedStyle(node).backgroundColor;
    if (color && color !== "transparent" && color !== "rgba(0, 0, 0, 0)") return color;
  }
  return cssVar("--background-primary");
}
