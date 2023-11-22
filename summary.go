package main

import (
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"
)

type commonSummary struct {
	numStatements int
	hitCount      int
	coverage      float64
}

type packageSummary struct {
	commonSummary
	name  string
	files map[string]*fileSummary
}

type fileSummary struct {
	name string
	commonSummary
}

type summary struct {
	commonSummary
	packages map[string]*packageSummary
}

func createSummary(coverage map[string]map[block]*cov) *summary {
	sum := &summary{}

	for filename, fileBlocks := range coverage {
		var (
			numStatements int
			hitCount      int
		)
		for _, cov := range fileBlocks {
			numStatements += cov.numStatements
			if cov.hitCount > 0 {
				hitCount += cov.numStatements
			}
		}

		var (
			pkg    = path.Dir(filename)
			pkgSum *packageSummary
			ok     bool
		)
		if pkgSum, ok = sum.packages[pkg]; !ok {
			if len(sum.packages) == 0 {
				sum.packages = make(map[string]*packageSummary)
			}
			pkgSum = &packageSummary{name: pkg}
			sum.packages[pkg] = pkgSum
		}

		pkgSum.numStatements += numStatements
		pkgSum.hitCount += hitCount

		var fileSum *fileSummary
		if fileSum, ok = sum.packages[pkg].files[filename]; !ok {
			if len(sum.packages[pkg].files) == 0 {
				sum.packages[pkg].files = make(map[string]*fileSummary)
			}
			fileSum = &fileSummary{name: path.Base(filename)}
			sum.packages[pkg].files[fileSum.name] = fileSum
		}

		fileSum.numStatements = numStatements
		fileSum.hitCount = hitCount
		if numStatements > 0 {
			fileSum.coverage = float64(hitCount) / float64(numStatements)
		}
	}

	for _, pkgSum := range sum.packages {
		if pkgSum.numStatements > 0 {
			pkgSum.coverage = float64(pkgSum.hitCount) / float64(pkgSum.numStatements)
		}
	}

	if sum.numStatements > 0 {
		sum.coverage = float64(sum.hitCount) / float64(sum.numStatements)
	}

	return sum
}

type summaryFormat int

const (
	text summaryFormat = iota + 1
	markdown
)

func (s *summary) Render(f summaryFormat) string {
	if len(s.packages) == 0 {
		return ""
	}
	pkgNames := make([]string, 0, len(s.packages))
	for pkgName := range s.packages {
		pkgNames = append(pkgNames, pkgName)
	}
	sort.Strings(pkgNames)

	out := []string{}
	for _, pkgName := range pkgNames {
		pkgSum := s.packages[pkgName]
		out = append(out, pkgSum.Render(f))

		if len(pkgSum.files) == 0 {
			continue
		}

		fnames := make([]string, 0, len(pkgSum.files))
		for fname := range pkgSum.files {
			fnames = append(fnames, fname)
		}
		sort.Strings(fnames)

		for _, fname := range fnames {
			out = append(out, pkgSum.files[fname].Render(f))
		}
	}

	return strings.Join(out, "\n")
}

func (s *packageSummary) Render(f summaryFormat) string {
	cov := strconv.FormatFloat(s.coverage*100, 'f', 2, 64)
	switch f {
	case markdown:
		return fmt.Sprintf("%s | %s%%", s.name, cov)
	default:
		return fmt.Sprintf("%s\t%s%%", s.name, cov)
	}
}

func (s *fileSummary) Render(f summaryFormat) string {
	cov := strconv.FormatFloat(s.coverage*100, 'f', 2, 64)
	switch f {
	case markdown:
		return fmt.Sprintf("- %s | %s%%", s.name, cov)
	default:
		return fmt.Sprintf("- %s\t%s%%", s.name, cov)
	}
}
