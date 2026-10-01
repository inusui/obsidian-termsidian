package main

import (
	"errors"
	"os/exec"
	"syscall"

	"github.com/aymanbagabas/go-pty"
)

// startOnHost lanza flatpak-spawn conectado a la PTY pero, a diferencia de
// p.Command, sin hacerla su terminal de control. Una terminal solo puede
// controlar una sesión: si se la quedara flatpak-spawn, el shell del host no
// podría tomarla y no habría Ctrl+C ni control de trabajos.
func startOnHost(p pty.Pty, path string, args []string, dir string) (*exec.Cmd, error) {
	u, ok := p.(pty.UnixPty)
	if !ok {
		return nil, errors.New("la PTY no es de Unix")
	}
	cmd := exec.Command(path, args...)
	cmd.Dir = dir
	cmd.Stdin, cmd.Stdout, cmd.Stderr = u.Slave(), u.Slave(), u.Slave()
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	return cmd, cmd.Start()
}
