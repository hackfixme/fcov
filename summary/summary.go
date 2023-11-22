package summary

import (
	"path"

	"github.com/friendlycaptcha/fcov/lib"
)

// Package holds coverage information related to a package.
type Package struct {
	lib.Stats
	Name  string
	Files map[string]*File
}

// File holds coverage information related to a file.
type File struct {
	lib.Stats
	Name    string
	Package string
}

// Summary holds global coverage information.
type Summary struct {
	lib.Stats
	Packages map[string]*Package
}

// Create a new summary based on the provided coverage.
func Create(cov *lib.Coverage) *Summary {
	sum := &Summary{}

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
		if pkgSum.NumStatements > 0 {
			pkgSum.Coverage = float64(pkgSum.HitCount) / float64(pkgSum.NumStatements)
		}
	}

	if sum.NumStatements > 0 {
		sum.Coverage = float64(sum.HitCount) / float64(sum.NumStatements)
	}

	return sum
}
