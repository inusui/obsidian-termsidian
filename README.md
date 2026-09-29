# Termsidian

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

1. Compila el helper y el plugin, como se explica en [Compilar](#compilar). El resultado queda en `plugin/build/`.
2. Copia el contenido de `plugin/build/` a `<tu-vault>/.obsidian/plugins/termsidian/`. `<tu-vault>` es la carpeta que contiene `.obsidian`.

   Desde la raíz del repositorio, en Git Bash (o en la terminal de macOS o Linux):

   ```bash
   dest="<tu-vault>/.obsidian/plugins/termsidian"
   mkdir -p "$dest"
   cp -r plugin/build/. "$dest/"
   ```

   En Git Bash, las rutas de Windows se escriben con barras normales y la unidad en minúscula: `C:\Users\Ana\Notas` se escribe `/c/Users/Ana/Notas`.

3. En Obsidian, ve a **Ajustes → Complementos de la comunidad**, pulsa el botón de recargar junto a "Complementos instalados" y activa **Termsidian**.

La carpeta del plugin queda así:

```
<tu-vault>/.obsidian/plugins/termsidian/
 ├─ main.js
 ├─ manifest.json
 ├─ styles.css
 └─ bin/
     ├─ termsidian-pty-windows-amd64.exe
     └─ termsidian-pty-windows-arm64.exe
```

### Actualizar

1. Desactiva el plugin, o cierra Obsidian. Windows no deja reemplazar un `.exe` que está en uso.
2. Vuelve a compilar y a copiar, con los mismos comandos de arriba.
3. Activa el plugin otra vez.

> **Vault sincronizado (Nextcloud, Dropbox, OneDrive…):** la carpeta `bin/` también se sincroniza. En tus otros equipos Windows funciona sin instalar nada. En macOS o Linux hace falta el binario de ese sistema; mira [Otros sistemas](#otros-sistemas).

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
3. Crea `plugin/build/` con `main.js`, `manifest.json` y `styles.css`, y copia los binarios de `helper/dist/` a `bin/`. Por eso hay que compilar el helper primero.

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
 ├─ .gitattributes         guarda los .sh con saltos de línea LF
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

## Pendiente

- Pestaña de ajustes: shell, argumentos, fuente y tamaño.
- Varias terminales, en pestañas.
- Qué hacer con los atajos de Obsidian cuando la terminal tiene el foco.
- Probar en macOS y Linux.
- Publicar con GitHub Actions: compilar los 6 binarios, generar `checksums.txt` y subirlos a una Release.
- Añadir el archivo de licencia (MIT).
