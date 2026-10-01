// termsidian-pty: el helper que lanza el shell del usuario.
//
// Paso 4: el shell corre dentro de una PTY (ConPTY en Windows), así que cree
// estar en una terminal real: colores, Tab, Ctrl+C y cambios de tamaño.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sync"
	"time"

	"github.com/aymanbagabas/go-pty"

	// Los paquetes propios se importan con la ruta del módulo (go.mod)
	// más la carpeta. Se usan con el nombre del paquete: protocol.Data.
	"github.com/inusui/obsidian-termsidian/helper/protocol"
)

// version es var y no const para poder cambiarla al compilar, con
// -ldflags "-X main.version=0.1.0" (ver build.sh). Sin eso vale "dev".
var version = "dev"

// Cuánto esperamos a que llegue la salida pendiente cuando el shell termina.
const drainTimeout = 200 * time.Millisecond

// options agrupa la configuración de una sesión. Un struct es un tipo con
// campos con nombre, como un objeto sin métodos.
type options struct {
	shell      string
	args       []string
	cwd        string
	cols, rows int
}

func main() {
	var opts options
	// Las variantes ...Var escriben en una variable que ya existe. &opts.shell
	// es la dirección de ese campo: un puntero para que flag lo rellene.
	flag.StringVar(&opts.shell, "shell", "", "shell a ejecutar (vacío = el del sistema)")
	flag.StringVar(&opts.cwd, "cwd", "", "directorio inicial (vacío = el actual)")
	flag.IntVar(&opts.cols, "cols", 80, "columnas iniciales")
	flag.IntVar(&opts.rows, "rows", 24, "filas iniciales")
	// flag.Bool devuelve un puntero; el valor se lee con *showVersion.
	showVersion := flag.Bool("version", false, "muestra la versión y sale")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}
	// Lo que sobra tras los flags va al shell, por ejemplo:
	//   termsidian-pty --shell powershell.exe -- -NoLogo
	opts.args = flag.Args()

	// Ctrl+C es para el shell, no para el helper. Sin esto el helper
	// moriría y dejaría al shell huérfano.
	signal.Ignore(os.Interrupt)

	if err := run(opts, os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "termsidian-pty:", err)
		os.Exit(1)
	}
}

