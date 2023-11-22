package main

import (
	"fmt"
	"strings"
)

type lineCol struct {
	line, col int
}

type block struct {
	start, end lineCol
}

type parsedLine struct {
	filename      string
	block         block
	numStatements int
	hitCount      int
}

type cov struct {
	numStatements int
	hitCount      int
}

// Parse a coverage line in the format:
// <filename.go>:<startLine>.<startColumn>,<endLine>.<endColumn> <numberOfStatements> <hitCount>
// See https://github.com/golang/go/blob/go1.21.1/src/cmd/vendor/golang.org/x/tools/cover/profile.go#L58
func parseLine(line string) (parsedLine, error) {
	var p parsedLine
	parts := strings.Split(line, ":")
	if len(parts) != 2 {
		return p, fmt.Errorf("wrong format")
	}

	p.filename = parts[0]
	n, err := fmt.Sscanf(parts[1], "%d.%d,%d.%d %d %d",
		&p.block.start.line, &p.block.start.col, &p.block.end.line,
		&p.block.end.col, &p.numStatements, &p.hitCount)
	if n != 6 {
		return p, fmt.Errorf("wrong format")
	}
	if err != nil {
		return p, err
	}

	return p, nil
}
