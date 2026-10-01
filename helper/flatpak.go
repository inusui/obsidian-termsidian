package main

import (
	"os"
	"os/exec"
	"runtime"
)

// Obsidian instalado con Flatpak (lo habitual en Linux) corre dentro de un
// sandbox con su propio sistema: ahí $SHELL es /bin/sh y no están los
// programas del usuario. flatpak-spawn --host lanza el shell fuera, en el
// sistema real, pero necesita que Obsidian pueda hablar con
// org.freedesktop.Flatpak, y eso hay que permitirlo una vez a mano.

// flatpakOverride es el comando que da ese permiso.
const flatpakOverride = "flatpak override --user --talk-name=org.freedesktop.Flatpak md.obsidian.Obsidian"

// flatpakNotice sale en la terminal si falta el permiso. \r\n porque la
// terminal no convierte \n en salto de línea completo.
const flatpakNotice = "\x1b[33m[Termsidian] Obsidian corre en un sandbox de Flatpak y esta terminal\r\n" +
	"está dentro de él, sin tus programas. Para usar el shell de tu sistema,\r\n" +
	"ejecuta esto fuera de Obsidian y reinícialo:\r\n\r\n" +
	"  " + flatpakOverride + "\x1b[0m\r\n\r\n"

// hostShell corre en el host con sh -c. Recibe la carpeta inicial, el shell
// (vacío = el del usuario) y los argumentos del shell. La carpeta puede no
// existir fuera del sandbox (su /tmp es otro); entonces empieza en $HOME,
// porque flatpak-spawn fallaría entero si no puede entrar en ella: por eso
// le decimos --directory=/, que existe siempre, y el cd lo hace el script. $SHELL es el del
// host; si no está, se lee de la base de usuarios.
const hostShell = `cd -- "$1" 2>/dev/null || cd; s=$2; shift 2
[ -n "$s" ] || s=${SHELL:-$(getent passwd "$(id -un)" | cut -d: -f7)}
exec "${s:-/bin/sh}" "$@"`

func insideFlatpak() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	_, err := os.Stat("/.flatpak-info")
	return err == nil
}

// shellCommand decide qué programa lanzar en la PTY y con qué argumentos.
// notice es un aviso para mostrar en la terminal antes del shell, o "".
func shellCommand(opts options) (name string, args []string, notice string) {
	shell := opts.shell
	if !insideFlatpak() {
		if shell == "" {
			shell = defaultShell()
		}
		return shell, opts.args, ""
	}

	// Sin permiso, flatpak-spawn --host falla al momento. Entonces usamos el
	// shell del sandbox, que algo hace, y avisamos de cómo arreglarlo.
	if err := exec.Command("flatpak-spawn", "--host", "true").Run(); err != nil {
		if shell == "" {
			shell = defaultShell()
		}
		return shell, opts.args, flatpakNotice
	}

	// --watch-bus: si flatpak-spawn muere, el proceso del host también.
	// El entorno del sandbox no pasa al host; TERM y COLORTERM hay que
	// mandarlos a mano.
	// "termsidian" hace de $0 en sh -c; lo que sigue son $1, $2…
	args = []string{
		"--host", "--watch-bus", "--directory=/", "--env=TERM=xterm-256color", "--env=COLORTERM=truecolor",
		"/bin/sh", "-c", hostShell, "termsidian", opts.cwd, shell,
	}
	return "flatpak-spawn", append(args, opts.args...), ""
}
