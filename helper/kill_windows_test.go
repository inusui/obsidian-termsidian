package main

import (
	"os"
	"regexp"
	"strconv"
	"testing"
	"time"
)

// RF-06: si el helper muere de golpe (Administrador de tareas, Obsidian
// colgado), el shell no puede quedarse vivo.
func TestKillingHelperKillsShell(t *testing.T) {
	p, helper := startHelperProcess(t, "--shell", "powershell.exe", "--", "-NoLogo", "-NoProfile")

	// Le pedimos al shell su PID. La expresión regular busca el número que
	// imprime, no lo que tecleamos ($PID sin expandir).
	p.typeLine(`Write-Host "pid=$PID."`)
	pidRe := regexp.MustCompile(`pid=(\d+)\.`)
	var match []string
	for match == nil {
		p.next()
		match = pidRe.FindStringSubmatch(p.screen)
	}
	pid, _ := strconv.Atoi(match[1])

	// Abrimos el shell ANTES de matar al helper. Si lo hiciéramos después y
	// fallara, no sabríamos si es porque murió o por otro motivo.
	shell, err := os.FindProcess(pid)
	if err != nil {
		t.Fatalf("no se pudo abrir el shell (PID %d): %v", pid, err)
	}

	// En Windows, Kill es TerminateProcess: muerte inmediata, sin avisar y
	// sin que el helper pueda ejecutar nada más.
	helper.Process.Kill()
	helper.Wait()

	exited := make(chan struct{})
	go func() {
		shell.Wait()
		close(exited)
	}()
	select {
	case <-exited:
	case <-time.After(5 * time.Second):
		shell.Kill() // no dejemos basura del propio test
		t.Fatalf("el shell (PID %d) sigue vivo 5 s después de matar al helper", pid)
	}
}
