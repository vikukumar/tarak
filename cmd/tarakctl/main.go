// tarakctl — Unified CLI for Tarak container orchestration platform.
package main

import (
	"fmt"
	"os"

	"github.com/vikukumar/tarak/pkg/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
