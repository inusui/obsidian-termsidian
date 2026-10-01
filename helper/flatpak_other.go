//go:build !linux

package main

import (
	"errors"
	"os/exec"

	"github.com/aymanbagabas/go-pty"
)

// Flatpak solo existe en Linux; insideFlatpak nunca es true aquí.
func startOnHost(p pty.Pty, path string, args []string, dir string) (*exec.Cmd, error) {
	return nil, errors.New("Flatpak solo existe en Linux")
}
