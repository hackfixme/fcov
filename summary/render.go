package summary

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	gitignore "github.com/sabhiram/go-gitignore"
)

// Format is the type of format a summary can be rendered in.
type Format string

// Supported format types.
const (
	Text     Format = "txt"
	Markdown Format = "md"
)

// Render the summary as a string in the provided format.
func (s *Summary) Render(
	ft Format, groupFiles bool, exclude *gitignore.GitIgnore,
	lowerThreshold, upperThreshold float64, trimPackagePrefix string,
) string {
	if len(s.Packages) == 0 {
		return ""
	}

	buf := &strings.Builder{}
	table := tablewriter.NewWriter(buf)
	var data [][]string

	if ft == Markdown {
		buf.Write([]byte(fmt.Sprintf("![Total Coverage](%s)\n\n",
			generateBadgeURL(s.Coverage*100, lowerThreshold, upperThreshold))))
	} else {
		table.SetColumnSeparator("")
	}

	pkgNames := make([]string, 0, len(s.Packages))
	for pkgName := range s.Packages {
		pkgNames = append(pkgNames, pkgName)
	}
	sort.Strings(pkgNames)

	for _, pkgName := range pkgNames {
		pkgSum := s.Packages[pkgName]

		pkgName = strings.TrimPrefix(pkgName, trimPackagePrefix)

		fnames := make([]string, 0, len(pkgSum.Files))
		for fname := range pkgSum.Files {
			fnames = append(fnames, fname)
		}
		sort.Strings(fnames)

		filePrefix := "    "
		if !groupFiles {
			filePrefix = fmt.Sprintf("%s/", pkgName)
		}
		pkgFiles := make([][]string, 0, len(pkgSum.Files))
		for _, fname := range fnames {
			file := pkgSum.Files[fname]
			if exclude.MatchesPath(file.AbsPath()) {
				continue
			}

			fileCov := strconv.FormatFloat(file.Coverage*100, 'f', 2, 64)
			pkgFiles = append(pkgFiles, []string{fmt.Sprintf("%s%s", filePrefix, fname),
				fmt.Sprintf("%s%%", fileCov)})
		}

		if len(pkgFiles) > 0 {
			pkgCov := strconv.FormatFloat(pkgSum.Coverage*100, 'f', 2, 64)
			data = append(data, []string{pkgName, fmt.Sprintf("%s%%", pkgCov)})
			data = append(data, pkgFiles...)
		}
	}

	table.SetAutoWrapText(false)
	table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_RIGHT})
	table.SetBorder(false)
	table.SetNoWhiteSpace(true)
	table.SetTablePadding(" ")
	table.AppendBulk(data)
	table.Render()

	if ft == Text {
		buf.Write([]byte(fmt.Sprintf("\nTotal Coverage: %.2f%%", s.Coverage*100)))
	}

	return buf.String()
}

// FormatFromString parses s into a valid Format value.
func FormatFromString(s string) Format {
	switch Format(s) {
	case Markdown:
		return Markdown
	case Text:
		return Text
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
