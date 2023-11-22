package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
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

	sum := createSummary(coverage)
	fmt.Println(sum.Render(markdown))
}
