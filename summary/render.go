package summary

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type Format int

const (
	Text Format = iota + 1
	Markdown
)

func (s *Summary) Render(f Format) string {
	if len(s.Packages) == 0 {
		return ""
	}
	pkgNames := make([]string, 0, len(s.Packages))
	for pkgName := range s.Packages {
		pkgNames = append(pkgNames, pkgName)
	}
	sort.Strings(pkgNames)

	out := []string{}
	for _, pkgName := range pkgNames {
		pkgSum := s.Packages[pkgName]
		out = append(out, pkgSum.Render(f))

		if len(pkgSum.Files) == 0 {
			continue
		}

		fnames := make([]string, 0, len(pkgSum.Files))
		for fname := range pkgSum.Files {
			fnames = append(fnames, fname)
		}
		sort.Strings(fnames)

		for _, fname := range fnames {
			out = append(out, pkgSum.Files[fname].Render(f))
		}
	}

	return strings.Join(out, "\n")
}

func (p *Package) Render(ft Format) string {
	cov := strconv.FormatFloat(p.Coverage*100, 'f', 2, 64)
	switch ft {
	case Markdown:
		return fmt.Sprintf("%s | %s%%", p.Name, cov)
	default:
		return fmt.Sprintf("%s\t%s%%", p.Name, cov)
	}
}

func (f *File) Render(ft Format) string {
	cov := strconv.FormatFloat(f.Coverage*100, 'f', 2, 64)
	switch ft {
	case Markdown:
		return fmt.Sprintf("- %s | %s%%", f.Name, cov)
	default:
		return fmt.Sprintf("- %s\t%s%%", f.Name, cov)
	}
}
