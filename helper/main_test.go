package main

import (
	"encoding/binary"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/inusui/obsidian-termsidian/helper/protocol"
)

// Enter en una terminal es \r, no \n: es lo que manda xterm.js.
const enter = "\r"

// Tiempo máximo esperando al helper antes de dar el test por fallido.
const testTimeout = 15 * time.Second

// Variable de entorno con la que el binario de tests hace de helper.
const beHelperEnv = "TERMSIDIAN_BE_HELPER"

// TestMain, si existe, sustituye al arranque normal de los tests. Lo usamos
// para un truco habitual en Go: cuando un test necesita el helper como
// proceso aparte, relanza este mismo binario con beHelperEnv=1, y entonces
// en vez de correr los tests se comporta como el helper de verdad.
func TestMain(m *testing.M) {
	if os.Getenv(beHelperEnv) == "1" {
		main()
		os.Exit(0)
	}
	os.Exit(m.Run())
}

type frame struct {
	kind    protocol.Kind
	payload []byte
}

// fakePlugin hace de plugin en los tests: manda mensajes al helper y lee los
// que devuelve, como hará Obsidian por stdin/stdout.
type fakePlugin struct {
	t        *testing.T
	toHelper io.WriteCloser
	frames   chan frame // lo que va mandando el helper
	runErr   chan error // lo que devolvió run (nil si el helper es otro proceso)
	screen   string     // todo el texto DATA recibido hasta ahora
}

// newFakePlugin conecta un fakePlugin a un helper y espera su HELLO.
func newFakePlugin(t *testing.T, toHelper io.WriteCloser, fromHelper io.Reader) *fakePlugin {
	t.Helper() // si algo falla aquí, el error apunta a la línea del test
	p := &fakePlugin{t: t, toHelper: toHelper, frames: make(chan frame, 1000)}
	go func() {
		for {
			kind, payload, err := protocol.ReadFrame(fromHelper)
			if err != nil {
				close(p.frames)
				return
			}
			p.frames <- frame{kind, payload}
		}
	}()
	// t.Cleanup se ejecuta al terminar el test, pase lo que pase.
	t.Cleanup(func() { toHelper.Close() })

	if f := p.next(); f.kind != protocol.Hello || f.payload[0] != protocol.Version {
		t.Fatalf("primer mensaje: %v %q, se esperaba HELLO", f.kind, f.payload)
	}
	return p
}

// startHelper ejecuta run en una goroutine de este mismo proceso, conectado
// por tuberías en memoria (io.Pipe). A diferencia de un bytes.Buffer, quien
// lee de un io.Pipe espera a que haya datos en vez de recibir EOF.
func startHelper(t *testing.T, opts options) *fakePlugin {
	t.Helper()
	inR, inW := io.Pipe()
	outR, outW := io.Pipe()
	runErr := make(chan error, 1)
	go func() {
		runErr <- run(opts, inR, outW)
		outW.Close()
	}()
	p := newFakePlugin(t, inW, outR)
	p.runErr = runErr
	return p
}

// startHelperProcess lanza el helper como proceso aparte, igual que hará
// Obsidian. Devuelve también el proceso, para poder matarlo.
func startHelperProcess(t *testing.T, args ...string) (*fakePlugin, *exec.Cmd) {
	t.Helper()
	// os.Args[0] es la ruta de este binario de tests.
	cmd := exec.Command(os.Args[0], args...)
	cmd.Env = append(os.Environ(), beHelperEnv+"=1")
	cmd.Stderr = os.Stderr // los logs del helper, a la vista
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cmd.Process.Kill()
		cmd.Wait()
	})
	return newFakePlugin(t, stdin, stdout), cmd
}

func (p *fakePlugin) send(kind protocol.Kind, payload []byte) {
	p.t.Helper()
	if err := protocol.WriteFrame(p.toHelper, kind, payload); err != nil {
		p.t.Fatalf("enviando %v: %v", kind, err)
	}
}

func (p *fakePlugin) typeLine(line string) {
	p.t.Helper()
	p.send(protocol.Data, []byte(line+enter))
}

func (p *fakePlugin) resize(cols, rows uint16) {
	p.t.Helper()
	payload := make([]byte, 4)
	binary.BigEndian.PutUint16(payload[0:], cols)
	binary.BigEndian.PutUint16(payload[2:], rows)
	p.send(protocol.Resize, payload)
}

