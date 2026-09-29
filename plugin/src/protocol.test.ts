import assert from "node:assert/strict";
import { test } from "node:test";
import { decodeExit, decodeHello, encodeFrame, encodeResize, FrameParser, Kind } from "./protocol.ts";

const utf8 = (text: string) => new TextEncoder().encode(text);

test("reconstruye los mensajes aunque lleguen partidos en trozos de cualquier tamaño", () => {
  const exitPayload = new Uint8Array([0, 0, 0, 3]);
  const bytes = new Uint8Array([
    ...encodeFrame(Kind.Data, utf8("ñandú 🐧")),
    ...encodeFrame(Kind.Data, new Uint8Array(0)),
    ...encodeFrame(Kind.Exit, exitPayload),
  ]);

  // Probamos todos los tamaños de trozo, de 1 byte hasta todo de golpe.
  for (let size = 1; size <= bytes.length; size++) {
    const parser = new FrameParser();
    const frames = [];
    for (let i = 0; i < bytes.length; i += size) {
      frames.push(...parser.push(bytes.subarray(i, i + size)));
    }
    assert.equal(frames.length, 3, `trozos de ${size} bytes`);
    assert.equal(new TextDecoder().decode(frames[0]!.payload), "ñandú 🐧");
    assert.equal(frames[1]!.payload.length, 0);
    assert.equal(decodeExit(frames[2]!.payload), 3);
  }
});

// Estos bytes son los mismos que usan los tests de Go: si un lado cambia el
// formato, falla aquí o allí.
test("usa el mismo formato que el helper de Go", () => {
  assert.deepEqual([...encodeResize(120, 40)], [Kind.Resize, 0, 0, 0, 4, 0, 120, 0, 40]);
  // 0xC000013A: código de Windows al cerrar con Ctrl+C (TestExitPayloadWindowsCode).
  assert.equal(decodeExit(new Uint8Array([0xc0, 0x00, 0x01, 0x3a])), -1073741510);
  assert.deepEqual(decodeHello(new Uint8Array([1, ...utf8("0.1.0")])), { protocol: 1, version: "0.1.0" });
});

test("rechaza un mensaje de más de 1 MiB", () => {
  const header = new Uint8Array(5);
  header[0] = Kind.Data;
  new DataView(header.buffer).setUint32(1, (1 << 20) + 1);
  assert.throws(() => new FrameParser().push(header), /demasiado grande/);
});
