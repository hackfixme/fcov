package cli

import (
	"encoding"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mandelsoft/vfs/pkg/vfs"
	gitignore "github.com/sabhiram/go-gitignore"

	actx "github.com/friendlycaptcha/fcov/app/context"
	"github.com/friendlycaptcha/fcov/parse"
	"github.com/friendlycaptcha/fcov/report"
	"github.com/friendlycaptcha/fcov/types"
)

// Report is the fcov report command.
type Report struct {
	Files             []string         `arg:"" help:"One or more coverage files." type:"existingfile"`
	Filter            []string         `help:"Glob patterns applied on file paths to filter files from the coverage calculation and output. \n Example: '*,!*pkg*' would exclude all files except those that contain 'pkg'. " placeholder:"<glob pattern>"`
	FilterOutput      []string         `help:"Glob patterns applied on file paths to filter files from the output, but *not* from the coverage calculation. " placeholder:"<glob pattern>"`
	NestFiles         bool             `help:"Nest files under packages when rendering to text or Markdown. " default:"true" negatable:""`
	Output            OutputOption     `short:"o" help:"Write the report to stdout or a file. More than one value can be provided, separated by comma.\nValues can either be formats ('txt' or 'md'), or filenames whose formats will be inferred by their extension.\n Example: 'txt,report.md' would write the report in text format to stdout, and to a report.md file in Markdown format. " default:"txt"`
	Thresholds        ThresholdsOption `help:"Lower and upper threshold percentages for badge and health indicators. " default:"50,75"`
	TrimPackagePrefix string           `help:"Trim this prefix string from the package path in the output. "`
}

// Output is a destination the report should be written to. If Filename is
// empty, the report will be written to stdout.
type Output struct {
	Format   report.Format
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
		format := report.FormatFromString(option)
		if format == "" {
			// Assume it's a filename, and infer the format from the extension.
			ext := filepath.Ext(option)
			if ext == "" {
				return fmt.Errorf("invalid output value: %s", option)
			}
			format = report.FormatFromString(ext[1:]) // Remove the leading dot
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

// ThresholdsOption is a custom type that parses the thresholds option.
type ThresholdsOption struct {
	Lower, Upper float64
}

var _ encoding.TextUnmarshaler = &ThresholdsOption{}

// UnmarshalText implements the encoding.TextUnmarshaler interface for
// ThresholdsOption.
func (o *ThresholdsOption) UnmarshalText(text []byte) error {
	parts := strings.Split(string(text), ",")
	if len(parts) != 2 {
		return fmt.Errorf("invalid thresholds value: %s", text)
	}

	var err error
	if o.Lower, err = strconv.ParseFloat(parts[0], 64); err != nil {
		return fmt.Errorf("invalid lower threshold '%s': %w", parts[0], err)
	}
	if o.Upper, err = strconv.ParseFloat(parts[1], 64); err != nil {
		return fmt.Errorf("invalid upper threshold '%s': %w", parts[1], err)
	}

	return nil
}

// Run the fcov report command.
func (s *Report) Run(appCtx *actx.Context) error {
	cov := types.NewCoverage()
	filterCov := gitignore.CompileIgnoreLines(s.Filter...)
	filterOut := gitignore.CompileIgnoreLines(s.FilterOutput...)

	for _, fpath := range s.Files {
		file, err := appCtx.FS.Open(fpath)
		if err != nil {
			return err
		}
		defer file.Close()

		if err = parse.Go(file, cov, filterCov); err != nil {
			return err
		}
	}

	sum := report.Create(cov)

	renders := make(map[report.Format]string)
	for _, out := range s.Output {
		var (
			render string
			ok     bool
		)
		if render, ok = renders[out.Format]; !ok {
			render = sum.Render(
				out.Format, s.NestFiles, filterOut, s.Thresholds.Lower,
				s.Thresholds.Upper, s.TrimPackagePrefix)
			renders[out.Format] = render
		}

		if out.Filename == "" {
			if _, err := fmt.Fprintln(appCtx.Stdout, render); err != nil {
				return err
			}
			continue
		}

		if err := vfs.WriteFile(appCtx.FS, out.Filename, []byte(render), 0644); err != nil {
			return err
		}
	}

	return nil
}
