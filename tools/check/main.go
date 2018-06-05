package main

// given the path to a schedule config file:
// - read it in
// - check it for validity

import (
	"fmt"
	"log"
	"os"

	"github.com/echohead/stickyshift"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: check $FILE")
	}
	f := os.Args[1]

	if _, err := stickyshift.Read(f); err != nil {
		log.Fatal(f, ": ", err)
	}

	fmt.Printf("%s is ok\n", f)
}
