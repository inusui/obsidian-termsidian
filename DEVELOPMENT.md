# Desarrollo

Notas de mantenimiento para mí mismo: cómo compilar, probar y publicar System Terminal. No es una invitación a contribuir — ver la sección [Contribuir](README.es.md#contribuir) del README: no se aceptan pull requests.

## Requisitos

- [Go](https://go.dev/) 1.27 o superior: `winget install GoLang.Go`
- [Node.js](https://nodejs.org/) 24 LTS, que incluye npm: `winget install OpenJS.NodeJS.LTS`

Después de instalarlos, abre una terminal nueva. Si no, ve a [Problemas frecuentes al compilar](#problemas-frecuentes-al-compilar).

Todos los comandos de este documento son para Git Bash en Windows, o para la terminal de macOS o Linux.

## Compilar

### Instalar tu propio build en un vault

1. Compila el helper y el plugin, como se explica abajo. El resultado queda en `plugin/build/`.
2. Copia el contenido de `plugin/build/` a `<tu-vault>/.obsidian/plugins/termsidian/`. Desde la raíz del repositorio, en Git Bash (o en la terminal de macOS o Linux):

   ```bash
   dest="<tu-vault>/.obsidian/plugins/termsidian"
   mkdir -p "$dest"
   cp -r plugin/build/. "$dest/"
   ```

3. En Obsidian, ve a **Ajustes → Complementos de la comunidad**, pulsa el botón de recargar junto a "Complementos instalados" y activa **System Terminal**.

A diferencia de instalar desde una Release, `main.js` solo lleva dentro el binario de tu propio sistema (no los 6), porque solo compilaste ese. Para un vault sincronizado entre varios sistemas, instala desde una Release en su lugar.

### 1. El helper (Go)

```bash
cd helper
bash build.sh 0.1.0
```

El número es la versión que queda dentro del binario. Si no pones ninguna, vale `dev`.

El script deja en `helper/dist/`:

| Archivo | Para |
|---|---|
| `termsidian-pty-windows-amd64.exe` | Windows en procesadores Intel/AMD |
| `termsidian-pty-windows-arm64.exe` | Windows en procesadores ARM |
| `checksums.txt` | Hash SHA-256 de cada binario |

El script compila con estas opciones:

- `CGO_ENABLED=0`: binario estático, que no depende de nada instalado.
- `-trimpath`: el binario no guarda rutas de tu disco. Además, el mismo código genera siempre el mismo binario, con el mismo hash.
- `-ldflags "-s -w"`: quita la información de depuración, lo que reduce el tamaño a unos 3 MB.
- `-ldflags "-X main.version=…"`: incrusta la versión en el binario.

Para comprobar la versión compilada y los hashes:

```bash
./dist/termsidian-pty-windows-amd64.exe --version
```

```bash
(cd dist && sha256sum -c checksums.txt)
```

#### Otros sistemas

Por defecto, el script compila solo para Windows. Para otros sistemas, indícalos en la variable `TARGETS`, con la forma `sistema/arquitectura` y separados por espacios:

```bash
TARGETS="darwin/arm64 darwin/amd64 linux/amd64 linux/arm64" bash build.sh 0.1.0
```

El script borra `dist/` antes de compilar. Si quieres todos los sistemas a la vez, ponlos todos en `TARGETS` (incluidos `windows/amd64 windows/arm64`).

El plugin busca un binario llamado `termsidian-pty-<sistema>-<arquitectura>`, con la extensión `.exe` en Windows:

- `<sistema>`: `windows`, `darwin` (macOS) o `linux`.
- `<arquitectura>`: `amd64` (Intel/AMD) o `arm64` (Apple Silicon y otros procesadores ARM).

Al compilar el plugin, estos binarios se embeben en base64 dentro de `main.js` (ver `generate-helper-binaries.mjs` y `ensureHelper` en `plugin/src/main.ts`). La primera vez que se abre una terminal, el plugin escribe a disco el binario de tu sistema en `bin/` y lo reusa de ahí en adelante. Así funciona igual si lo instalaste a mano desde el `.zip` o si lo instalaste desde el listado de Community Plugins de Obsidian, que solo descarga `main.js`, `manifest.json` y `styles.css`.

### 2. El plugin (TypeScript)

```bash
cd plugin
npm install
npm run build
```

`npm run build` hace tres cosas:

1. Embebe en base64 los binarios de `helper/dist/` en `src/helper-binaries.generated.ts` (`generate-helper-binaries.mjs`). Por eso hay que compilar el helper primero.
2. Revisa los tipos con `tsc`.
3. Empaqueta el código con esbuild y crea `plugin/build/` con `main.js` (con los binarios ya dentro), `manifest.json`, `styles.css` y las licencias (`LICENSE` y `THIRD_PARTY_NOTICES.md`).

Con `npm run dev`, el plugin se recompila cada vez que guardas un archivo. Después hay que copiarlo al vault igualmente.

## Tests

El helper, desde `helper/`:

```bash
go test ./...
```

Estos tests comprueban el protocolo y lanzan shells reales para verificar que:

- los programas detectan una terminal de verdad;
- el shell recibe los cambios de tamaño;
- el shell muere si se cierra el plugin o si el helper muere de golpe.

Un test más, `TestFlatpakHostShell`, solo se ejecuta dentro de un sandbox de Flatpak: comprueba que el shell es el del sistema real, que Ctrl+C funciona y que el shell muere al cerrar el plugin. Para ejecutarlo con el sandbox de Obsidian (el binario tiene que estar dentro de `$HOME`, porque el sandbox no ve `/tmp`):

```bash
CGO_ENABLED=0 go test -c -o ~/helper.test
flatpak run --talk-name=org.freedesktop.Flatpak --command=$HOME/helper.test md.obsidian.Obsidian -test.v
```

Sin `--talk-name`, el mismo test comprueba que aparece el aviso del permiso.

El plugin, desde `plugin/`:

```bash
npm test
```

Estos tests comprueban que los mensajes se reconstruyen aunque lleguen partidos en trozos, y que el formato es idéntico byte a byte al de Go. Uno de ellos habla con el binario de Go de verdad; se omite si `helper/dist/` no está compilado.

## Integración continua y releases

GitHub Actions ([.github/workflows/release.yml](.github/workflows/release.yml)) hace cuatro cosas:

- **En cada push, a cualquier rama,** ejecuta los tests del helper y del plugin en Windows, macOS y Linux. El resultado se ve en la pestaña **Actions** del repositorio y en la insignia del principio del README.
- **También en cada push,** comprueba que las dependencias son las admitidas (ver [Dependencias](#dependencias)) y que ninguna tiene vulnerabilidades conocidas, con `npm audit` (ver [Vulnerabilidades](#vulnerabilidades)).
- **Cada lunes a las 8:00 UTC,** repite lo anterior sobre `main` aunque nadie haya hecho push, para enterarte de vulnerabilidades publicadas después. Esta ejecución nunca publica una versión.
- **Cuando llegan commits a `main`** (por ejemplo, al hacer merge de `Dev`), y si los tests pasan en los tres sistemas y las dependencias están admitidas, publica una versión nueva sin que hagas nada más: calcula el número, lo guarda en `manifest.json`, crea la etiqueta y publica una Release con:

| Archivo | Qué es |
|---|---|
| `termsidian-<versión>.zip` | La carpeta del plugin completa, lista para descomprimir en `.obsidian/plugins/` |
| `main.js`, `manifest.json`, `styles.css` | El plugin — lo único que descarga Obsidian al instalar desde Community Plugins. El helper para los 6 sistemas va embebido dentro de `main.js` |
| `LICENSE`, `THIRD_PARTY_NOTICES.md` | La licencia de System Terminal y las de las dependencias que van dentro |

Además, `main.js` y `styles.css` llevan una [attestation de GitHub](https://docs.github.com/en/actions/security-guides/using-artifact-attestations-to-establish-provenance-for-builds) que certifica que salieron de este repositorio y este workflow (`gh attestation verify`).

Los binarios de una Release los compila siempre GitHub, nunca un equipo personal.

### Cómo se decide la versión

El número sale de los mensajes de commit (formato [Conventional Commits](https://www.conventionalcommits.org/es/)) que hay desde la última versión:

| Commits desde la última versión | Versión nueva |
|---|---|
| Alguno con `!` tras el tipo (`feat!:`) o con `BREAKING CHANGE:` en el mensaje | major: `0.2.3` → `1.0.0` |
| Alguno `feat:` | minor: `0.2.3` → `0.3.0` |
| Alguno `fix:` o `perf:` | patch: `0.2.3` → `0.2.4` |
| Solo `docs:`, `chore:`, `ci:`, `test:`, `refactor:`… | no se publica versión |

Casos especiales:

- **La primera vez,** sin ninguna etiqueta, se publica la versión que ya tiene `manifest.json`.
- **Para elegir un número concreto,** ponlo tú en `manifest.json`, por ejemplo `1.0.0`. Si es mayor que la última versión publicada, se usa ese.
- **Si haces squash merge,** solo cuenta el mensaje del commit resultante. Ponle un título con el formato correcto, por ejemplo `feat: …`.

Para ver qué versión saldría con los commits actuales, desde la raíz del repositorio:

```bash
bash .github/scripts/next-version.sh
```

### Publicar una versión

1. `main` solo acepta cambios por pull request: sube tu rama, abre un PR de `Dev` a `main` en GitHub y haz merge.

2. Sigue el progreso en la pestaña **Actions**. Al terminar, la Release aparece en la página de Releases.

3. El pipeline sube a `main` un commit `chore(release): <versión>` con el `manifest.json` actualizado. Tráelo a tu rama de trabajo para que las dos queden iguales:

   ```bash
   git switch main && git pull && git switch Dev && git merge main
   ```

### Dependencias

La Release lleva dentro código de otros proyectos: xterm.js en `main.js` y `styles.css`, y la biblioteca estándar de Go y varios módulos de Go en los binarios. Sus licencias piden incluir su aviso de copyright, que está en [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).

Ese archivo es también la lista de dependencias admitidas: cada sección `## nombre` es una. En cada push, el pipeline compara esa lista con las dependencias reales y falla si no coinciden:

- **Dependencias nuevas que no están en THIRD_PARTY_NOTICES.md:** añadiste una dependencia (o llegó como dependencia de otra). Si la quieres, añade su sección con el texto de su archivo `LICENSE`. Si no, quítala.
- **Dependencias que ya no se usan:** borra su sección.

Las dependencias de desarrollo (`devDependencies` de npm, como TypeScript o esbuild) no cuentan: no van dentro de la Release.

Para comprobarlo antes de subir, desde la raíz del repositorio (necesita haber hecho `npm install` en `plugin/`):

```bash
bash .github/scripts/check-deps.sh
```

### Vulnerabilidades

GitHub avisa cuando una dependencia tiene una vulnerabilidad conocida: es **Dependabot alerts**, activado en **Settings → Code security**. Solo avisa: no abre pull requests ni cambia nada en el repositorio. Como `package-lock.json` no se sube, Dependabot solo ve las dependencias npm que están en `package.json`, no las que llegan dentro de otras.

Esas las revisa `npm audit` en el pipeline, en cada push y cada lunes. Si encuentra una, el job **Dependencias admitidas** falla y GitHub te avisa por correo. El log del paso **Vulnerabilidades (npm audit)** dice qué paquete es, de dónde llega y a qué versión hay que subir. Para comprobarlo en local, desde `plugin/`, ejecuta `npm audit`.

La actualización se hace a mano:

1. Abre la alerta en la pestaña **Security → Dependabot** del repositorio, o el log de `npm audit`. Dice qué dependencia es, qué gravedad tiene y a qué versión hay que subir.
2. En tu rama de trabajo, sube la dependencia a esa versión (los números son de ejemplo):

   ```bash
   # Un módulo de Go
   cd helper && go get golang.org/x/crypto@v0.52.0 && go mod tidy
   ```

   ```bash
   # Un paquete npm que está en package.json
   cd plugin && npm install @xterm/xterm@6.0.1
   ```

   Un paquete npm que llega como dependencia de otro se fuerza con `overrides` en `plugin/package.json`, con la versión exacta, y después `npm install`:

   ```json
   "overrides": {
     "moment": "2.31.0"
   }
   ```

   Si es una acción del workflow, cambia su versión en `release.yml`.

3. Pasa los tests, `bash .github/scripts/check-deps.sh` y `npm audit`.
4. Haz commit. El tipo decide si se publica una versión al llegar a `main`:

| Qué actualizaste | Commit | Versión nueva |
|---|---|---|
| Algo que va dentro de la Release (módulos de Go, xterm.js) | `fix(deps): …` | patch, con el arreglo |
| Una herramienta de compilación (TypeScript, esbuild…) | `chore(deps): …` | ninguna |
| Una acción del workflow | `ci(deps): …` | ninguna |

5. Haz merge en `main` como siempre. La alerta de Dependabot se cierra sola cuando el arreglo llega a `main`.

## Problemas frecuentes al compilar

**`npm: command not found` o `go: command not found`**

La terminal se abrió antes de instalar el programa y no sabe que existe. Cierra Git Bash y ábrelo de nuevo. Si la terminal está dentro de otra aplicación (VS Code, por ejemplo), reinicia esa aplicación. Para comprobar que ya los encuentra:

```bash
go version && node --version && npm --version
```

**`$'\r': command not found` al ejecutar `build.sh`**

El script tiene saltos de línea de Windows (CRLF). El archivo `.gitattributes` del repositorio obliga a Git a guardar los `.sh` con saltos LF. Si aun así pasa, conviértelo:

```bash
sed -i 's/\r$//' helper/build.sh
```

**Si usas PowerShell en lugar de Git Bash**

Si ves *"No se puede cargar el archivo …\npm.ps1 porque la ejecución de scripts está deshabilitada"*, es la política de scripts de PowerShell. Usa `npm.cmd` en su lugar (por ejemplo `npm.cmd run build`), o ejecuta los comandos en Git Bash, donde no pasa.

**`npm error enoent Could not read package.json`**

Ejecuta npm dentro de `plugin/`, no en la raíz del repositorio.

## Cómo funciona

```
Obsidian
 └─ Plugin (TypeScript)
      ├─ main.ts      registra el panel, el icono y el comando
      ├─ view.ts      el panel, con xterm.js
      └─ session.ts   lanza el helper y habla con él
             │  stdin/stdout: mensajes [tipo][longitud][datos]
             ▼
   termsidian-pty (Go)
      └─ ConPTY (Windows) / PTY (macOS, Linux) ── tu shell
```

El plugin y el helper se comunican con mensajes binarios: 1 byte de tipo, 4 bytes de longitud (big-endian) y los datos.

| Tipo | Dirección | Contenido |
|---|---|---|
| `HELLO` | helper → plugin | versión del protocolo y del helper |
| `DATA` | en los dos sentidos | bytes de la terminal (UTF-8) |
| `RESIZE` | plugin → helper | columnas y filas (uint16 cada uno) |
| `EXIT` | helper → plugin | código de salida del shell (int32) |

La salida de error (stderr) del helper se usa solo para mensajes de depuración. Si el plugin y el helper no tienen la misma versión de protocolo, el plugin lo avisa en lugar de conectarse.

Si xterm.js no da abasto con una salida muy grande, el plugin deja de leer al helper. La tubería se llena, y el helper y el shell esperan solos hasta que la terminal se pone al día.

**¿Por qué un helper en Go?**

- `node-pty` hay que recompilarlo para cada versión de Electron, y se rompe cuando Obsidian se actualiza.
- Un helper en Python necesita Python instalado, y el módulo `pty` no existe en Windows.
- Go genera un único binario sin dependencias para cada sistema.

## Estructura del repositorio

```
obsidian-termsidian/
 ├─ manifest.json          manifest del plugin (id, versión, versión mínima de Obsidian)
 ├─ LICENSE                licencia MIT
 ├─ THIRD_PARTY_NOTICES.md licencias de las dependencias, y lista de las admitidas
 ├─ SECURITY.md            cómo reportar una vulnerabilidad
 ├─ .gitattributes         guarda los .sh con saltos de línea LF
 ├─ .github/
 │   ├─ workflows/release.yml    tests en cada push y release al llegar a main
 │   ├─ scripts/next-version.sh  calcula la versión a partir de los commits
 │   └─ scripts/check-deps.sh    compara las dependencias con THIRD_PARTY_NOTICES.md
 ├─ helper/                termsidian-pty, en Go
 │   ├─ main.go            opciones, PTY y conexión con el plugin
 │   ├─ flatpak*.go        lanzar el shell fuera del sandbox de Flatpak
 │   ├─ protocol/          formato de los mensajes, con sus tests
 │   ├─ build.sh           compila los binarios en helper/dist/
 │   └─ go.mod, go.sum     dependencias y sus hashes
 └─ plugin/                el plugin, en TypeScript
     ├─ src/                      main.ts, view.ts, session.ts, protocol.ts, styles.css y tests
     ├─ generate-helper-binaries.mjs  embebe helper/dist/ en src/helper-binaries.generated.ts
     ├─ esbuild.config.mjs        genera plugin/build/
     └─ package.json
```

`helper/dist/`, `plugin/build/`, `plugin/node_modules/`, `plugin/src/helper-binaries.generated.ts` y `plugin/package-lock.json` se generan al compilar o al hacer `npm install` y no se suben al repositorio. `go.sum` sí se sube.
