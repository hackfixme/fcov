package main

import (
	"encoding"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/friendlycaptcha/fcov/lib"
	"github.com/friendlycaptcha/fcov/parse"
	"github.com/friendlycaptcha/fcov/summary"
)

// Summary is the fcov summary command.
type Summary struct {
	Files      []string     `arg:"" help:"One or more coverage files." type:"existingfile"`
	GroupFiles bool         `help:"Group files under packages when rendering to text or Markdown. " default:"true" negatable:""`
	Output     OutputOption `short:"o" help:"Write the summary to stdout or a file. More than one value can be provided, separated by comma.\nValues can be either formats ('md' or 'txt'), or filenames whose formats will be inferred by their extension.\n Example: 'txt,summary.md' would write the summary in text format to stdout, and to a summary.md file in Markdown format. " default:"txt"`
}

// Output is a destination the summary should be written to. If Filename is
// empty, the summary will be written to stdout.
type Output struct {
	Format   summary.Format
	Filename string
}

// OutputOption is a custom type that parses the output option.
type OutputOption []Output

var _ encoding.TextUnmarshaler = &OutputOption{}

// UnmarshalText implements the encoding.TextUnmarshaler interface for
// OutputOption.
func (o *OutputOption) UnmarshalText(text []byte) error {
	options := strings.Split(string(text), ",")

	for _, option := range options {
		out := Output{}
		format := summary.FormatFromString(option)
		if format == "" {
			// Assume it's a filename, and infer the format from the extension.
			ext := filepath.Ext(option)
			if ext == "" {
				return fmt.Errorf("invalid output value: %s", option)
			}
			format = summary.FormatFromString(ext[1:]) // Remove the leading dot
			if format == "" {
				return fmt.Errorf("invalid output format: %s", ext[1:])
			}
			out.Filename = option
		}

		out.Format = format

		*o = append(*o, out)
	}

	return nil
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

	renders := make(map[summary.Format]string)
	for _, out := range s.Output {
		var (
			render string
			ok     bool
		)
		if render, ok = renders[out.Format]; !ok {
			render = sum.Render(out.Format, s.GroupFiles)
			renders[out.Format] = render
		}

		if out.Filename != "" {
			err := ioutil.WriteFile(out.Filename, []byte(render), 0644)
			if err != nil {
				return err
			}
		} else {
			fmt.Println(render)
		}
	}

	return nil
}
