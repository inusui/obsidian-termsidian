import assert from "node:assert/strict";
import { existsSync } from "node:fs";
import { test } from "node:test";
import { fileURLToPath } from "node:url";
import { HelperSession } from "./session.ts";

// El helper que genera helper/build.sh.
const helperPath = fileURLToPath(
  new URL("../../helper/dist/termsidian-pty-windows-amd64.exe", import.meta.url),
);
const skip =
  process.platform !== "win32" || !existsSync(helperPath)
    ? "requiere Windows y el helper compilado con helper/build.sh"
    : false;

// Prueba de extremo a extremo: TypeScript habla con el binario de Go. Si
// alguno de los dos lados cambia el protocolo, este test falla.
test("habla con el helper de Go", { skip, timeout: 15_000 }, async () => {
  const decoder = new TextDecoder();
  let screen = "";
  let sentExit = false;

  const code = await new Promise<number>((resolve, reject) => {
    const session = new HelperSession(
      { helperPath, cwd: process.cwd(), cols: 80, rows: 24, shell: "cmd.exe" },
      {
        onData: (data) => {
          screen += decoder.decode(data, { stream: true });
          // tty-Windows_NT solo aparece si el shell expandió %OS%.
          if (!sentExit && screen.includes("tty-Windows_NT")) {
            sentExit = true;
            session.write("exit 3\r");
          }
        },
        onExit: resolve,
        onError: (message) => reject(new Error(message)),
      },
    );
    session.write("echo tty-%OS%\r");
  });

  assert.equal(code, 3);
});
