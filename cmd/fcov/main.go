package main

import (
	"os"

	"github.com/mandelsoft/vfs/pkg/osfs"

	"github.com/friendlycaptcha/fcov/app"
)

func main() {
	app.New(
		app.WithExit(os.Exit),
		app.WithArgs(os.Args[1:]),
		app.WithEnv(osEnv{}),
		app.WithFDs(os.Stdin, os.Stdout, os.Stderr),
		app.WithFS(osfs.New()),
	).Run()
}

type osEnv struct{}

func (e osEnv) Get(key string) string {
	return os.Getenv(key)
}

func (e osEnv) Set(key, val string) error {
	return os.Setenv(key, val)
}
