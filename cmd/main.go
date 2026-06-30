package main

import (
	"fmt"
	"os"

	"ai-manager/internal/cli"
	"ai-manager/internal/tui"
)

func main() {
	root := cli.NewRootCmd()

	var tuiFlag bool
	root.PersistentFlags().BoolVar(&tuiFlag, "tui", false, "Launch TUI mode")

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if tuiFlag {
		if err := tui.Launch(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
