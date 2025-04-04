package report

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/olekukonko/tablewriter"
	gitignore "github.com/sabhiram/go-gitignore"
)

// Format is the type of format a report can be rendered in.
type Format string

// Supported format types.
const (
	Text     Format = "txt"
	Markdown Format = "md"
)

// Marker used to distinguish package from file paths in the pre-rendered output.
const pkgMarker = '\x00'

// Render the report as a string in the provided format, applying the provided
// filter, and style adjustments.
// The lower and upper coverage thresholds are used by formats like Markdown to
// apply different colors depending on their values. trimPackagePrefix will
// remove the matching prefix from the absolute file path.
func (s *Report) Render(
	ft Format, nestFiles bool, filter *gitignore.GitIgnore,
	lowerThreshold, upperThreshold float64, trimPackagePrefix string,
) string {
	if len(s.Packages) == 0 {
		return ""
	}

	sum := s.preRender(filter, nestFiles, trimPackagePrefix)

	buf := &strings.Builder{}
	table := tablewriter.NewWriter(buf)
	data := [][]string{}

	switch ft {
	case Text:
		table.SetColumnSeparator("")
		table.SetNoWhiteSpace(true)
		table.SetBorder(false)
		if nestFiles {
			renderTextNested(sum, &data)
		} else {
			data = sum
		}
	case Markdown:
		table.SetCenterSeparator("|")
		table.SetAutoFormatHeaders(false)
		table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
		table.SetCenterSeparator("|")

		buf.Write([]byte(fmt.Sprintf("![Total Coverage](%s)\n\n",
			generateBadgeURL(s.Coverage*100, lowerThreshold, upperThreshold))))

		if len(sum) == 0 {
			break
		}

		// Set the headers manually instead of using table.SetHeader because it
		// doesn't support GitHub's column alignment syntax.
		// See https://github.com/olekukonko/tablewriter/pull/181
		data = append(data, []string{"Package", "Coverage"},
			[]string{":------", "-------:"})

		if nestFiles {
			renderMarkdownNested(sum, &data)
		} else {
			renderMarkdown(sum, &data)
		}
	}

	table.SetAutoWrapText(false)
	table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_RIGHT})
	table.SetTablePadding(" ")
	table.AppendBulk(data)
	table.Render()

	if ft == Text {
		buf.Write([]byte(fmt.Sprintf("\nTotal Coverage: %.2f%%", s.Coverage*100)))
	}

	out := buf.String()
	// tablewriter appends an extra newline at the end that I can't seem to
	// disable, so remove it.
	out, _ = strings.CutSuffix(out, "\n")

	return out
}

// preRender sorts and flattens the report, applying any filters, and
// optionally trimming the file paths as needed.
func (s *Report) preRender(filter *gitignore.GitIgnore, nestFiles bool, trimPackagePrefix string) [][]string {
	pkgNames := make([]string, 0, len(s.Packages))
	for pkgName := range s.Packages {
		pkgNames = append(pkgNames, pkgName)
	}
	sort.Strings(pkgNames)

	sum := make([][]string, 0)
	for _, pkgName := range pkgNames {
		pkgSum := s.Packages[pkgName]

		fnames := make([]string, 0, len(pkgSum.Files))
		for fname := range pkgSum.Files {
			fnames = append(fnames, fname)
		}
		sort.Strings(fnames)

		pkgFiles := make([][]string, 0, len(pkgSum.Files))
		for _, fname := range fnames {
			file := pkgSum.Files[fname]
			absPath := file.AbsPath()
			if filter.MatchesPath(absPath) {
				continue
			}
			if !nestFiles {
				fname = strings.TrimPrefix(absPath, trimPackagePrefix)
			}
			fileCov := strconv.FormatFloat(file.Coverage*100, 'f', 2, 64)
			pkgFiles = append(pkgFiles, []string{fname, fmt.Sprintf("%s%%", fileCov)})
		}

		if !filter.MatchesPath(pkgName) || (nestFiles && len(pkgFiles) > 0) {
			pkgName = strings.TrimPrefix(pkgName, trimPackagePrefix)

			// HACK: Mark package lines with a prefix marker, so that they can
			// be distinguished during final rendering. Otherwise the sum data
			// structure would have to be more complicated.
			pkgName = string(pkgMarker) + pkgName
			sum = append(sum, []string{
				pkgName,
				fmt.Sprintf("%s%%", strconv.FormatFloat(pkgSum.Coverage*100, 'f', 2, 64)),
			})
		}
		sum = append(sum, pkgFiles...)
	}

	return sum
}

func renderMarkdown(sum [][]string, data *[][]string) {
	for _, line := range sum {
		// HACK: Package lines are distinguished by a prefix marker. Otherwise
		// the sum data structure would have to be more complicated.
		if line[0][0] == pkgMarker {
			line[0] = line[0][1:]
		}
		line[0] = fmt.Sprintf("`%s`", line[0])
		*data = append(*data, []string{line[0], line[1]})
	}
}

func renderMarkdownNested(sum [][]string, data *[][]string) {
	pkgDataTmpl := "<details><summary>`%s`</summary>%s</details>"
	tableTmpl := "<table>{{range .}}<tr><td>`{{index . 0}}`</td>" +
		"<td>{{index . 1}}</td></tr>{{end}}" +
		"</table>"
	tmpl := template.Must(template.New("table").Parse(tableTmpl))

	var (
		pkgName, pkgCov string
		files           [][]string
	)
	for i, line := range sum {
		// HACK: Package lines are distinguished by a prefix marker. Otherwise
		// the sum data structure would have to be more complicated.
		if line[0][0] == pkgMarker {
			pkgName = line[0][1:]
			pkgCov = line[1]
		} else {
			files = append(files, line)
		}

		// If we reached the end, or the next line is a different package, that
		// means we're done with the current one, so render it.
		if i == len(sum)-1 || (i+1 < len(sum) && sum[i+1][0][0] == pkgMarker) {
			var fileData bytes.Buffer
			if err := tmpl.Execute(&fileData, files); err != nil {
				panic(err)
			}
			pkgData := fmt.Sprintf(pkgDataTmpl, pkgName, fileData.String())
			*data = append(*data, []string{pkgData, pkgCov})
			files = [][]string{}
		}
	}
}

func renderTextNested(sum [][]string, data *[][]string) {
	for _, line := range sum {
		// HACK: Package lines are distinguished by a prefix marker. Otherwise
		// the sum data structure would have to be more complicated.
		if line[0][0] == pkgMarker {
			line[0] = line[0][1:]
		} else {
			line[0] = fmt.Sprintf("    %s", line[0])
		}
		*data = append(*data, []string{line[0], line[1]})
	}
}

// FormatFromString parses s into a valid Format value.
func FormatFromString(s string) Format {
	switch Format(s) {
	case Text:
		return Text
	case Markdown:
		return Markdown
	default:
		return ""
	}
}

func generateBadgeURL(cov float64, lowerThreshold, upperThreshold float64) string {
	color := "success"
	if cov < lowerThreshold {
		color = "critical"
	} else if cov < upperThreshold {
		color = "yellow"
	}

	return fmt.Sprintf("https://img.shields.io/badge/Total%%20Coverage-%.2f%%25-%s?style=flat", cov, color)
}
