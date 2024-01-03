package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gotestyourself/gotestyourself/assert"
	stx "github.com/pcbuildpluscoding/strucex/std"
	tpt "github.com/pcbuildpluscoding/transport"
)

// ----------------------------------------------------------------//
// TestTrovian
// ----------------------------------------------------------------//
func TestTrovian(t *testing.T) {
	stx.SetLogger(logger)

	rw, err := MarkupToRunware(*dataPath, true)
	if err != nil {
		t.Fatalf("testdata loading error : %v", err)
	}

	t.Run("n", func(t *testing.T) {
		if tcslice, err := getTestbookA(rw, *testcases); err != nil {
			t.Fatal(err)
		} else {
			for _, tc := range tcslice {
				if tc.actor == nil {
					t.Fatalf("testcase |%s| is undefined", tc.name)
				} else if x := rw.SubNode(tc.dataKey, false); x.Empty() {
					t.Fatalf("%s dataKey does not exist in dataset", tc.name)
				} else if err := tc.actor(t, x); err != nil {
					t.Fatal(err)
				}
			}
		}
	})
}

// ----------------------------------------------------------------//
// tc_asMap
// ----------------------------------------------------------------//
func tc_asMap(t *testing.T, rw *stx.Strucex) error {
	logger.Debugf("running tc_asMap ...")

	y, err := tpt.NewApiRecord(rw.AsStruct())
	assert.NilError(t, err, "assert-0")
	assert.Assert(t, !y.Empty(), "assert-1")
	z := y.AsMap()
	assert.Equal(t, 400, int(z["Code"].(float64)), "assert-2")
	assert.Equal(t, "some test data", stringify(z["Data"], " "), "assert-3")
	zz, err := stx.NewRunware(z["ApiError"])
	assert.NilError(t, err, "assert-4")
	assert.Equal(t, 402, zz.Int("Code"), "assert-4")
	assert.Equal(t, "test-key", zz.String("Key"), "assert-5")
	assert.Equal(t, "test error", zz.String("Error"), "assert-6")
	logger.Infof("tc_asMap is complete.")
	return nil
}

// ----------------------------------------------------------------//
// tc_errorAs
// ----------------------------------------------------------------//
func tc_errorAs(t *testing.T, rw *stx.Strucex) error {
	logger.Debugf("running tc_errorAs ...")

	y, err := tpt.NewApiRecord(rw.AsStruct())
	assert.NilError(t, err, "assert-0")
	assert.Assert(t, y.Unwrap() != nil, "assert-1")
	y.ErrorAs(&err)
	assert.Assert(t, err != nil, "assert-2")
	assert.Equal(t, "test error", err.Error(), "assert-3")

	logger.Infof("tc_errorAs is complete.")
	return nil
}

// ------------------------------------------------------------------//
// getTestCaseKeys
// ------------------------------------------------------------------//
func getTestbookA(rw *stx.Strucex, x string) ([]Testcase, error) {
	w, err := getTestbookKeys(rw, "testbookA", x)
	if err != nil {
		return nil, err
	}
	y := make([]Testcase, len(w))
	for i, z := range w {
		switch strings.TrimSpace(z) {
		case "tc_asMap":
			y[i] = Testcase{actor: tc_asMap, name: "tc_asMap", dataKey: "AsMap"}
		case "tc_errorAs":
			y[i] = Testcase{actor: tc_errorAs, name: "tc_errorAs", dataKey: "ErrorAs"}
		default:
			return nil, fmt.Errorf("unknown testcase name : |%s|", z)
		}
	}
	return y, nil
}
