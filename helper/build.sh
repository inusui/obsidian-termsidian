#!/usr/bin/env bash
# Compila el helper en dist/ y genera checksums.txt. Es la versión local de lo
# que hará GitHub Actions al publicar una release.
#
# Uso:  bash build.sh [versión]                  por ejemplo: bash build.sh 0.1.0
#       TARGETS="darwin/arm64 linux/amd64" bash build.sh 0.1.0

# Para en el primer error (-e), si falta una variable (-u) o si falla algo
# dentro de una tubería (pipefail).
set -euo pipefail

version="${1:-dev}"
# Pares sistema/arquitectura. Por defecto, solo Windows.
targets="${TARGETS:-windows/amd64 windows/arm64}"

# Trabajamos en la carpeta del script, se llame desde donde se llame.
cd "$(dirname "$0")"
rm -rf dist
mkdir -p dist

# -s -w  quita la tabla de símbolos y la información de depuración: el
#        binario ocupa bastante menos.
# -X     cambia el valor de una variable string al compilar: main.version.
ldflags="-s -w -X main.version=$version"

for target in $targets; do
  os="${target%/*}"   # lo que va antes de la barra
  arch="${target#*/}" # lo que va después
  name="termsidian-pty-$os-$arch"
  if [ "$os" = windows ]; then
    name="$name.exe"
  fi

  # CGO_ENABLED=0: sin C, binario estático que no depende de nada instalado.
  # GOOS/GOARCH: para qué sistema se compila (compilación cruzada).
  # Puestas delante del comando, solo valen para ese comando.
  # -trimpath: no guarda rutas de tu disco (C:\Users\...) en el binario.
  CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" \
    go build -trimpath -ldflags "$ldflags" -o "dist/$name" .
  echo "ok  $name"
done

# checksums.txt con el formato de sha256sum, "<hash>  <archivo>" (RF-13).
# macOS no trae sha256sum, pero sí shasum. En Windows, sha256sum marca los
# archivos binarios con "*" delante del nombre; sed lo quita.
cd dist
if command -v sha256sum >/dev/null; then
  sha256sum termsidian-pty-* | sed 's/ \*/  /' >checksums.txt
else
  shasum -a 256 termsidian-pty-* >checksums.txt
fi
echo "listo: $(pwd)"
