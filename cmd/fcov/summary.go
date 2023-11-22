package main

import (
	"fmt"
	"os"

	"github.com/friendlycaptcha/fcov/lib"
	"github.com/friendlycaptcha/fcov/parse"
	"github.com/friendlycaptcha/fcov/summary"
)

// Summary is the fcov summary command.
type Summary struct {
	Files      []string `arg:"name='filepath',required,min=1,help='One or more coverage files.'"`
	GroupFiles bool     `help:"Group files under packages when rendering to text or Markdown." default:"true" negatable:""`
}

// Run the fcov summary command.
func (s *Summary) Run() error {
	cov := lib.NewCoverage()

	for _, fpath := range s.Files {
		file, err := os.Open(fpath)
		if err != nil {
			return err
		}
		defer file.Close()

		if err = parse.Go(file, cov); err != nil {
			return err
		}
	}

	sum := summary.Create(cov)
	fmt.Println(sum.Render(summary.Markdown, s.GroupFiles))

	return nil
}
