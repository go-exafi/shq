package main

import (
	"fmt"
	"os"

	"github.com/go-exafi/shq"
)

func main() {
	spc := ""
	for i := 1; i < len(os.Args); i++ {
		fmt.Printf("%s%s", spc, shq.Arg(os.Args[i]))
		spc = " "
	}
}
