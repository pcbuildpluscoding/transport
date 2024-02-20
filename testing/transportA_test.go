package test

import (
	"encoding/base64"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	stx "github.com/pcbuildpluscoding/strucex/std"
	tpt "github.com/pcbuildpluscoding/transport"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

// ----------------------------------------------------------------//
// TestApiNote
// ----------------------------------------------------------------//
func TestApiNote(t *testing.T) {
	tpt.SetLogger(logger)

	rw, err := MarkupToRunware(*dataPath, true)
	if err != nil {
		t.Fatalf("testdata loading error : %v", err)
	}

	t.Run("ApiNote", func(t *testing.T) {
		if tcslice, err := getTestbookA(rw, *testcases); err != nil {
			t.Fatal(err)
		} else {
			for _, tc := range tcslice {
				if tc.actor == nil {
					t.Fatalf("testcase |%s| is undefined", tc.name)
				} else if !rw.HasKeys(tc.dataKey) {
					t.Fatalf("%s dataKey does not exist in dataset", tc.name)
				} else if err := tc.actor(t, rw.SubNode(tc.dataKey).Copy(), nil); err != nil {
					t.Fatal(err)
				}
			}
		}
	})
}

// ----------------------------------------------------------------//
// tc_AsMap
// ----------------------------------------------------------------//
func tc_AsMap(t *testing.T, rw *stx.Strucex, arg interface{}) error {
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
func tc_Bicode(t *testing.T, rw *stx.Strucex, arg interface{}) error {
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
func tc_Bytes(t *testing.T, rw *stx.Strucex, arg interface{}) error {
	logger.Debugf("running tc_Bytes ...")

	x, err := tpt.NewApiRecord(nil)
	assert.NilError(t, err, "assert-0")
	err = x.SetData([]byte(rw.String("Data")))
	assert.NilError(t, err, "assert-2")
	y, err := x.Bytes()
	assert.NilError(t, err, "assert-3")
	frame := []byte("container restart failed")
	assert.Equal(t, len(frame), len(y), "assert-4")
	for i, chr := range y {
		assert.Equal(t, frame[i], chr, "assert-5-%d", i)
	}

	logger.Infof("tc_Bytes is complete.")
	return nil
}

// ----------------------------------------------------------------//
// tc_ErrorAs
// ----------------------------------------------------------------//
func tc_ErrorAs(t *testing.T, rw *stx.Strucex, arg interface{}) error {
	logger.Debugf("running tc_ErrorAs ...")

	y, err := tpt.NewApiRecord(rw.AsStruct())
	assert.NilError(t, err, "assert-0")
	assert.Assert(t, y.Unwrap() != nil, "assert-1")

	assert.Assert(t, y.As(&err), "assert-2")
	assert.Assert(t, err != nil, "assert-3")
	assert.Equal(t, "test error", err.Error(), "assert-4")

	errA := stx.NewNilPathError("foo/bar")
	y = y.With(400, errA)

	assert.Assert(t, y.As(&err), "assert-5")
	assert.Assert(t, err != nil, "assert-6")
	assert.Equal(t, "foo/bar does not exist", err.Error(), "assert-7")

	logger.Infof("tc_errorAs is complete.")
	return nil
}

// ----------------------------------------------------------------//
// tc_Empty
// ----------------------------------------------------------------//
func tc_Empty(t *testing.T, rw *stx.Strucex, arg interface{}) error {
	logger.Debugf("running tc_Empty ...")

	x, err := tpt.NewApiRecord(nil)
	assert.NilError(t, err, "assert-0")
	assert.Assert(t, x.Empty(), "assert-1")
	assert.NilError(t, x.SetData("foo"), "assert-2")
	assert.Assert(t, !x.Empty(), "assert-3")
	assert.NilError(t, x.SetData(nil), "assert-4")
	assert.Assert(t, x.Empty(), "assert-5")
	x = x.With(200, nil)
	assert.Assert(t, !x.Empty(), "assert-6")
	y := x.With(0, nil)
	assert.Assert(t, y.Empty(), "assert-7")
	y = y.With(0, "foo")
	assert.Assert(t, !y.Empty(), "assert-8")
	y = y.With(0, nil, nil)
	assert.Assert(t, y.Empty(), "assert-9")
	y = y.With(0, errors.New("for assert-10"))
	assert.Assert(t, !y.Empty(), "assert-10")
	assert.Equal(t, "for assert-10", y.Error(), "assert-11")
	logger.Debugf("tc_Empty is complete ...")
	return nil
}

// ----------------------------------------------------------------//
// tc_Error
// ----------------------------------------------------------------//
func tc_Error(t *testing.T, rw *stx.Strucex, arg interface{}) error {
	logger.Debugf("running tc_Error ...")

	x, err := tpt.NewApiRecord(rw.AsStruct())
	assert.NilError(t, err, "assert-0")
	assert.Equal(t, "container restart failed", x.Error(), "assert-1")
	y := x.With(0, nil)
	assert.Equal(t, "container restart failed", y.Error(), "assert-2")
	y = x.With(0, errors.New(""))
	assert.Assert(t, y.Error() == "", "assert-3")
	assert.Assert(t, y.Unwrap() != nil, "assert-4")
	y = y.With(0, errors.New("setting new error"))
	assert.Equal(t, "setting new error", y.Error(), "assert-5")
	y, err = tpt.NewApiRecord(nil)
	assert.NilError(t, err, "assert-6")
	assert.Assert(t, y.Error() == "", "assert-7")
	logger.Debugf("tc_Error is complete ...")
	return nil
}

// ----------------------------------------------------------------//
// tc_Hardcopy
// ----------------------------------------------------------------//
func tc_Hardcopy(t *testing.T, rw *stx.Strucex, arg interface{}) error {
	logger.Debugf("running tc_Hardcopy ...")

	x, err := tpt.NewApiRecord(rw.AsStruct())
	assert.NilError(t, err, "assert-0")
	y := x.Copy()
	assert.NilError(t, err, "assert-1")
	assert.Equal(t, reflect.TypeOf(x.Value()), reflect.TypeOf(y.Value()), "assert-2")
	x, _ = tpt.NewApiRecord(nil)

	y = x.With(0, rw.Struct("Data"))
	z := y.Copy()
	assert.NilError(t, err, "assert-3")
	assert.Equal(t, reflect.TypeOf(y.Value()), reflect.TypeOf(z.Value()), "assert-4")
	err = y.Runware().Set("Key", "application-abc").Unwrap()
	assert.NilError(t, err, "assert-5")
	assert.Equal(t, z.Runware().String("Key"), "application-xyz", "assert-6")
	logger.Debugf("tc_Hardcopy is complete ...")
	return nil
}

type TestError struct {
	Text string
}

func (e *TestError) Error() string {
	return e.Text
}

// ----------------------------------------------------------------//
// tc_Is
// ----------------------------------------------------------------//
func tc_Is(t *testing.T, rw *stx.Strucex, arg interface{}) error {
	logger.Debugf("running tc_Is ...")

	x, err := tpt.NewApiRecord(rw.AsStruct())
	assert.NilError(t, err, "assert-0")
	assert.Equal(t, "container restart failed", x.Error(), "assert-1")
	assert.Assert(t, x.Is(errors.New("container restart failed")), "assert-2")

	var erra = errors.New("something broke!")
	y := x.With(0, erra)
	var errb error = fmt.Errorf("something else broke! : %w", erra)
	y = y.With(0, errb)
	assert.Assert(t, y.Is(erra), "assert-3")

	logger.Debugf("tc_Is is complete ...")
	return nil
}

// ----------------------------------------------------------------//
// tc_Parameter
// ----------------------------------------------------------------//
func tc_Parameter(t *testing.T, rw *stx.Strucex, arg interface{}) error {
	logger.Debugf("running tc_Parameter ...")

	x, err := tpt.NewApiRecord(rw.AsStruct())
	assert.NilError(t, err, "assert-0")
	assert.Equal(t, "container restart failed", x.Error(), "assert-1")
	y := x.Parameter().Runware()
	assert.Equal(t, "apple,orange,banana", stringify(y.StringList("Metric")), "assert-2")
	x = x.With(0, y.AsInterface("Metric"))
	z := x.Parameter()
	assert.NilError(t, z.Unwrap(), "assert-3")
	assert.Equal(t, "apple,orange,banana", stringify(z.StringList()), "assert-4")

	logger.Debugf("tc_Parameter is complete ...")
	return nil
}

// ----------------------------------------------------------------//
// tc_Runware
// ----------------------------------------------------------------//
func tc_Runware(t *testing.T, rw *stx.Strucex, arg interface{}) error {
	logger.Debugf("running tc_Runware ...")

	x, err := tpt.NewApiRecord(rw.AsStruct())
	assert.NilError(t, err, "assert-0")
	assert.Equal(t, "container restart failed", x.Error(), "assert-1")
	y := x.Runware()
	assert.NilError(t, y.Unwrap(), "assert-3")
	assert.Equal(t, "apple,orange,banana", stringify(y.StringList("Metric")), "assert-4")

	x = x.With(0, map[string]interface{}{
		"Code":  505,
		"JobId": "ab51cf5a-ab92-11ee-9946-abe5079abce3",
	})
	z := x.Runware()
	assert.NilError(t, z.Unwrap(), "assert-5")
	assert.Equal(t, 505, z.Int("Code"), "assert-6")
	assert.Equal(t, "ab51cf5a-ab92-11ee-9946-abe5079abce3", z.String("JobId"), "assert-7")

	z = x.With(0, *rw).Runware()
	assert.NilError(t, z.Unwrap(), "assert-8")
	assert.Equal(t, 402, z.Int("Data/Code"), "assert-9")

	logger.Debugf("tc_Runware is complete ...")
	return nil
}

// ----------------------------------------------------------------//
// tc_SetData
// ----------------------------------------------------------------//
func tc_SetData(t *testing.T, rw *stx.Strucex, arg interface{}) error {
	logger.Debugf("running tc_SetData ...")

	x, err := tpt.NewApiRecord(nil)
	assert.NilError(t, err, "assert-0")
	err = x.SetData(rw)
	assert.NilError(t, err, "assert-1")
	y := x.Runware()
	assert.NilError(t, y.Unwrap(), "assert-2")
	assert.Equal(t, "apple,orange,banana", stringify(y.StringList("Data/Metric")), "assert-3")
	err = x.SetData([]string{"dog", "cat", "bird"})
	assert.NilError(t, err, "assert-4")
	z := x.Parameter()
	assert.Equal(t, "dog,cat,bird", stringify(z.StringList()), "assert-5")
	err = x.SetData(nil)
	assert.NilError(t, err, "assert-6")
	assert.Assert(t, x.Empty(), "assert-7")
	err = x.SetData([]interface{}{10, 15.4, 2222})
	assert.NilError(t, err, "assert-8")
	z = x.Parameter()
	assert.Equal(t, 2247, sumIntSlice(z.List().AsSlice()...), "assert-9")
	logger.Debugf("tc_SetData is complete ...")
	return nil
}

// ----------------------------------------------------------------//
// tc_Value
// ----------------------------------------------------------------//
func tc_Value(t *testing.T, rw *stx.Strucex, arg interface{}) error {
	logger.Debugf("running tc_Value ...")

	x, err := tpt.NewApiRecord(rw.AsStruct())
	assert.NilError(t, err, "assert-0")
	assert.Assert(t, !x.Empty(), "assert-1")
	ds := getDataset("tc_AsMap")
	assert.Assert(t, is.DeepEqual(ds["Data"], x.Value()), "assert-2")
	err = x.SetData("foo")
	assert.NilError(t, err, "assert-3")
	assert.Equal(t, "foo", x.Value(), "assert-4")
	err = x.SetData([]byte(x.Error()))
	assert.NilError(t, err, "assert-5")
	s := base64.StdEncoding.EncodeToString([]byte("container restart failed"))
	assert.Equal(t, s, x.Value(), "assert-6")
	x, _ = tpt.NewApiRecord(nil)
	assert.Assert(t, x.Empty(), "assert-7")
	assert.Equal(t, nil, x.Value(), "assert-8")

	logger.Debugf("tc_Value is complete ...")
	return nil
}

// ----------------------------------------------------------------//
// tc_With
// ----------------------------------------------------------------//
func tc_With(t *testing.T, rw *stx.Strucex, arg interface{}) error {
	logger.Debugf("running tc_With ...")

	x, _ := tpt.NewApiRecord(nil)
	assert.Assert(t, x.Empty(), "assert-0")
	assert.Equal(t, nil, x.Value(), "assert-1")
	y := x.With(300, "some data")
	assert.Equal(t, 300, y.Code(), "assert-2")
	assert.Equal(t, "some data", y.Value(), "assert-3")
	y = x.With(400, rw.AsMap())
	assert.Equal(t, 400, y.Code(), "assert-4")
	ds := getDataset("tc_AsMap")
	assert.Assert(t, is.DeepEqual(ds, x.Value()), "assert-5")

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

	x, _ = tpt.NewApiRecord(nil)
	err := errors.New("something broke")
	y = x.With(400, err, 1234)
	assert.Equal(t, 400, y.Code(), "assert-13")
	assert.Equal(t, "something broke", y.Error(), "assert-14")
	assert.Equal(t, 1234, y.Parameter().Int(), "assert-15")
	err = errors.New("just now!")
	y = y.Wrapf(500, "something else broke : %v", err)
	assert.Equal(t, 500, y.Code(), "assert-16")
	// logger.Debugf("wrapped error : |%s|", y.Error())
	assert.Equal(t, "something else broke : just now! : something broke", y.Error(), "assert-17")
	assert.Equal(t, 1234, y.Parameter().Int(), "assert-18")
	y = y.With(501, nil, map[string]interface{}{
		"foo": "bar",
	})
	assert.Equal(t, 501, y.Code(), "assert-19")
	assert.Equal(t, "something else broke : just now! : something broke", y.Error(), "assert-20")
	assert.Equal(t, "bar", y.Runware().String("foo"), "assert-21")

	y = y.With(503, []string{"a new", "data", "kind"})
	assert.Equal(t, "a new data kind", stringify(y.Parameter().StringList(), " "), "assert-22")

	r := tpt.ApiResult{}
	err = errors.New("something broke")
	y = r.With(400, err, 1234)
	assert.Equal(t, 400, y.Code(), "assert-23")
	assert.Equal(t, "something broke", y.Error(), "assert-24")
	assert.Equal(t, 1234, y.Parameter().Int(), "assert-25")

	y = r.With(503, []string{"a new", "data", "kind"})
	assert.Equal(t, 503, y.Code(), "assert-26")
	assert.NilError(t, y.Unwrap(), "assert-27")
	assert.Equal(t, "a new data kind", stringify(y.Parameter().StringList(), " "), "assert-28")

	y = y.With(500, errors.New("small error"), errors.New("data error"))
	assert.Equal(t, 500, y.Code(), "assert-28")
	assert.Equal(t, "small error", y.Error(), "assert-29")
	assert.Equal(t, "data error", y.Parameter().String(), "assert-30")

	logger.Debugf("tc_With is complete ...")
	return nil
}

// ------------------------------------------------	------------------//
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
		case "tc_Empty":
			y[i] = Testcase{actor: tc_Empty, name: "tc_Empty", dataKey: "default"}
		case "tc_Error":
			y[i] = Testcase{actor: tc_Error, name: "tc_Error", dataKey: "AsMap"}
		case "tc_ErrorAs":
			y[i] = Testcase{actor: tc_ErrorAs, name: "tc_ErrorAs", dataKey: "ErrorAs"}
		case "tc_Hardcopy":
			y[i] = Testcase{actor: tc_Hardcopy, name: "tc_Hardcopy", dataKey: "AsMap"}
		case "tc_Is":
			y[i] = Testcase{actor: tc_Is, name: "tc_Is", dataKey: "AsMap"}
		case "tc_Parameter":
			y[i] = Testcase{actor: tc_Parameter, name: "tc_Parameter", dataKey: "AsMap"}
		case "tc_Runware":
			y[i] = Testcase{actor: tc_Runware, name: "tc_Runware", dataKey: "AsMap"}
		case "tc_SetData":
			y[i] = Testcase{actor: tc_SetData, name: "tc_SetData", dataKey: "AsMap"}
		case "tc_Value":
			y[i] = Testcase{actor: tc_Value, name: "tc_Value", dataKey: "AsMap"}
		case "tc_With":
			y[i] = Testcase{actor: tc_With, name: "tc_With", dataKey: "AsMap"}
		default:
			return nil, fmt.Errorf("unknown testcase name : |%s|", z)
		}
	}
	return y, nil
}
