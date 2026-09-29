// Protocolo plugin ↔ helper: el mismo que helper/protocol/protocol.go.
// Cada mensaje es [1 byte tipo][4 bytes longitud big-endian][payload].

export const PROTOCOL_VERSION = 1;

// Tipos de mensaje. Un objeto "as const" en lugar de un enum: Node puede
// ejecutar este archivo quitándole los tipos, y los enum no se pueden quitar.
export const Kind = {
  Hello: 1,
  Data: 2,
  Resize: 3,
  Exit: 4,
} as const;
export type Kind = (typeof Kind)[keyof typeof Kind];

const HEADER_SIZE = 5;
const MAX_PAYLOAD = 1 << 20; // 1 MiB, igual que en Go

export interface Frame {
  kind: number;
  payload: Uint8Array;
}

export function encodeFrame(kind: Kind, payload: Uint8Array): Uint8Array {
  const buf = new Uint8Array(HEADER_SIZE + payload.length);
  buf[0] = kind;
  // DataView escribe en big-endian por defecto, como binary.BigEndian en Go.
  new DataView(buf.buffer).setUint32(1, payload.length);
  buf.set(payload, HEADER_SIZE);
  return buf;
}

export function encodeResize(cols: number, rows: number): Uint8Array {
  const payload = new Uint8Array(4);
  const view = new DataView(payload.buffer);
  view.setUint16(0, cols);
  view.setUint16(2, rows);
  return encodeFrame(Kind.Resize, payload);
}

export function decodeHello(payload: Uint8Array): { protocol: number; version: string } {
  return {
    protocol: payload[0] ?? 0,
    version: new TextDecoder().decode(payload.subarray(1)),
  };
}

export function decodeExit(payload: Uint8Array): number {
  return dataView(payload).getInt32(0);
}

// FrameParser junta los trozos que llegan por stdout y devuelve los mensajes
// completos. Es el equivalente de io.ReadFull en el helper: un trozo puede
// traer medio mensaje, o tres mensajes y medio.
export class FrameParser {
  private pending: Uint8Array = new Uint8Array(0);

  push(chunk: Uint8Array): Frame[] {
    this.pending = concat(this.pending, chunk);
    const frames: Frame[] = [];
    while (this.pending.length >= HEADER_SIZE) {
      const size = dataView(this.pending).getUint32(1);
      if (size > MAX_PAYLOAD) {
        throw new Error(`mensaje demasiado grande: ${size} bytes`);
      }
      if (this.pending.length < HEADER_SIZE + size) {
        break; // falta el resto; llegará en el próximo trozo
      }
      frames.push({
        kind: this.pending[0]!,
        payload: this.pending.subarray(HEADER_SIZE, HEADER_SIZE + size),
      });
      this.pending = this.pending.subarray(HEADER_SIZE + size);
    }
    return frames;
  }
}

// Los Buffer de Node suelen ser vistas dentro de un bloque de memoria más
// grande: la DataView tiene que empezar donde empieza la vista, no en 0.
function dataView(bytes: Uint8Array): DataView {
  return new DataView(bytes.buffer, bytes.byteOffset, bytes.byteLength);
}

function concat(a: Uint8Array, b: Uint8Array): Uint8Array {
  if (a.length === 0) return b;
  const out = new Uint8Array(a.length + b.length);
  out.set(a);
  out.set(b, a.length);
  return out;
}
