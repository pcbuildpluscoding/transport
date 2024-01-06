package test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/gotestyourself/gotestyourself/assert"
	is "github.com/gotestyourself/gotestyourself/assert/cmp"
	stx "github.com/pcbuildpluscoding/strucex/std"
	tpt "github.com/pcbuildpluscoding/transport"
)

// ----------------------------------------------------------------//
// TestApiResult
// ----------------------------------------------------------------//
func TestApiResult(t *testing.T) {
	tpt.SetLogger(logger)

	rw, err := MarkupToRunware(*dataPath, true)
	if err != nil {
		t.Fatalf("testdata loading error : %v", err)
	}

	t.Run("ApiResult", func(t *testing.T) {
		if tcslice, err := getTestbookB(rw, *testcases); err != nil {
			t.Fatal(err)
		} else {
			for _, tc := range tcslice {
				if tc.actor == nil {
					t.Fatalf("testcase |%s| is undefined", tc.name)
				} else if !rw.HasKeys(tc.dataKey) {
					t.Fatalf("%s dataKey does not exist in dataset", tc.name)
				} else if err := tc.actor(t, rw.SubNode(tc.dataKey, false).Copy()); err != nil {
					t.Fatal(err)
				}
			}
		}
	})
}

// ----------------------------------------------------------------//
// tcb_CheckErr
// ----------------------------------------------------------------//
func tcb_CheckErr(t *testing.T, rw *stx.Strucex) error {
	logger.Debugf("running tc_CheckErr ...")

	x := tpt.ApiResult{}
	err := errors.New("for ApiResult.CheckErr testing")
	y := x.CheckErr(400, err)
	assert.Assert(t, y.Unwrap() != nil, "assert-0")
	assert.Equal(t, "for ApiResult.CheckErr testing", y.Error(), "assert-1")
	logger.Infof("CheckErr default error code : %d", y.Code())
	assert.Equal(t, 400, y.Code(), "assert-2")

	// proving that the optional success_code does not influence the error condition assessment
	y = x.CheckErr(404, err, 201)
	assert.Assert(t, y.Unwrap() != nil, "assert-3")
	assert.Equal(t, "for ApiResult.CheckErr testing", y.Error(), "assert-4")
	assert.Equal(t, 404, y.Code(), "assert-5")

	y = x.CheckErr(404, nil)
	assert.NilError(t, y.Unwrap(), "assert-6")
	assert.Equal(t, 200, y.Code(), "assert-7")

	y = x.CheckErr(404, nil, 201)
	assert.NilError(t, y.Unwrap(), "assert-8")
	assert.Equal(t, 201, y.Code(), "assert-9")

	logger.Infof("tc_CheckErr is complete.")
	return nil
}

// ----------------------------------------------------------------//
// tcb_With
// ----------------------------------------------------------------//
func tcb_With(t *testing.T, rw *stx.Strucex) error {
	logger.Debugf("running tc_With ...")

	x := tpt.ApiResult{}
	y := x.With(300, "some data")
	assert.Equal(t, 300, y.Code(), "assert-2")
	assert.Equal(t, "some data", y.Value(), "assert-3")
	y = x.With(400, rw.AsMap())
	assert.Equal(t, 400, y.Code(), "assert-4")
	ds := getDataset("tc_AsMap")
	assert.Assert(t, is.DeepEqual(ds, y.Value()), "assert-5")

	y = x.Withf(500, "a testing error")
	assert.Equal(t, 500, y.Code(), "assert-6")
	assert.Equal(t, "a testing error", y.Error(), "assert-7")
	y = x.Withf(500, "")
	assert.Equal(t, "", y.Error(), "assert-8")
	y = x.Withf(300, "a testing error")
	assert.Equal(t, 300, y.Code(), "assert-9")
	assert.Equal(t, "", y.Error(), "assert-10")
	assert.Equal(t, "a testing error", y.Value(), "assert-11")

	y = x.Withf(500, "a testing error")
	y = y.Wrapf(500, "another testing error")
	assert.Equal(t, "another testing error : a testing error", y.Error(), "assert-12")

	logger.Debugf("tc_With is complete ...")
	return nil
}

// ------------------------------------------------	------------------//
// getTestCaseKeys
// ------------------------------------------------------------------//
func getTestbookB(rw *stx.Strucex, x string) ([]Testcase, error) {
	w, err := getTestbookKeys(rw, "testbookB", x)
	if err != nil {
		return nil, err
	}
	y := make([]Testcase, len(w))
	for i, z := range w {
		switch strings.TrimSpace(z) {
		case "tcb_CheckErr":
			y[i] = Testcase{actor: tcb_CheckErr, name: "tcb_CheckErr", dataKey: "AsMap"}
		case "tcb_With":
			y[i] = Testcase{actor: tcb_With, name: "tcb_With", dataKey: "AsMap"}
		default:
			return nil, fmt.Errorf("unknown testcase name : |%s|", z)
		}
	}
	return y, nil
}
