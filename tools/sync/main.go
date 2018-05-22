package main

// given the path to a schedule config file:
// - read it in
// - check it for validity
// - apply it to pagerduty

import (
	"log"
	"os"

	"github.com/echohead/stickyshift"
	"github.com/echohead/stickyshift/pagerduty"
)

var ()

func fatalIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: PD_TOKEN='***' sync $FILE")
	}
	f := os.Args[1]

	s, err := stickyshift.Read(f)
	fatalIfErr(err)

	c, err := pagerduty.New()
	fatalIfErr(err)

	err = c.Sync(s.Id, s.Shifts)
	fatalIfErr(err)
}
