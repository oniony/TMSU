package main

import (
	"fmt"
	"flag"
)

type Command interface {
	Execute()
}

func main() {
	flag.Parse()

	for i := 0; i < flag.NArg(); i += 1 {
		fmt.Println(flag.Args()[i])
	}
}
