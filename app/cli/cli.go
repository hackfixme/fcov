package cli

import (
	"github.com/alecthomas/kong"
	actx "go.hackfix.me/fcov/app/context"
)

// CLI is the command line interface of fcov.
type CLI struct {
	kctx *kong.Context

	Report Report `kong:"cmd,help='Analyze coverage file(s) and create a coverage report.'"`
}

// New initializes the command-line interface.
func New(appCtx *actx.Context, args []string, exitFn func(int)) (*CLI, error) {
	c := &CLI{}
	kparser, err := kong.New(c,
		kong.Name("fcov"),
		kong.UsageOnError(),
		kong.DefaultEnvars("FCOV"),
		kong.Exit(exitFn),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}),
	)
	if err != nil {
		return nil, err
	}

	kparser.Stdout = appCtx.Stdout
	kparser.Stderr = appCtx.Stderr

	kctx, err := kparser.Parse(args)
	if err != nil {
		return nil, err
	}

	c.kctx = kctx

	return c, nil
}

// Execute starts the command execution.
func (c *CLI) Execute(appCtx *actx.Context) error {
	return c.kctx.Run(appCtx)
}
