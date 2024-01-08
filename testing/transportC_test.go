package test

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	// is "github.com/gotestyourself/gotestyourself/assert/cmp"
	stx "github.com/pcbuildpluscoding/strucex/std"
	tpt "github.com/pcbuildpluscoding/transport"
	"gotest.tools/v3/assert"
)

// ----------------------------------------------------------------//
// TestIoUtils
// ----------------------------------------------------------------//
func TestIoUtils(t *testing.T) {
	tpt.SetLogger(logger)

	rw, err := MarkupToRunware(*dataPath, true)
	if err != nil {
		t.Fatalf("testdata loading error : %v", err)
	}

	req := rw.SubNode("TestClient", false)
	client, err := NewTestClient(req)
	if err != nil {
		t.Fatalf("client error : %v", err)
	}

	remoteAddr := req.String("RemoteAddr")

	conn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		t.Fatalf("net.Dial error : %v", err)
	}

	err = client.Init(conn, req)
	if err != nil {
		t.Fatalf("client error : %v", err)
	}

	t.Cleanup(func() {
		logger.Debugf("running IoUtils cleanup ...")
		client.Close(req.String("Request/JobId"))
	})

	t.Run("IoUtils", func(t *testing.T) {
		if tcslice, err := getTestbookC(rw, *testcases); err != nil {
			t.Fatal(err)
		} else {
			for _, tc := range tcslice {
				if tc.actor == nil {
					t.Fatalf("testcase |%s| is undefined", tc.name)
				} else if !rw.HasKeys(tc.dataKey) {
					t.Fatalf("%s dataKey does not exist in dataset", tc.name)
				} else if err := tc.actor(t, req, client); err != nil {
					t.Fatal(err)
				}
			}
		}
	})
}

// ----------------------------------------------------------------//
// tcc_ioutils
// ----------------------------------------------------------------//
func tcc_ioutils(t *testing.T, rw *stx.Strucex, arg interface{}) error {
	logger.Debugf("running tc_readConn ...")

	// wdata := rw.Pop("WriteData")
	// if wdata.Empty() {
	// 	return fmt.Errorf("write timeout data is undefined")
	// }
	tc, _ := arg.(*TestClient)
	req := rw.SubNode("Request", false).Copy()
	_ = req.Set("Action", "ReadTimeout")
	_ = req.Set("ReadTimeout", rw.Int("ReadTimeout")+1)
	logger.Debugf("running with request : %v", req.AsMap())
	resp := toApiNote(tc.Request(req))
	if !assert.Check(t, errors.Is(resp.Unwrap(), os.ErrDeadlineExceeded), "assert-0") {
		return fmt.Errorf("assert-0 failed : %v", resp.Unwrap())
	}
	assert.Equal(t, 0, resp.Parameter().Int(), "assert-1")

	req = rw.SubNode("Request", false).Copy()
	_ = req.Set("Action", "WriteTimeout")
	_ = req.Set("WriteTimeout", rw.Int("WriteTimeout")+2)
	logger.Debugf("sending async request : %v", req.AsMap())
	err := tc.AsyncRequest(req)
	if err != nil {
		return fmt.Errorf("write timeout request error : %v", err)
	}

	if *hugeDataPath == "" {
		return fmt.Errorf("huge data path is undefined")
	}

	infile, err := os.Open(*hugeDataPath)
	if err != nil {
		return fmt.Errorf("huge data open file error : %v", err)
	}

	scanner := bufio.NewScanner(infile)
	scanner.Split(bufio.ScanLines)

	tc.ResetVars()
	for scanner.Scan() {
		if err := tc.AddLine(scanner.Text()); err != nil {
			if !assert.Check(t, errors.Is(err, os.ErrDeadlineExceeded), "assert-2") {
				return fmt.Errorf("assert-2 failed : %v", err)
			} else if err = tc.conn.SetWriteDeadline(time.Time{}); err != nil {
				return err
			}
			tc.ResetVars()
			return nil
		}
	}
	return nil
}

// ------------------------------------------------	------------------//
// getTestCaseKeys
// ------------------------------------------------------------------//
func getTestbookC(rw *stx.Strucex, x string) ([]Testcase, error) {
	w, err := getTestbookKeys(rw, "testbookB", x)
	if err != nil {
		return nil, err
	}
	y := make([]Testcase, len(w))
	for i, z := range w {
		switch strings.TrimSpace(z) {
		case "tcc_ioutils":
			y[i] = Testcase{actor: tcc_ioutils, name: "tcc_ioutils", dataKey: "TestClient"}
		default:
			return nil, fmt.Errorf("unknown testcase name : |%s|", z)
		}
	}
	return y, nil
}
