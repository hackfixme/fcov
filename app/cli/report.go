package cli

import (
	"bytes"
	"encoding"
	"fmt"
	"math/rand"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/mandelsoft/vfs/pkg/vfs"
	gitignore "github.com/sabhiram/go-gitignore"

	actx "go.hackfix.me/fcov/app/context"
	aerrors "go.hackfix.me/fcov/app/errors"
	"go.hackfix.me/fcov/parse"
	"go.hackfix.me/fcov/report"
	"go.hackfix.me/fcov/types"
)

// Report is the fcov report command.
type Report struct {
	Paths             []string         `arg:"" help:"One or more paths to coverage files or directories."` // not using 'existingfile' modifier since it makes it difficult to test with an in-memory FS
	Merge             bool             `help:"If true, coverage paths will be calculated as one. Otherwise multiple reports will be generated, one for each path. "`
	Filter            []string         `help:"Glob patterns applied on file paths to filter files from the coverage calculation and output. \n Example: '*,!*pkg*' would exclude all files except those that contain 'pkg'. " placeholder:"<glob pattern>"`
	FilterOutput      []string         `help:"Glob patterns applied on file paths to filter files from the output, but *not* from the coverage calculation. " placeholder:"<glob pattern>"`
	FilterOutputFile  string           `help:"Path to a file that contains newline-separated file paths to include in the output.\nIf specified, it overrides --filter-output. " placeholder:"<path>"`
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
// TODO: This currently assumes Go coverage processing. Either correctly infer so,
// or add a CLI flag to use Go mode reporting.
func (s *Report) Run(appCtx *actx.Context) error {
	filterCov := gitignore.CompileIgnoreLines(s.Filter...)

	filterOutLines := s.FilterOutput
	if s.FilterOutputFile != "" {
		file, err := appCtx.FS.Open(s.FilterOutputFile)
		if err != nil {
			return aerrors.NewRuntimeError("failed opening filter output file", err, "")
		}
		defer file.Close()

		// TODO: Merge with s.FilterOutput instead of overriding it?
		if len(filterOutLines) > 0 {
			appCtx.Logger.Warn("--filter-output-file overrides --filter-output")
		}

		filterOutLines, err = createOutputFilterFromFile(file)
		if err != nil {
			return aerrors.NewRuntimeError("failed reading filter output file", err, "")
		}
	}
	filterOut := gitignore.CompileIgnoreLines(filterOutLines...)

	var covFiles []string
	for _, path := range s.Paths {
		info, err := appCtx.FS.Stat(path)
		if err != nil {
			if vfs.IsNotExist(err) {
				return err
			}
			return aerrors.NewRuntimeError(fmt.Sprintf("failed getting information on %s", path), err, "")
		}

		if info.IsDir() {
			covDirs, err := findCoverageDirectories(appCtx.FS, path)
			if err != nil {
				return aerrors.NewRuntimeError(fmt.Sprintf("failed reading directory %s", path), err, "")
			}

			if len(covDirs) == 0 {
				appCtx.Logger.Warn(fmt.Sprintf("No coverage directories found in %s", path))
			}

			genCovFiles, err := generateTextCoverage(covDirs)
			if err != nil {
				return err
			}
			covFiles = append(covFiles, genCovFiles...)
		} else {
			covFiles = append(covFiles, path)
		}
	}

	covs := map[string]*types.Coverage{}

	// Random key used to distinguish whether the coverage report should be merged or not.
	// A hackish way of doing this, but the alternative would require more code.
	mergeKey := fmt.Sprintf("%x\x00", rand.Int31())

	for _, covFile := range covFiles {
		file, err := appCtx.FS.Open(covFile)
		if err != nil {
			return err
		}
		defer file.Close()

		covKey := mergeKey
		if !s.Merge {
			covKey = strings.TrimSuffix(filepath.Base(covFile), filepath.Ext(covFile))
		}

		covs[covKey] = types.NewCoverage()
		if err := parse.Go(file, covs[covKey], filterCov); err != nil {
			return err
		}
	}

	reports := make(map[string]*report.Report, len(covs))
	for name, cov := range covs {
		reports[name] = report.Create(cov)
	}

	renders := make(map[string]map[report.Format]string)
	for _, out := range s.Output {
		var (
			render string
			ok     bool
		)
		for name, r := range reports {
			if _, ok = renders[name]; !ok {
				renders[name] = make(map[report.Format]string)
			}
			render = r.Render(
				out.Format, s.NestFiles, filterOut, s.TrimPackagePrefix,
			)
			renders[name][out.Format] = render
		}

		names := make([]string, 0, len(renders))
		for n := range renders {
			names = append(names, n)
		}
		sort.Strings(names)

		var output bytes.Buffer
		for j, name := range names {
			if name != mergeKey {
				header := reports[name].RenderHeader(
					out.Format, name, j > 0, s.Thresholds.Lower, s.Thresholds.Upper,
				)
				if _, err := fmt.Fprintln(&output, header); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprintln(&output, renders[name][out.Format]); err != nil {
				return err
			}
		}

		if out.Filename == "" {
			if _, err := fmt.Fprintln(appCtx.Stdout, output.String()); err != nil {
				return err
			}
			continue
		}

		if err := vfs.WriteFile(appCtx.FS, out.Filename, output.Bytes(), 0o644); err != nil {
			return err
		}
	}

	return nil
}
