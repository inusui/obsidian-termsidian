// Genera src/helper-binaries.generated.ts: los binarios de helper/dist, en
// base64, embebidos en un objeto. main.js los lleva todos dentro (para los
// 6 sistemas) y el plugin escribe a disco solo el que toca en cada equipo,
// la primera vez que se abre una terminal (ver ensureHelper en main.ts).
//
// Así el plugin funciona instalado desde el listado de Community Plugins de
// Obsidian, que solo descarga main.js, manifest.json y styles.css: el .zip
// de cada Release ya no necesita llevar bin/ aparte.
//
// Uso: node generate-helper-binaries.mjs

import { existsSync, readFileSync, readdirSync, writeFileSync } from "node:fs";

const helperDist = "../helper/dist";
const outFile = "src/helper-binaries.generated.ts";

const entries = [];
if (existsSync(helperDist)) {
  for (const file of readdirSync(helperDist)) {
    if (!file.startsWith("termsidian-pty-")) continue;
    const key = file.replace(/^termsidian-pty-/, "").replace(/\.exe$/, "");
    const base64 = readFileSync(`${helperDist}/${file}`).toString("base64");
    entries.push([key, base64]);
  }
} else {
  console.warn("aviso: no existe helper/dist; compila el helper con helper/build.sh");
}

const body = entries.map(([key, base64]) => `  "${key}": "${base64}",`).join("\n");
writeFileSync(
  outFile,
  `// Generado por generate-helper-binaries.mjs a partir de helper/dist. No se versiona.\n` +
    `export const HELPER_BINARIES: Readonly<Record<string, string>> = {\n${body}\n};\n`,
);
console.log(`listo: ${outFile} (${entries.length} binarios)`);
