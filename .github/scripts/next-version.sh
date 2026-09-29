#!/usr/bin/env bash
# Calcula la próxima versión a partir de los commits desde la última etiqueta,
# siguiendo Conventional Commits:
#
#   tipo!: o "BREAKING CHANGE"  → major  (1.2.3 → 2.0.0)
#   feat:                       → minor  (1.2.3 → 1.3.0)
#   fix: o perf:                → patch  (1.2.3 → 1.2.4)
#   docs:, chore:, ci:, ...     → no hay versión nueva
#
# Casos especiales:
#   - Sin etiquetas todavía: la primera versión es la de manifest.json.
#   - Si en manifest.json hay a mano una versión mayor que la última etiqueta,
#     se usa esa (para elegir un número concreto).
#
# Imprime la versión nueva, o nada si no hay que publicar.
# Uso (desde la raíz del repositorio): bash .github/scripts/next-version.sh

set -euo pipefail

current="$(node -p 'require("./manifest.json").version')"
last_tag="$(git describe --tags --abbrev=0 2>/dev/null || true)"

if [ -z "$last_tag" ]; then
  echo "$current"
  exit 0
fi

# sort -V ordena por versión (0.10.0 va después de 0.9.0).
highest="$(printf '%s\n%s\n' "$last_tag" "$current" | sort -V | tail -1)"
if [ "$current" != "$last_tag" ] && [ "$highest" = "$current" ]; then
  echo "$current"
  exit 0
fi

# %s es la primera línea de cada commit; %b, el resto del mensaje.
subjects="$(git log "$last_tag..HEAD" --format=%s)"
bodies="$(git log "$last_tag..HEAD" --format=%b)"

if grep -qE '^[a-z]+(\([^)]*\))?!:' <<<"$subjects" || grep -q 'BREAKING CHANGE' <<<"$bodies"; then
  bump=major
elif grep -qE '^feat(\([^)]*\))?:' <<<"$subjects"; then
  bump=minor
elif grep -qE '^(fix|perf)(\([^)]*\))?:' <<<"$subjects"; then
  bump=patch
else
  exit 0 # solo docs, chore, ci...: nada que publicar
fi

IFS=. read -r major minor patch <<<"$last_tag"
case "$bump" in
  major) echo "$((major + 1)).0.0" ;;
  minor) echo "$major.$((minor + 1)).0" ;;
  patch) echo "$major.$minor.$((patch + 1))" ;;
esac