// next espera el siguiente mensaje del helper, con límite de tiempo.
func (p *fakePlugin) next() frame {
	p.t.Helper()
	select {
	case f, ok := <-p.frames:
		if !ok {
			p.t.Fatalf("el helper cerró la salida. Pantalla:\n%s", p.screen)
		}
		if f.kind == protocol.Data {
			p.screen += string(f.payload)
		}
		return f
	case <-time.After(testTimeout):
		p.t.Fatalf("el helper no respondió en %v. Pantalla:\n%s", testTimeout, p.screen)
	}
	return frame{} // no se llega aquí: Fatalf detiene el test
}

// waitFor lee mensajes hasta que la pantalla contiene text.
func (p *fakePlugin) waitFor(text string) {
	p.t.Helper()
	for !strings.Contains(p.screen, text) {
		if f := p.next(); f.kind == protocol.Exit {
			p.t.Fatalf("el shell salió sin mostrar %q. Pantalla:\n%s", text, p.screen)
		}
	}
}

// waitExit lee mensajes hasta EXIT y devuelve el código de salida.
func (p *fakePlugin) waitExit() int32 {
	p.t.Helper()
	for {
		f := p.next()
		if f.kind != protocol.Exit {
			continue
		}
		if p.runErr != nil {
			if err := <-p.runErr; err != nil {
				p.t.Fatalf("run devolvió un error: %v", err)
			}
		}
		return int32(binary.BigEndian.Uint32(f.payload))
	}
}

func TestVersionFlag(t *testing.T) {
	cmd := exec.Command(os.Args[0], "--version")
	cmd.Env = append(os.Environ(), beHelperEnv+"=1")
	// Output ejecuta el proceso y devuelve todo lo que escribió en stdout.
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(out)); got != version {
		t.Errorf("--version imprimió %q, se esperaba %q", got, version)
	}
}

// Sin PTY, timeout se negaba a funcionar (paso 2). El texto que buscamos
// (tty-Windows_NT, tty-2) solo aparece si el shell lo calcula; no está en lo
// que tecleamos, que también sale en pantalla porque la terminal hace eco.
func TestRealTerminal(t *testing.T) {
	opts := options{shell: "/bin/sh", cols: 80, rows: 24, cwd: t.TempDir()}
	command, want := "[ -t 0 ] && echo tty-$((1+1))", "tty-2"
	if runtime.GOOS == "windows" {
		opts.shell = "cmd.exe"
		// Ruta completa: con Git Bash en el PATH (como en GitHub Actions),
		// "timeout" sería el de GNU, que no entiende /t.
		command, want = `%SystemRoot%\System32\timeout.exe /t 1 /nobreak >nul && echo tty-%OS%`, "tty-Windows_NT"
	}

	p := startHelper(t, opts)
	p.typeLine(command)
	p.waitFor(want)
	p.typeLine("exit 3")
	if code := p.waitExit(); code != 3 {
		t.Errorf("código de salida = %d, se esperaba 3", code)
	}
	t.Logf("pantalla:\n%s", p.screen)
}

func TestResize(t *testing.T) {
	opts := options{shell: "/bin/sh", cols: 80, rows: 24}
	command := `stty size | { read r c; echo "size=${c}x${r}"; }`
	if runtime.GOOS == "windows" {
		opts.shell = "powershell.exe"
		opts.args = []string{"-NoLogo", "-NoProfile"}
		command = `Write-Host "size=$([Console]::WindowWidth)x$([Console]::WindowHeight)"`
	}

	p := startHelper(t, opts)
	p.resize(123, 37)
	p.typeLine(command)
	p.waitFor("size=123x37")
	p.typeLine("exit")
	p.waitExit()
}

// RF-06: si el plugin se va (se cierra Obsidian), el shell no puede quedarse
// vivo aunque nadie haya escrito exit.
func TestClosingInputKillsShell(t *testing.T) {
	shell := "/bin/sh"
	if runtime.GOOS == "windows" {
		shell = "cmd.exe"
	}

	p := startHelper(t, options{shell: shell, cols: 80, rows: 24})
	p.toHelper.Close()
	code := p.waitExit()
	t.Logf("código de salida tras cerrar la entrada: %d", code)
}
