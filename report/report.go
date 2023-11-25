package report

import (
	"path"

	"github.com/friendlycaptcha/fcov/types"
)

// Package holds coverage information related to a package.
type Package struct {
	types.Stats
	Name  string
	Files map[string]*File
}

// File holds coverage information related to a file.
type File struct {
	types.Stats
	Name    string
	Package string
}

// AbsPath returns the absolute path of the file.
func (f File) AbsPath() string {
	return path.Join(f.Package, f.Name)
}

// Report holds global coverage information.
type Report struct {
	types.Stats
	Packages map[string]*Package
}

// Create a new report based on the provided coverage.
func Create(cov *types.Coverage) *Report {
	sum := &Report{}

	for filename, fileBlocks := range cov.Files {
		var (
			numStatements int
			hitCount      int
		)
		for _, stat := range fileBlocks {
			numStatements += stat.NumStatements
			if stat.HitCount > 0 {
				hitCount += stat.NumStatements
			}
		}

		var (
			pkg    = path.Dir(filename)
			pkgSum *Package
			ok     bool
		)
		if pkgSum, ok = sum.Packages[pkg]; !ok {
			if len(sum.Packages) == 0 {
				sum.Packages = make(map[string]*Package)
			}
			pkgSum = &Package{Name: pkg}
			sum.Packages[pkg] = pkgSum
		}

		pkgSum.NumStatements += numStatements
		pkgSum.HitCount += hitCount

		var fileSum *File
		if fileSum, ok = sum.Packages[pkg].Files[filename]; !ok {
			if len(sum.Packages[pkg].Files) == 0 {
				sum.Packages[pkg].Files = make(map[string]*File)
			}
			fileSum = &File{Name: path.Base(filename), Package: pkg}
			sum.Packages[pkg].Files[fileSum.Name] = fileSum
		}

		fileSum.NumStatements = numStatements
		fileSum.HitCount = hitCount
		if numStatements > 0 {
			fileSum.Coverage = float64(hitCount) / float64(numStatements)
		}
	}

	for _, pkgSum := range sum.Packages {
		sum.NumStatements += pkgSum.NumStatements
		sum.HitCount += pkgSum.HitCount
		if pkgSum.NumStatements > 0 {
			pkgSum.Coverage = float64(pkgSum.HitCount) / float64(pkgSum.NumStatements)
		}
	}

	if sum.NumStatements > 0 {
		sum.Coverage = float64(sum.HitCount) / float64(sum.NumStatements)
	}

	return sum
}
