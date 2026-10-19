# Termsidian

[![CI y release](https://github.com/inusui/obsidian-termsidian/actions/workflows/release.yml/badge.svg)](https://github.com/inusui/obsidian-termsidian/actions/workflows/release.yml)

Terminal real en la barra lateral derecha de Obsidian, con el shell de tu sistema (PowerShell, cmd, bash, zsh…).

Sin Python y sin node-pty: un programa pequeño escrito en Go, `termsidian-pty`, abre una terminal del sistema (ConPTY en Windows, PTY en macOS y Linux), y el plugin la dibuja con [xterm.js](https://xtermjs.org/).

*Termsidian es un proyecto independiente, sin afiliación ni respaldo de Obsidian.*

(Read in English: [README.md](README.md))

## Sistemas probados

| Sistema | Instalación de Obsidian | Estado |
|---|---|---|
| Windows 11 | Instalador oficial | ✅ Funciona |
| Linux (Fedora KDE) | Flatpak | ✅ Funciona (requiere [un permiso](#linux-con-flatpak)) |
| macOS | — | ⚪ Compila, sin probar |

## Qué hace

- Abre una terminal en la barra lateral derecha, empezando en la carpeta del vault.
- Es una terminal de verdad: colores, autocompletado con Tab, Ctrl+C, historial y programas interactivos.
- Se reajusta al cambiar el tamaño del panel.
- Usa la fuente monoespaciada y los colores de tu tema de Obsidian, y cambia con él.
- Al cerrar el panel no quedan procesos huérfanos.

## Uso

Abre la terminal con el icono de terminal de la barra izquierda, o con `Ctrl+P` → **Termsidian: Abrir terminal**.

El shell es PowerShell en Windows, y en macOS y Linux el de la variable `$SHELL`. Si Obsidian está instalado con Flatpak, hace falta un paso más: ve a [Linux con Flatpak](#linux-con-flatpak).

| Tecla | Qué hace |
|---|---|
| `Ctrl+C` | Con texto seleccionado, lo copia. Sin selección, interrumpe el programa, como siempre. |
| `Ctrl+V` | Pega el portapapeles. |
| `Enter` después de `exit` | Abre un shell nuevo. |

Al cerrar el panel se cierran el shell y todo lo que corría dentro de él. Los programas con ventana propia que abriste desde la terminal (por ejemplo `code .`) siguen abiertos.

## Instalación en tu vault

**Requisitos:** Obsidian de escritorio 1.7.2 o superior, con los complementos de la comunidad activados (modo restringido desactivado).

`<tu-vault>` es la carpeta que contiene `.obsidian`. En Git Bash, las rutas de Windows se escriben con barras normales y la unidad en minúscula: `C:\Users\Ana\Notas` se escribe `/c/Users/Ana/Notas`.

1. En la página de [Releases](https://github.com/inusui/obsidian-termsidian/releases), descarga `termsidian-<versión>.zip`. Trae el plugin y el helper para los 6 sistemas.
2. Descomprímelo dentro de `<tu-vault>/.obsidian/plugins/`. En Windows, basta con clic derecho → **Extraer todo**. Tiene que quedar una carpeta `termsidian` dentro de `plugins`.

   Con Git Bash, macOS o Linux (cambia la versión por la que quieras):

   ```bash
   cd "<tu-vault>/.obsidian/plugins"
   curl -LO https://github.com/inusui/obsidian-termsidian/releases/download/0.1.0/termsidian-0.1.0.zip
   unzip -o termsidian-0.1.0.zip && rm termsidian-0.1.0.zip
   ```

3. En Obsidian, ve a **Ajustes → Complementos de la comunidad**, pulsa el botón de recargar junto a "Complementos instalados" y activa **Termsidian**.

La carpeta del plugin queda así:

```
<tu-vault>/.obsidian/plugins/termsidian/
 ├─ main.js
 ├─ manifest.json
 ├─ styles.css
 ├─ LICENSE
 ├─ THIRD_PARTY_NOTICES.md
 └─ bin/
     ├─ termsidian-pty-windows-amd64.exe
     ├─ termsidian-pty-windows-arm64.exe
     ├─ termsidian-pty-darwin-amd64
     ├─ termsidian-pty-darwin-arm64
     ├─ termsidian-pty-linux-amd64
     └─ termsidian-pty-linux-arm64
```

¿Prefieres compilarlo tú en vez de descargar la Release? Mira [DEVELOPMENT.md](DEVELOPMENT.md).

### Actualizar

1. Desactiva el plugin, o cierra Obsidian. Windows no deja reemplazar un `.exe` que está en uso.
2. Instala la versión nueva descomprimiendo la Release nueva en el mismo sitio. Los archivos se sobrescriben.
3. Activa el plugin otra vez.

> **Vault sincronizado (Nextcloud, Dropbox, OneDrive…):** la carpeta `bin/` también se sincroniza, así que el plugin funciona en todos tus equipos, sean Windows, macOS o Linux.

### Linux con Flatpak

Si instalaste Obsidian desde Flathub, Discover o GNOME Software, corre dentro de un sandbox de Flatpak que tiene su propio sistema. Sin este paso, la terminal se abre dentro de ese sandbox: muestra un prompt como `sh-5.3$` y no encuentra tus programas.

Termsidian detecta Flatpak y abre el shell de tu sistema real con `flatpak-spawn --host`. Para eso, Obsidian necesita un permiso que no trae por defecto. Ejecútalo una sola vez en una terminal normal, fuera de Obsidian:

```bash
flatpak override --user --talk-name=org.freedesktop.Flatpak md.obsidian.Obsidian
```

Después cierra Obsidian del todo y vuelve a abrirlo. Mientras falte el permiso, la terminal funciona dentro del sandbox y muestra un aviso en amarillo con este mismo comando.

> **Qué implica este permiso:** deja que Obsidian y todos sus complementos ejecuten programas fuera del sandbox, igual que una instalación sin Flatpak. Es lo que necesita una terminal para ver tu sistema. Para quitarlo:
>
> ```bash
> flatpak override --user --no-talk-name=org.freedesktop.Flatpak md.obsidian.Obsidian
> ```

La terminal empieza en la carpeta del vault. Si esa carpeta no existe fuera del sandbox (por ejemplo, un vault abierto a través del portal de documentos), empieza en tu carpeta personal.

## Problemas frecuentes

**La terminal muestra "[Termsidian] No encuentro el helper en …"**

Falta el binario de tu sistema en `bin/`. Vuelve a instalar desde una Release ([Instalación en tu vault](#instalación-en-tu-vault)).

**En macOS o Linux, el helper no arranca por falta de permisos**

Algunos servicios de sincronización no conservan el permiso de ejecución. Dáselo de nuevo:

```bash
chmod +x "<tu-vault>/.obsidian/plugins/termsidian/bin/"termsidian-pty-*
```

**En Linux, el prompt es `sh-5.3$` y no encuentra mis programas**

Obsidian está instalado con Flatpak y le falta el permiso. Ve a [Linux con Flatpak](#linux-con-flatpak).

**Ver los mensajes de error**

Abre la consola de Obsidian con `Ctrl+Shift+I`, pestaña **Console**. Los mensajes del helper empiezan por `[termsidian-pty]`.

## ¿Por qué un helper en Go?

- `node-pty` hay que recompilarlo para cada versión de Electron, y se rompe cuando Obsidian se actualiza.
- Un helper en Python necesita Python instalado, y el módulo `pty` no existe en Windows.
- Go genera un único binario sin dependencias para cada sistema.

Más detalle de cómo se comunican el plugin y el helper en [DEVELOPMENT.md](DEVELOPMENT.md#cómo-funciona).

## Contribuir

Termsidian lo mantiene una sola persona. Las sugerencias e informes de fallos son bienvenidos como [issue](https://github.com/inusui/obsidian-termsidian/issues), pero **no se aceptan pull requests**: cualquier PR se cierra automáticamente con un comentario que enlaza aquí. Si quieres proponer un cambio, abre un issue describiéndolo.

## Seguridad

Si encuentras una vulnerabilidad, no abras un issue público: repórtala en privado como explica [SECURITY.md](SECURITY.md).

## Licencia

[MIT](LICENSE).
