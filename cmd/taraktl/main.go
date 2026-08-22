// taraktl is a convenience alias for tarakctl.
package main

import (
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	// Look for tarakctl in the same directory or PATH
	selfDir := filepath.Dir(os.Args[0])
	tarakctlPath := filepath.Join(selfDir, "tarakctl")
	if _, err := os.Stat(tarakctlPath + ".exe"); err == nil {
		tarakctlPath += ".exe"
	}

	cmd := exec.Command(tarakctlPath, os.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	}
}
