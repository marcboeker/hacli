package main

import (
	"fmt"
	"os"

	"github.com/marcboeker/hacli/internal/cmds"
)

func main() {
	if err := cmds.NewRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
