package parse

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	gitignore "github.com/sabhiram/go-gitignore"

	"github.com/friendlycaptcha/fcov/types"
)

type parsedGoLine struct {
	filename string
	block    types.FileBlock
	stats    types.Stats
}

// Go parses a Go coverage file into the provided coverage, applying the
// provided file filter.
func Go(r io.Reader, cov *types.Coverage, filter *gitignore.GitIgnore) error {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		// TODO: Handle modes properly.
		if strings.HasPrefix(line, "mode:") {
			continue
		}
		p, err := parseGoLine(line)
		if err != nil {
			return fmt.Errorf("failed parsing line '%s': %w", line, err)
		}

		if filter.MatchesPath(p.filename) {
			continue
		}

		if _, ok := cov.Files[p.filename]; !ok {
			cov.Files[p.filename] = map[types.FileBlock]*types.Stats{}
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
	_, err := fmt.Sscanf(parts[1], "%d.%d,%d.%d %d %d",
		&p.block.Start.Line, &p.block.Start.Col, &p.block.End.Line,
		&p.block.End.Col, &p.stats.NumStatements, &p.stats.HitCount)
	if err != nil {
		return p, err
	}

	return p, nil
}
