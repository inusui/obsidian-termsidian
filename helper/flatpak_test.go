package main

import (
	"os/exec"
	"regexp"
	"testing"
	"time"

	"github.com/inusui/obsidian-termsidian/helper/protocol"
)

// Este test solo corre dentro de un sandbox de Flatpak; en GitHub Actions se
// salta. Para probarlo con el sandbox de Obsidian:
//
//	CGO_ENABLED=0 go test -c -o ~/helper.test   (en $HOME: el sandbox no ve /tmp)
//	flatpak run --talk-name=org.freedesktop.Flatpak --command=$HOME/helper.test md.obsidian.Obsidian -test.run Flatpak -test.v
//
// Sin --talk-name comprueba el aviso; con él, que el shell es el del host.
func TestFlatpakHostShell(t *testing.T) {
	if !insideFlatpak() {
		t.Skip("no estamos dentro de Flatpak")
	}

	p := startHelper(t, options{cols: 80, rows: 24, cwd: "/tmp"})
	if exec.Command("flatpak-spawn", "--host", "true").Run() != nil {
		p.waitFor(flatpakOverride)
		p.typeLine("exit")
		p.waitExit()
		return
	}

	// En el host no existe /.flatpak-info.
	p.typeLine(`[ -e /.flatpak-info ] || echo "host-$((1+1)) $TERM pid-$$"`)
	p.waitFor("host-2 xterm-256color")
	pid := regexp.MustCompile(`pid-(\d+)`).FindStringSubmatch(p.screen)[1]

	p.resize(123, 37)
	p.typeLine(`stty size | { read r c; echo "size=${c}x${r}"; }`)
	p.waitFor("size=123x37")

	// Ctrl+C solo llega a sleep si el shell del host controla la PTY. Si no,
	// el echo esperaría los 30 s y el test se pasaría de tiempo.
	p.typeLine("sleep 30")
	p.send(protocol.Data, []byte("\x03"))
	p.typeLine(`echo "ctrl-c-$((2+2))"`)
	p.waitFor("ctrl-c-4")

	// RF-06: al cerrarse el plugin, el shell del host también tiene que morir.
	p.toHelper.Close()
	p.waitExit()
	for i := 0; ; i++ {
		alive := exec.Command("flatpak-spawn", "--host", "kill", "-0", pid).Run() == nil
		if !alive {
			break
		}
		if i == 50 {
			t.Fatalf("el shell del host (PID %s) sigue vivo tras cerrar el plugin", pid)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
