// Command toolguard is the ToolGuard CLI. All logic lives in internal/cli.
package main

import (
	"os"

	"github.com/KunjJoshi/toolguard/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
