package main

import (
	"os"

	"github.com/mklinovsky/gira/internal/cli"
)

var version = "dev"

func main() {
	os.Exit(cli.Run(os.Args[1:], version))
}
