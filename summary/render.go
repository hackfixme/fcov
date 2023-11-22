package summary

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

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
func (s *Summary) Render(ft Format, groupFiles bool, exclude *gitignore.GitIgnore) string {
	if len(s.Packages) == 0 {
		return ""
	}
	pkgNames := make([]string, 0, len(s.Packages))
	for pkgName := range s.Packages {
		if exclude.MatchesPath(pkgName) {
			continue
		}
		pkgNames = append(pkgNames, pkgName)
	}
	sort.Strings(pkgNames)

	out := []string{}
	for _, pkgName := range pkgNames {
		pkgSum := s.Packages[pkgName]
		out = append(out, pkgSum.Render(ft))

		if len(pkgSum.Files) == 0 {
			continue
		}

		fnames := make([]string, 0, len(pkgSum.Files))
		for fname := range pkgSum.Files {
			fnames = append(fnames, fname)
		}
		sort.Strings(fnames)

		for _, fname := range fnames {
			file := pkgSum.Files[fname]
			if exclude.MatchesPath(file.AbsPath()) {
				continue
			}
			out = append(out, file.Render(ft, groupFiles))
		}
	}

	return strings.Join(out, "\n")
}

// Render the package summary as a string in the provided format.
func (p *Package) Render(ft Format) string {
	cov := strconv.FormatFloat(p.Coverage*100, 'f', 2, 64)
	switch ft {
	case Markdown:
		return fmt.Sprintf("%s | %s%%", p.Name, cov)
	default:
		return fmt.Sprintf("%s\t%s%%", p.Name, cov)
	}
}

// Render the file summary as a string in the provided format.
func (f *File) Render(ft Format, group bool) string {
	cov := strconv.FormatFloat(f.Coverage*100, 'f', 2, 64)

	prefix := "- "
	if !group {
		prefix = fmt.Sprintf("%s/", f.Package)
	}
	switch ft {
	case Markdown:
		return fmt.Sprintf("%s%s | %s%%", prefix, f.Name, cov)
	default:
		return fmt.Sprintf("%s%s\t%s%%", prefix, f.Name, cov)
	}
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
