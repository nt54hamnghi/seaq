package main

import (
	"os"

	"github.com/nt54hamnghi/seaq/cmd"
)

func main() {
	if _, err := cmd.New().ExecuteC(); err != nil {
		os.Exit(1)
	}
}
