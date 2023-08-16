package main

import (
	"fmt"
	"os"

	"github.com/pattonjp/xbox-sync/pkg/cmds"
)

func main() {
	if err := cmds.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
