package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
)

func main() {
	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	coverage := make(map[string]map[block]*cov)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "mode:") {
			continue
		}
		p, err := parseLine(line)
		if err != nil {
			log.Fatal(fmt.Sprintf("failed parsing line '%s': %v", line, err))
		}

		if _, ok := coverage[p.filename]; !ok {
			coverage[p.filename] = map[block]*cov{}
		}

		cov := cov{numStatements: p.numStatements, hitCount: p.hitCount}
		if c, ok := coverage[p.filename][p.block]; ok {
			c.hitCount += p.hitCount
		} else {
			coverage[p.filename][p.block] = &cov
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	fnames := make([]string, 0, len(coverage))
	for fn := range coverage {
		fnames = append(fnames, fn)
	}
	sort.Strings(fnames)

	for _, fn := range fnames {
		var (
			numStatements int
			hitCount      int
			fileCoverage  float64
			f             = coverage[fn]
		)
		for _, cov := range f {
			numStatements += cov.numStatements
			if cov.hitCount > 0 {
				hitCount += cov.numStatements
			}
			// covered += cov.covered
		}

		if numStatements > 0 {
			fileCoverage = float64(hitCount) / float64(numStatements)
		}

		fmt.Printf("%s %f\n", fn, fileCoverage)
	}
}
