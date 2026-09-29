import { spawn, type ChildProcessWithoutNullStreams } from "node:child_process";
import {
  decodeExit,
  decodeHello,
  encodeFrame,
  encodeResize,
  FrameParser,
  Kind,
  PROTOCOL_VERSION,
} from "./protocol.ts";

export interface SessionOptions {
  helperPath: string;
  cwd: string;
  cols: number;
  rows: number;
  shell?: string;
  args?: string[];
}

export interface SessionEvents {
  onData(data: Uint8Array): void;
  onExit(code: number): void;
  onError(message: string): void;
}

// Si al cerrar la sesión el helper no sale solo en este tiempo, se mata.
const KILL_TIMEOUT_MS = 2000;

// HelperSession lanza termsidian-pty y traduce entre mensajes del protocolo
// y eventos. No sabe nada de Obsidian ni de xterm.js, así que los tests
// pueden usarla con el helper de verdad.
export class HelperSession {
  private readonly proc: ChildProcessWithoutNullStreams;
  private readonly events: SessionEvents;
  private readonly parser = new FrameParser();
  private readonly encoder = new TextEncoder();
  // true cuando ya no hay que avisar de nada: llegó EXIT, hubo un error o
  // cerramos nosotros.
  private finished = false;

  constructor(opts: SessionOptions, events: SessionEvents) {
    this.events = events;

    const args = ["--cwd", opts.cwd, "--cols", String(opts.cols), "--rows", String(opts.rows)];
    if (opts.shell) args.push("--shell", opts.shell);
    if (opts.args?.length) args.push("--", ...opts.args);

    // windowsHide: sin esto, en Windows aparecería una ventana de consola.
    this.proc = spawn(opts.helperPath, args, { windowsHide: true });

    this.proc.stdout.on("data", (chunk: Buffer) => this.onChunk(chunk));
    // stderr del helper son logs de depuración (consola de Obsidian: Ctrl+Shift+I).
    this.proc.stderr.on("data", (chunk: Buffer) => {
      console.debug("[termsidian-pty]", chunk.toString().trimEnd());
    });
    // Escribir a un helper que ya salió da error; de eso avisa "close".
    this.proc.stdin.on("error", () => {});
    this.proc.on("error", (err) => this.fail(`no se pudo lanzar el helper: ${err.message}`));
    // "close" llega cuando el proceso terminó y ya se leyó toda su salida.
    this.proc.on("close", (code) => this.fail(`el helper terminó sin avisar (código ${code})`));
  }

  write(text: string): void {
    this.send(encodeFrame(Kind.Data, this.encoder.encode(text)));
  }

  resize(cols: number, rows: number): void {
    if (cols > 0 && rows > 0) this.send(encodeResize(cols, rows));
  }

  // pause/resume dejan de leer la salida del helper y vuelven a leerla. Si
  // no leemos, la tubería se llena y el helper y el shell esperan solos.
  pause(): void {
    this.proc.stdout.pause();
  }

  resume(): void {
    this.proc.stdout.resume();
  }

  // close cierra la entrada del helper: él mata al shell y sale (RF-06).
  close(): void {
    this.finished = true;
    this.proc.stdin.end();
    const timer = setTimeout(() => this.proc.kill(), KILL_TIMEOUT_MS);
    this.proc.once("close", () => clearTimeout(timer));
  }

  private send(frame: Uint8Array): void {
    if (!this.finished && this.proc.stdin.writable) this.proc.stdin.write(frame);
  }

  private onChunk(chunk: Uint8Array): void {
    if (this.finished) return;
    let frames;
    try {
      frames = this.parser.push(chunk);
    } catch (err) {
      this.fail(String(err));
      this.proc.kill();
      return;
    }
    for (const frame of frames) {
      switch (frame.kind) {
        case Kind.Hello: {
          const hello = decodeHello(frame.payload);
          if (hello.protocol !== PROTOCOL_VERSION) {
            this.fail(
              `el helper ${hello.version} usa el protocolo ${hello.protocol} y este plugin el ${PROTOCOL_VERSION}`,
            );
            this.proc.kill();
            return;
          }
          break;
        }
        case Kind.Data:
          this.events.onData(frame.payload);
          break;
        case Kind.Exit:
          this.finished = true;
          this.events.onExit(decodeExit(frame.payload));
          return;
      }
    }
  }

  private fail(message: string): void {
    if (this.finished) return;
    this.finished = true;
    this.events.onError(message);
  }
}