// run lanza el shell en una PTY y lo conecta con el plugin: lee mensajes de in
// y escribe mensajes en out. Recibe io.Reader/io.Writer en lugar de usar
// os.Stdin y os.Stdout directamente para que los tests puedan simular al plugin.
func run(opts options, in io.Reader, out io.Writer) error {
	// Dentro de Flatpak el shell se lanza a través de flatpak-spawn.
	name, args, notice := shellCommand(opts)
	// Ruta completa del shell. go-pty no la busca y, si hay cwd, en Windows
	// buscaría "cmd.exe" dentro de cwd.
	shellPath, err := exec.LookPath(name)
	if err != nil {
		return err
	}

	p, err := pty.New()
	if err != nil {
		return err
	}
	// La PTY se cierra a mano más abajo, y con defer por si salimos antes por
	// un error. Pero en Windows cerrarla dos veces libera la misma memoria dos
	// veces y el programa se cae. sync.Once garantiza que Close corra una
	// sola vez, la llamen quien la llame.
	var closeOnce sync.Once
	closePty := func() { closeOnce.Do(func() { p.Close() }) }
	// defer ejecuta closePty al salir de run, pase lo que pase.
	defer closePty()

	if err := p.Resize(opts.cols, opts.rows); err != nil {
		return err
	}

	// proc es el proceso lanzado y wait espera a que termine. Son dos
	// variables porque go-pty y os/exec devuelven tipos distintos.
	var proc *os.Process
	var wait func() *os.ProcessState
	// stop mata al shell cuando el plugin se va.
	stop := func() { proc.Kill() }
	if name == "flatpak-spawn" {
		cmd, err := startOnHost(p, shellPath, args, opts.cwd)
		if err != nil {
			return err
		}
		proc, wait = cmd.Process, func() *os.ProcessState { cmd.Wait(); return cmd.ProcessState }
		// Matar a flatpak-spawn no mata al shell del host. Cerrar la PTY sí:
		// el sistema manda SIGHUP a la sesión que la controla.
		stop = func() { closePty(); proc.Kill() }
	} else {
		cmd := p.Command(shellPath, args...)
		cmd.Dir = opts.cwd
		// Les decimos a los programas qué terminal tienen delante: xterm.js.
		cmd.Env = append(os.Environ(), "TERM=xterm-256color", "COLORTERM=truecolor")
		if err := cmd.Start(); err != nil {
			return err
		}
		proc, wait = cmd.Process, func() *os.ProcessState { cmd.Wait(); return cmd.ProcessState }
	}

	// En Unix cerramos nuestra copia del lado de la PTY que usa el shell; si
	// no, leer la PTY esperaría para siempre aunque el shell ya haya salido.
	// p.(pty.UnixPty) es una "type assertion": pregunta si p es también un
	// UnixPty. En Windows ok vale false y no se hace nada.
	if u, ok := p.(pty.UnixPty); ok {
		u.Slave().Close()
	}

	// Lo primero que ve el plugin: estamos vivos y esta es nuestra versión.
	if err := protocol.WriteFrame(out, protocol.Hello, protocol.HelloPayload(version)); err != nil {
		return err
	}
	if notice != "" {
		if err := protocol.WriteFrame(out, protocol.Data, []byte(notice)); err != nil {
			return err
		}
	}

	// Plugin → shell. Si el plugin cierra la tubería (se cerró Obsidian o la
	// pestaña), matamos al shell para no dejar huérfanos (RF-06). Si es el
	// helper el que muere de golpe, no hace falta nada: en Unix el sistema
	// manda SIGHUP a la sesión del shell al cerrarse la PTY, y en Windows la
	// ConPTY se cierra y termina a los procesos de su consola (lo comprueba
	// kill_windows_test.go). Los programas con ventana propia sobreviven,
	// como en cualquier terminal.
	go func() {
		readInput(in, p)
		stop()
	}()

	// Shell → plugin, también en paralelo: con una PTY la salida no siempre
	// se acaba sola cuando el shell sale. done es un canal; close(done) avisa
	// a quien esté esperando en <-done.
	done := make(chan struct{})
	go func() {
		pumpOutput(p, out)
		close(done)
	}()

	state := wait()

	// El shell terminó. En Unix la lectura acaba sola al vaciarse la PTY; en
	// Windows la ConPTY mantiene la tubería abierta hasta que la cerramos.
	// select espera a lo primero que ocurra: que acabe la lectura o que pase
	// el tiempo.
	select {
	case <-done:
	case <-time.After(drainTimeout):
	}
	closePty()
	<-done

	return protocol.WriteFrame(out, protocol.Exit, protocol.ExitPayload(state.ExitCode()))
}

// terminal es lo que readInput necesita de la PTY. En Go es habitual definir
// interfaces pequeñas donde se usan: pty.Pty ya cumple esta sin saberlo.
type terminal interface {
	io.Writer // una interfaz puede incluir otra
	Resize(width, height int) error
}

// readInput lee mensajes del plugin hasta que se cierra la entrada.
func readInput(in io.Reader, term terminal) error {
	for {
		kind, payload, err := protocol.ReadFrame(in)
		if err != nil {
			return err // io.EOF: el plugin cerró la tubería
		}
		switch kind {
		case protocol.Data:
			if _, err := term.Write(payload); err != nil {
				return err
			}
		case protocol.Resize:
			cols, rows, err := protocol.ParseResize(payload)
			if err != nil {
				log.Print(err) // stderr: solo para logs, el plugin lo lee aparte
				continue
			}
			if err := term.Resize(int(cols), int(rows)); err != nil {
				log.Print(err)
			}
		default:
			// Tipos desconocidos se ignoran: así un plugin más nuevo no
			// rompe a un helper más viejo.
		}
	}
}

// pumpOutput lee la salida del shell y la reenvía en mensajes DATA.
func pumpOutput(shellOut io.Reader, out io.Writer) error {
	buf := make([]byte, 32*1024)
	for {
		n, err := shellOut.Read(buf)
		// Regla de io.Reader: primero se usan los n bytes leídos y después
		// se mira el error, porque pueden llegar las dos cosas a la vez.
		if n > 0 {
			if err := protocol.WriteFrame(out, protocol.Data, buf[:n]); err != nil {
				return err
			}
		}
		if err != nil {
			return err // la PTY se cerró
		}
	}
}

// defaultShell elige el shell según el sistema operativo (RF-02).
func defaultShell() string {
	if runtime.GOOS == "windows" {
		return "powershell.exe"
	}
	if s := os.Getenv("SHELL"); s != "" {
		return s
	}
	if runtime.GOOS == "darwin" {
		return "/bin/zsh"
	}
	return "/bin/bash"
}
