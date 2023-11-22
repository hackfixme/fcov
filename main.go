package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type lineCol struct {
	line, col int
}

type parsedLine struct {
	fname      string
	start, end lineCol
	statements int
	covered    int
}

type cov struct {
	statements int
	covered    int
}

type block struct {
	start, end lineCol
}

func main() {
	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	coverage := make(map[string]map[block]*cov)
	rx := regexp.MustCompile(`^(.+):(\d+)\.(\d+),(\d+)\.(\d+) (\d+) (\d+)$`)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "mode:") {
			continue
		}
		p, err := parseLine(rx, line)
		if err != nil {
			log.Fatal(fmt.Sprintf("failed parsing line '%s': %v", line, err))
		}

		if _, ok := coverage[p.fname]; !ok {
			coverage[p.fname] = map[block]*cov{}
		}

		cov := cov{statements: p.statements, covered: p.covered}
		b := block{start: p.start, end: p.end}
		if c, ok := coverage[p.fname][b]; ok {
			c.covered = max(c.covered, p.covered)
		} else {
			coverage[p.fname][b] = &cov
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
			statements   int
			covered      int
			fileCoverage float64
			f            = coverage[fn]
		)
		for _, cov := range f {
			statements += cov.statements
			if cov.covered > 0 {
				covered += cov.statements
			}
			// covered += cov.covered
		}

		if statements > 0 {
			fileCoverage = float64(covered) / float64(statements)
		}

		fmt.Printf("%s %f\n", fn, fileCoverage)
	}
}

func parseLine(rx *regexp.Regexp, line string) (p parsedLine, err error) {
	match := rx.FindStringSubmatch(line)

	if len(match) != 8 {
		return p, fmt.Errorf("unexpected format")
	}

	p.fname = match[1]
	p.start.line, err = strconv.Atoi(match[2])
	if err != nil {
		return p, fmt.Errorf("failed parsing start line: %w", err)
	}
	p.start.col, err = strconv.Atoi(match[3])
	if err != nil {
		return p, fmt.Errorf("failed parsing start col: %w", err)
	}
	p.end.line, err = strconv.Atoi(match[4])
	if err != nil {
		return p, fmt.Errorf("failed parsing end line: %w", err)
	}
	p.end.col, err = strconv.Atoi(match[5])
	if err != nil {
		return p, fmt.Errorf("failed parsing end col: %w", err)
	}
	p.statements, err = strconv.Atoi(match[6])
	if err != nil {
		return p, fmt.Errorf("failed parsing numStats: %w", err)
	}
	p.covered, err = strconv.Atoi(match[7])
	if err != nil {
		return p, fmt.Errorf("failed parsing count: %w", err)
	}

	return p, nil
}
