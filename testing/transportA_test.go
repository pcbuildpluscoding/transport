package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gotestyourself/gotestyourself/assert"
	stx "github.com/pcbuildpluscoding/strucex/std"
	tpt "github.com/pcbuildpluscoding/transport"
	spb "google.golang.org/protobuf/types/known/structpb"
)

// ----------------------------------------------------------------//
// TestTransport
// ----------------------------------------------------------------//
func TestTransport(t *testing.T) {
	tpt.SetLogger(logger)

	rw, err := MarkupToRunware(*dataPath, true)
	if err != nil {
		t.Fatalf("testdata loading error : %v", err)
	}

	t.Run("transport", func(t *testing.T) {
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
// tc_AsMap
// ----------------------------------------------------------------//
func tc_AsMap(t *testing.T, rw *stx.Strucex) error {
	logger.Debugf("running tc_AsMap ...")

	y, err := tpt.NewApiRecord(rw.AsStruct())
	assert.NilError(t, err, "assert-0")
	assert.Assert(t, !y.Empty(), "assert-1")
	assert.Equal(t, "container restart failed", y.Error(), "assert-2")
	assert.Equal(t, 400, y.Code(), "assert-3")
	z := y.Runware()
	assert.NilError(t, z.Unwrap(), "assert-4")
	assert.Equal(t, 402, z.Int("Code"), "assert-5")
	assert.Equal(t, "application-xyz", z.String("Key"), "assert-6")
	assert.Equal(t, "apple,orange,banana", stringify(z.StringList("Metric")), "assert-7")
	logger.Infof("tc_asMap is complete.")
	return nil
}

// ----------------------------------------------------------------//
// tc_Bicode
// ----------------------------------------------------------------//
func tc_Bicode(t *testing.T, rw *stx.Strucex) error {
	logger.Debugf("running tc_Bicode ...")

	x, err := tpt.NewApiRecord(rw.AsStruct())
	assert.NilError(t, err, "assert-0")
	assert.Assert(t, !x.Empty(), "assert-1")
	assert.Equal(t, "container restart failed", x.Error(), "assert-2")
	frame, err := x.Encode()
	assert.NilError(t, err, "assert-3")
	y := tpt.ApiNote{}
	err = y.Decode(frame)
	assert.NilError(t, err, "assert-4")
	assert.Equal(t, "container restart failed", y.Error(), "assert-5")
	assert.Equal(t, 400, y.Code(), "assert-6")
	z := y.Runware()
	assert.NilError(t, z.Unwrap(), "assert-7")
	assert.Equal(t, 402, z.Int("Code"), "assert-8")
	assert.Equal(t, "application-xyz", z.String("Key"), "assert-9")
	assert.Equal(t, "apple,orange,banana", stringify(z.StringList("Metric")), "assert-10")

	logger.Infof("tc_Bicode is complete.")
	return nil
}

// ----------------------------------------------------------------//
// tc_Bytes
// ----------------------------------------------------------------//
func tc_Bytes(t *testing.T, rw *stx.Strucex) error {
	logger.Debugf("running tc_Bytes ...")

	x := rw.AsStruct()
	assert.NilError(t, rw.Unwrap(), "assert-0")
	var err error
	x.Fields["Data"], err = spb.NewValue([]byte(x.Fields["Data"].GetStringValue()))
	assert.NilError(t, err, "assert-1")
	y, err := tpt.NewApiRecord(x)
	assert.NilError(t, err, "assert-2")
	z, err := y.Bytes()
	assert.NilError(t, err, "assert-3")
	assert.Equal(t, "container restart failed", string(z), "assert-4")

	logger.Infof("tc_Bytes is complete.")
	return nil
}

// ----------------------------------------------------------------//
// tc_ErrorAs
// ----------------------------------------------------------------//
func tc_ErrorAs(t *testing.T, rw *stx.Strucex) error {
	logger.Debugf("running tc_ErrorAs ...")

	y, err := tpt.NewApiRecord(rw.AsStruct())
	assert.NilError(t, err, "assert-0")
	assert.Assert(t, y.Unwrap() != nil, "assert-1")
	assert.Assert(t, y.Unwrap() != nil, "assert-1A")
	assert.Assert(t, y.ErrorAs(&err), "assert-2")
	assert.Assert(t, err != nil, "assert-3")
	assert.Equal(t, "test error", err.Error(), "assert-4")

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
		case "tc_AsMap":
			y[i] = Testcase{actor: tc_AsMap, name: "tc_AsMap", dataKey: "AsMap"}
		case "tc_Bicode":
			y[i] = Testcase{actor: tc_Bicode, name: "tc_Bicode", dataKey: "AsMap"}
		case "tc_Bytes":
			y[i] = Testcase{actor: tc_Bytes, name: "tc_Bytes", dataKey: "Bytes"}
		case "tc_ErrorAs":
			y[i] = Testcase{actor: tc_ErrorAs, name: "tc_ErrorAs", dataKey: "ErrorAs"}
		default:
			return nil, fmt.Errorf("unknown testcase name : |%s|", z)
		}
	}
	return y, nil
}
