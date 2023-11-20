package main

import (
	"fmt"

	tpt "github.com/pcbuildpluscoding/transport"
)

func tcA() error {

  logger.Infof("testcase A - checks NewApiRecord performance")

  rcd, err := tpt.NewApiRecord(map[string]interface{}{
    "Code": 200,
    "Error": "",
    "Data": "ApiRecord creation test",
  })

  if err != nil {
    return err
  }

  logger.Infof("got new ApiRecord : %v", rcd.AsMap())
  if rcd.AppFailed() {
    logger.Errorf("new ApiRecord should not have an error since the Errstr argument was blank, but it does: |%v|", rcd.Error())
  } else {
    logger.Infof("new ApiRecord did not create an error as required")
  }

  if rcd.Parameter().String() == "ApiRecord creation test" {
    logger.Infof("ApiRecord.Parameter.String correctly returned the expected value")
  }
  return nil
}

func tcB() error {

  logger.Infof("testcase A - checks NewApiRecord performance")

  x := map[string]interface{}{
    "Code": 200,
    "Error": "",
  }

  data := map[string]interface{}{
    "Action": "Resume",
    "JobId": "ed07bed0-c30d-49f5-9fde-9cbff6763a64",
    "RemoteAddr": []interface{}{"127.0.0.1:9999","10.0.14.2:8888"},
    "ResumeRef": map[string]interface{}{
        "TaskId": "Task1B",
    },
  }

  rcd, err := tpt.NewApiRecord(x)
  if err != nil {
    return err
  }

  err = rcd.SetData(data)
  if err != nil {
    return err
  }

  logger.Infof("got new ApiRecord : %v", rcd.AsMap())
  if rcd.AppFailed() {
    logger.Errorf("new ApiRecord should not have an error since the Errstr argument was blank, but it does: |%v|", rcd.Error())
    return rcd.Unwrap()
  } else {
    logger.Infof("new ApiRecord did not create an error as required")
  }

  y := rcd.SubNode()
  if err := y.Unwrap(); err != nil {
    return err
  }

  if v := y.String("ResumeRef/TaskId"); v == "Task1B" {
    logger.Infof("ApiRecord.SubNode.String correctly returned the expected value : %s", v)
  } else {
    return fmt.Errorf("expected result : Task1B not received : |%s|", v)
  }

  data["ResumeRef"] = map[string]string{"TaskId": "Task1B"}
  err = rcd.SetData(data)
  if err == nil {
    return fmt.Errorf("ApiRecord did not create a data proto error as required")
  }

  return nil
}