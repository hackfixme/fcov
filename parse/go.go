package parse

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/friendlycaptcha/fcov/lib"
)

type parsedGoLine struct {
	filename string
	block    lib.FileBlock
	stats    lib.Stats
}

// Go parses a Go coverage file into the provided coverage.
func Go(r io.Reader, cov *lib.Coverage) error {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		// TODO: Handle modes properly.
		if strings.HasPrefix(line, "mode:") {
			continue
		}
		p, err := parseGoLine(line)
		if err != nil {
			return fmt.Errorf("failed parsing line '%s': %v", line, err)
		}

		if _, ok := cov.Files[p.filename]; !ok {
			cov.Files[p.filename] = map[lib.FileBlock]*lib.Stats{}
		}

		if s, ok := cov.Files[p.filename][p.block]; ok {
			s.HitCount += p.stats.HitCount
		} else {
			cov.Files[p.filename][p.block] = &p.stats
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed scanning input: %w", err)
	}

	return nil
}

// Parse a Go coverage line in the format:
// <filename.go>:<startLine>.<startColumn>,<endLine>.<endColumn> <numberOfStatements> <hitCount>
// See https://github.com/golang/go/blob/go1.21.1/src/cmd/vendor/golang.org/x/tools/cover/profile.go#L58
func parseGoLine(line string) (parsedGoLine, error) {
	var p parsedGoLine
	parts := strings.Split(line, ":")
	if len(parts) != 2 {
		return p, fmt.Errorf("wrong format")
	}

	p.filename = parts[0]
	n, err := fmt.Sscanf(parts[1], "%d.%d,%d.%d %d %d",
		&p.block.Start.Line, &p.block.Start.Col, &p.block.End.Line,
		&p.block.End.Col, &p.stats.NumStatements, &p.stats.HitCount)
	if n != 6 {
		return p, fmt.Errorf("wrong format")
	}
	if err != nil {
		return p, err
	}

	return p, nil
}
