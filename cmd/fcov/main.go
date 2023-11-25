package main

import (
	"github.com/alecthomas/kong"
)

// CLI is the command line interface of fcov.
type CLI struct {
	Report Report `kong:"cmd,help='Analyze coverage file(s) and create a coverage report.'"`
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli,
		kong.Name("fcov"),
		kong.UsageOnError(),
		kong.DefaultEnvars("FCOV"),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}),
	)
	ctx.FatalIfErrorf(ctx.Run())
}
