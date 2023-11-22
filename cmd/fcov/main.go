package main

import (
	"github.com/alecthomas/kong"
)

// CLI is the command line interface of fcov.
type CLI struct {
	Summary Summary `kong:"cmd,help='Generate a coverage summary.'"`
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli,
		kong.Name("fcov"),
		kong.UsageOnError(),
		kong.DefaultEnvars("FCOV"),
	)
	ctx.FatalIfErrorf(ctx.Run())
}
