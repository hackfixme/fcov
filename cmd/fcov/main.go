package main

import (
	"os"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"

	"go.hackfix.me/fcov/app"
	aerrors "go.hackfix.me/fcov/app/errors"
)

func main() {
	a, err := app.New("fcov",
		app.WithFDs(
			os.Stdin,
			colorable.NewColorable(os.Stdout),
			colorable.NewColorable(os.Stderr),
		),
		app.WithFS(osfs.New()),
		app.WithLogger(
			isatty.IsTerminal(os.Stdout.Fd()),
			isatty.IsTerminal(os.Stderr.Fd()),
		),
		app.WithEnv(osEnv{}),
	)
	if err != nil {
		aerrors.Errorf(err)
		os.Exit(1)
	}
	if err = a.Run(os.Args[1:]); err != nil {
		aerrors.Errorf(err)
		os.Exit(1)
	}
}

type osEnv struct{}

func (e osEnv) Get(key string) string {
	return os.Getenv(key)
}

func (e osEnv) Set(key, val string) error {
	return os.Setenv(key, val)
}
