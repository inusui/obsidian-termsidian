#!/usr/bin/env bash
# Comprueba que las dependencias que acaban dentro de la Release son
# exactamente las de THIRD_PARTY_NOTICES.md (sus secciones "## nombre").
# Si se añade o se quita una, falla hasta que se actualice ese archivo: así
# no entra ninguna sin que te des cuenta, y los avisos de copyright no se
# quedan desactualizados.
#
# Cuenta como dependencia:
#   - Go: la biblioteca estándar, que va en todos los binarios del helper.
#   - Los módulos de Go que se compilan dentro del helper, en cualquiera de
#     los 6 sistemas (macOS y Linux usan alguno más que Windows).
#   - Los paquetes npm que no son solo de desarrollo: esbuild los mete dentro
#     de main.js.
#
# Uso: bash .github/scripts/check-deps.sh

set -euo pipefail

# Trabajamos en la raíz del repositorio, se llame desde donde se llame.
cd "$(dirname "$0")/../.."

notices=THIRD_PARTY_NOTICES.md
# Los mismos sistemas que compila release.yml.
targets="${TARGETS:-windows/amd64 windows/arm64 darwin/amd64 darwin/arm64 linux/amd64 linux/arm64}"

# Las listas van a archivos temporales. trap borra la carpeta al salir del
# script, pase lo que pase.
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

# 1. Las dependencias reales, una por línea.
{
  echo Go
  for target in $targets; do
    # go list -deps: todos los paquetes que se compilan con el helper. La
    # plantilla -f imprime el módulo de cada uno, salvo el nuestro (.Main) y
    # los de la biblioteca estándar, que no tienen módulo.
    (cd helper && CGO_ENABLED=0 GOOS="${target%/*}" GOARCH="${target#*/}" \
      go list -deps -f '{{with .Module}}{{if not .Main}}{{.Path}}{{end}}{{end}}' .)
  done
  # package-lock.json marca con "dev": true lo que solo se usa para compilar
  # (typescript, esbuild, obsidian...). Cada clave es una ruta como
  # "node_modules/@xterm/xterm": el nombre es lo que va después del último
  # "node_modules/".
  node -e '
    const lock = require("./plugin/package-lock.json");
    for (const [path, pkg] of Object.entries(lock.packages)) {
      if (path.startsWith("node_modules/") && !pkg.dev) {
        console.log(path.slice(path.lastIndexOf("node_modules/") + "node_modules/".length));
      }
    }
  '
} >"$tmp/actual.raw"

# 2. Las admitidas: los títulos "## " de THIRD_PARTY_NOTICES.md. tr quita el
# \r que Git añade en Windows (core.autocrlf).
tr -d '\r' <"$notices" | sed -n 's/^## //p' >"$tmp/allowed.raw"

# 3. Comparar. comm necesita las dos listas ordenadas igual: LC_ALL=C da el
# mismo orden en todos los sistemas, y -u quita los repetidos (un módulo sale
# una vez por cada sistema). comm -23 deja lo que solo está en la primera
# lista; comm -13, lo que solo está en la segunda.
LC_ALL=C sort -u "$tmp/actual.raw" >"$tmp/actual"
LC_ALL=C sort -u "$tmp/allowed.raw" >"$tmp/allowed"
missing="$(LC_ALL=C comm -23 "$tmp/actual" "$tmp/allowed")"
unused="$(LC_ALL=C comm -13 "$tmp/actual" "$tmp/allowed")"

if [ -n "$missing" ]; then
  echo "Dependencias nuevas que no están en $notices:"
  sed 's/^/  + /' <<<"$missing"
fi
if [ -n "$unused" ]; then
  echo "Dependencias de $notices que ya no se usan:"
  sed 's/^/  - /' <<<"$unused"
fi
if [ -n "$missing$unused" ]; then
  echo
  echo "Añade o quita su sección \"## nombre\" en $notices. Al añadirla, copia"
  echo "el archivo LICENSE de la dependencia: es su aviso de copyright."
  exit 1
fi
echo "ok: las $(wc -l <"$tmp/actual" | tr -d ' ') dependencias están en $notices"
