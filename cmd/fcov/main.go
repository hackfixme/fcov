package main

import (
	"fmt"
	"log"
	"os"

	"github.com/friendlycaptcha/fcov/lib"
	"github.com/friendlycaptcha/fcov/parse"
	"github.com/friendlycaptcha/fcov/summary"
)

func main() {
	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	cov := lib.NewCoverage()

	if err = parse.Go(file, cov); err != nil {
		log.Fatal(err)
	}

	sum := summary.Create(cov)
	fmt.Println(sum.Render(summary.Markdown))
}
