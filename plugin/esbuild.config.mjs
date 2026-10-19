// Genera build/, la carpeta que se copia al vault:
//   main.js        (src/main.ts y todo lo que importa, en un solo archivo;
//                  lleva dentro los 6 binarios del helper, en base64 — ver
//                  generate-helper-binaries.mjs y ensureHelper en main.ts)
//   styles.css     (src/styles.css + el CSS de xterm)
//   manifest.json  (el de la raíz del repositorio)
//   LICENSE, THIRD_PARTY_NOTICES.md  (licencias: la nuestra y las de lo que
//                  va dentro de main.js y styles.css)
//
// Uso: node esbuild.config.mjs [--watch]

import * as esbuild from "esbuild";
import { cpSync, mkdirSync, rmSync } from "node:fs";

const watch = process.argv.includes("--watch");
const outdir = "build";

rmSync(outdir, { recursive: true, force: true });
mkdirSync(outdir, { recursive: true });
for (const file of ["manifest.json", "LICENSE", "THIRD_PARTY_NOTICES.md"]) {
  cpSync(`../${file}`, `${outdir}/${file}`);
}

const common = { bundle: true, logLevel: "info", minify: !watch };

const contexts = await Promise.all([
  esbuild.context({
    ...common,
    entryPoints: ["src/main.ts"],
    outfile: `${outdir}/main.js`,
    // Obsidian carga los plugins como módulos CommonJS.
    format: "cjs",
    // platform node: los módulos de Node (child_process, path...) no se
    // empaquetan; Obsidian (Electron) ya los trae.
    platform: "node",
    target: "es2022",
    // Los aporta Obsidian en tiempo de ejecución.
    external: ["obsidian", "electron"],
    sourcemap: watch ? "inline" : false,
  }),
  esbuild.context({
    ...common,
    entryPoints: ["src/styles.css"],
    outfile: `${outdir}/styles.css`,
  }),
]);

if (watch) {
  await Promise.all(contexts.map((ctx) => ctx.watch()));
} else {
  await Promise.all(contexts.map((ctx) => ctx.rebuild()));
  await Promise.all(contexts.map((ctx) => ctx.dispose()));
}
