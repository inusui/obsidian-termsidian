# Termsidian

[![CI y release](https://github.com/inusui/obsidian-termsidian/actions/workflows/release.yml/badge.svg)](https://github.com/inusui/obsidian-termsidian/actions/workflows/release.yml)

Terminal real en la barra lateral derecha de Obsidian, con el shell de tu sistema (PowerShell, cmd, bash, zsh…).

Sin Python y sin node-pty: un programa pequeño escrito en Go, `termsidian-pty`, abre una terminal del sistema (ConPTY en Windows, PTY en macOS y Linux), y el plugin la dibuja con [xterm.js](https://xtermjs.org/).

> **Estado:** v0.1, probado en Windows 11. En macOS y Linux compila, pero todavía no se ha probado.

## Qué hace

- Abre una terminal en la barra lateral derecha, empezando en la carpeta del vault.
- Es una terminal de verdad: colores, autocompletado con Tab, Ctrl+C, historial y programas interactivos.
- Se reajusta al cambiar el tamaño del panel.
- Usa la fuente monoespaciada y los colores de tu tema de Obsidian, y cambia con él.
- Al cerrar el panel no quedan procesos huérfanos.

## Uso

Abre la terminal con el icono de terminal de la barra izquierda, o con `Ctrl+P` → **Termsidian: Abrir terminal**.

El shell es PowerShell en Windows, y en macOS y Linux el de la variable `$SHELL`.

| Tecla | Qué hace |
|---|---|
| `Ctrl+C` | Con texto seleccionado, lo copia. Sin selección, interrumpe el programa, como siempre. |
| `Ctrl+V` | Pega el portapapeles. |
| `Enter` después de `exit` | Abre un shell nuevo. |

Al cerrar el panel se cierran el shell y todo lo que corría dentro de él. Los programas con ventana propia que abriste desde la terminal (por ejemplo `code .`) siguen abiertos.

## Instalación en tu vault

**Requisitos:** Obsidian de escritorio 1.7.2 o superior, con los complementos de la comunidad activados (modo restringido desactivado).

`<tu-vault>` es la carpeta que contiene `.obsidian`. En Git Bash, las rutas de Windows se escriben con barras normales y la unidad en minúscula: `C:\Users\Ana\Notas` se escribe `/c/Users/Ana/Notas`.

### Opción A: desde una Release (sin compilar)

1. En la página de [Releases](https://github.com/inusui/obsidian-termsidian/releases), descarga `termsidian-<versión>.zip`. Trae el plugin y el helper para los 6 sistemas.
2. Descomprímelo dentro de `<tu-vault>/.obsidian/plugins/`. En Windows, basta con clic derecho → **Extraer todo**. Tiene que quedar una carpeta `termsidian` dentro de `plugins`.

   Con Git Bash, macOS o Linux (cambia la versión por la que quieras):

   ```bash
   cd "<tu-vault>/.obsidian/plugins"
   curl -LO https://github.com/inusui/obsidian-termsidian/releases/download/0.1.0/termsidian-0.1.0.zip
   unzip -o termsidian-0.1.0.zip && rm termsidian-0.1.0.zip
   ```

3. Actívalo en Obsidian, como se explica en el último paso de la opción B.

### Opción B: compilando tú

1. Compila el helper y el plugin, como se explica en [Compilar](#compilar). El resultado queda en `plugin/build/`.
2. Copia el contenido de `plugin/build/` a `<tu-vault>/.obsidian/plugins/termsidian/`. Desde la raíz del repositorio, en Git Bash (o en la terminal de macOS o Linux):

   ```bash
   dest="<tu-vault>/.obsidian/plugins/termsidian"
   mkdir -p "$dest"
   cp -r plugin/build/. "$dest/"
   ```

3. En Obsidian, ve a **Ajustes → Complementos de la comunidad**, pulsa el botón de recargar junto a "Complementos instalados" y activa **Termsidian**.

La carpeta del plugin queda así. Con la opción A, `bin/` trae además los binarios de macOS y Linux:

```
<tu-vault>/.obsidian/plugins/termsidian/
 ├─ main.js
 ├─ manifest.json
 ├─ styles.css
 ├─ LICENSE
 ├─ THIRD_PARTY_NOTICES.md
 └─ bin/
     ├─ termsidian-pty-windows-amd64.exe
     └─ termsidian-pty-windows-arm64.exe
```

### Actualizar

1. Desactiva el plugin, o cierra Obsidian. Windows no deja reemplazar un `.exe` que está en uso.
2. Instala la versión nueva con cualquiera de las dos opciones. Los archivos se sobrescriben.
3. Activa el plugin otra vez.

> **Vault sincronizado (Nextcloud, Dropbox, OneDrive…):** la carpeta `bin/` también se sincroniza. Si instalas desde una Release, el plugin funciona en todos tus equipos, sean Windows, macOS o Linux. Si compilas tú, solo trae los binarios de Windows; mira [Otros sistemas](#otros-sistemas).

## Compilar

### Requisitos

- [Go](https://go.dev/) 1.27 o superior: `winget install GoLang.Go`
- [Node.js](https://nodejs.org/) 24 LTS, que incluye npm: `winget install OpenJS.NodeJS.LTS`

Después de instalarlos, abre una terminal nueva. Si no, ve a [Problemas frecuentes](#problemas-frecuentes).

Todos los comandos de este README son para Git Bash en Windows, o para la terminal de macOS o Linux.

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

### 2. El plugin (TypeScript)

```bash
cd plugin
npm install
npm run build
```

`npm run build` hace tres cosas:

1. Revisa los tipos con `tsc`.
2. Empaqueta el código con esbuild.
3. Crea `plugin/build/` con `main.js`, `manifest.json`, `styles.css` y las licencias (`LICENSE` y `THIRD_PARTY_NOTICES.md`), y copia los binarios de `helper/dist/` a `bin/`. Por eso hay que compilar el helper primero.

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

El plugin, desde `plugin/`:

```bash
npm test
```

Estos tests comprueban que los mensajes se reconstruyen aunque lleguen partidos en trozos, y que el formato es idéntico byte a byte al de Go. Uno de ellos habla con el binario de Go de verdad; se omite si `helper/dist/` no está compilado.

## Integración continua y releases

GitHub Actions ([.github/workflows/release.yml](.github/workflows/release.yml)) hace tres cosas:

- **En cada push, a cualquier rama,** ejecuta los tests del helper y del plugin en Windows, macOS y Linux. El resultado se ve en la pestaña **Actions** del repositorio y en la insignia del principio de este README.
- **También en cada push,** comprueba que las dependencias son las admitidas (ver [Dependencias](#dependencias)).
- **Cuando llegan commits a `main`** (por ejemplo, al hacer merge de `Dev`), y si los tests pasan en los tres sistemas y las dependencias están admitidas, publica una versión nueva sin que hagas nada más: calcula el número, lo guarda en `manifest.json`, crea la etiqueta y publica una Release con:

| Archivo | Qué es |
|---|---|
| `termsidian-<versión>.zip` | La carpeta del plugin completa, lista para descomprimir en `.obsidian/plugins/` |
| `main.js`, `manifest.json`, `styles.css` | El plugin, sin el helper |
| `termsidian-pty-<sistema>-<arquitectura>` | El helper para cada sistema |
| `checksums.txt` | Hash SHA-256 de cada binario del helper |
| `LICENSE`, `THIRD_PARTY_NOTICES.md` | La licencia de Termsidian y las de las dependencias que van dentro |

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

1. Haz merge de tu rama en `main` y súbelo:

   ```bash
   git switch main && git merge Dev && git push
   ```

2. Sigue el progreso en la pestaña **Actions**. Al terminar, la Release aparece en la página de Releases.

3. El pipeline sube a `main` un commit `chore(release): <versión>` con el `manifest.json` actualizado. Tráelo a tu rama de trabajo para que las dos queden iguales:

   ```bash
   git pull && git switch Dev && git merge main
   ```

Si `main` tiene reglas de protección que exigen pull requests, el pipeline no podrá subir ese commit. En ese caso, permite que GitHub Actions haga push a `main`.

### Dependencias

La Release lleva dentro código de otros proyectos: xterm.js en `main.js` y `styles.css`, y la biblioteca estándar de Go y varios módulos de Go en los binarios. Sus licencias piden incluir su aviso de copyright, que está en [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).

Ese archivo es también la lista de dependencias admitidas: cada sección `## nombre` es una. En cada push, el pipeline compara esa lista con las dependencias reales y falla si no coinciden:

- **Dependencias nuevas que no están en THIRD_PARTY_NOTICES.md:** añadiste una dependencia (o llegó como dependencia de otra). Si la quieres, añade su sección con el texto de su archivo `LICENSE`. Si no, quítala.
- **Dependencias que ya no se usan:** borra su sección.

Las dependencias de desarrollo (`devDependencies` de npm, como TypeScript o esbuild) no cuentan: no van dentro de la Release.

Para comprobarlo antes de subir, desde la raíz del repositorio:

```bash
bash .github/scripts/check-deps.sh
```

### Vulnerabilidades

GitHub avisa cuando una dependencia tiene una vulnerabilidad conocida: es **Dependabot alerts**, activado en **Settings → Code security**. Solo avisa: no abre pull requests ni cambia nada en el repositorio. La actualización se hace a mano:

1. Abre la alerta en la pestaña **Security → Dependabot** del repositorio. Dice qué dependencia es, qué gravedad tiene y a qué versión hay que subir.
2. En tu rama de trabajo, sube la dependencia a esa versión (los números son de ejemplo):

   ```bash
   # Un módulo de Go
   cd helper && go get golang.org/x/crypto@v0.52.0 && go mod tidy
   ```

   ```bash
   # Un paquete npm que está en package.json
   cd plugin && npm install @xterm/xterm@6.0.1
   ```

   ```bash
   # Un paquete npm que llega como dependencia de otro
   cd plugin && npm audit fix
   ```

   Si es una acción del workflow, cambia su versión en `release.yml`.

3. Pasa los tests y `bash .github/scripts/check-deps.sh`.
4. Haz commit. El tipo decide si se publica una versión al llegar a `main`:

| Qué actualizaste | Commit | Versión nueva |
|---|---|---|
| Algo que va dentro de la Release (módulos de Go, xterm.js) | `fix(deps): …` | patch, con el arreglo |
| Una herramienta de compilación (TypeScript, esbuild…) | `chore(deps): …` | ninguna |
| Una acción del workflow | `ci(deps): …` | ninguna |

5. Haz merge en `main` como siempre. La alerta se cierra sola cuando el arreglo llega a `main`.

## Problemas frecuentes

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

**La terminal muestra "[Termsidian] No encuentro el helper en …"**

Falta el binario de tu sistema en `bin/`. Compílalo, como se explica en [Compilar](#compilar), y vuelve a copiar el plugin.

**En macOS o Linux, el helper no arranca por falta de permisos**

Algunos servicios de sincronización no conservan el permiso de ejecución. Dáselo de nuevo:

```bash
chmod +x "<tu-vault>/.obsidian/plugins/termsidian/bin/"termsidian-pty-*
```

**Ver los mensajes de error**

Abre la consola de Obsidian con `Ctrl+Shift+I`, pestaña **Console**. Los mensajes del helper empiezan por `[termsidian-pty]`.

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
 │   ├─ protocol/          formato de los mensajes, con sus tests
 │   ├─ build.sh           compila los binarios en helper/dist/
 │   └─ go.mod, go.sum     dependencias y sus hashes
 └─ plugin/                el plugin, en TypeScript
     ├─ src/               main.ts, view.ts, session.ts, protocol.ts, styles.css y tests
     ├─ esbuild.config.mjs genera plugin/build/
     └─ package.json, package-lock.json
```

`helper/dist/`, `plugin/build/` y `plugin/node_modules/` se generan al compilar y no se suben al repositorio. `go.sum` y `package-lock.json` sí se suben.

## Seguridad

Si encuentras una vulnerabilidad, no abras un issue público: repórtala en privado como explica [SECURITY.md](SECURITY.md).

## Licencia

[MIT](LICENSE).
