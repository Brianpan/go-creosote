package main

import (
	"fmt"
	"os"

	"github.com/brianpan/go-creosote/scanner"
	_ "github.com/go-python/gpython/modules"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("require folder path to search with")
		return
	}
	dir := os.Args[1]

	rs, err := scanner.ScanAll(dir)
	if err != nil {
		fmt.Println("err, ", err)
		os.Exit(1)
	}
	for _, r := range rs {
		fmt.Println(r)
	}

	return
}
