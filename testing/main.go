package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/pcbuildpluscoding/logroll"
	tpt "github.com/pcbuildpluscoding/transport"
)

type ApiError = tpt.ApiError
type ApiNote = tpt.ApiNote

func checkError(errCh chan ApiNote) {
  logger.Infof("checkError is running ...")
  rcd, _ := <-errCh
  logger.Info("@@@@@@@@@@@@@@@ !! got a record !! @@@@@@@@@@@@@@@")
  err := rcd.Unwrap()
  if errors.Is(err, io.EOF) {
    apiErr := ApiError{}
    if errors.As(err, &apiErr) {
      logger.Infof("got ApiError with appKey : %s", apiErr.AppKey())
    } else {
      logger.Errorf("ApiError not provided, appKey is undefined, aborting reconnect ...")
    }
  }
  logger.Infof("checkError is done")
}

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
