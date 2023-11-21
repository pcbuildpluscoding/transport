package main

import (
	"fmt"
	"os"

	"github.com/pcbuildpluscoding/logroll"
	tpt "github.com/pcbuildpluscoding/transport"
)

type ApiError = tpt.ApiError
type ApiNote = tpt.ApiNote

var logger = logroll.Get()

// ---------------------------------------------------------------//
// NewScraperd()
// ---------------------------------------------------------------//
func main() {
	if len(os.Args) != 2 {
		logger.Errorf("Error - required testcase key is undefined")
		os.Exit(1)
	}

	var err error
	tcKey := os.Args[1]

	switch tcKey {
	case "tcA":
		err = tcA()
	case "tcB":
		err = tcB()
	default:
		err = fmt.Errorf("unknown testcase id : %s", tcKey)
	}

	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
}
