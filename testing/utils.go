package test

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pcbuildpluscoding/logroll"

	stx "github.com/pcbuildpluscoding/strucex/std"
	spb "google.golang.org/protobuf/types/known/structpb"
	"sigs.k8s.io/yaml"
)

type tcActor func(*testing.T, *stx.Strucex) error
type Testcase struct {
	actor   tcActor
	name    string
	dataKey string
}

var logger = logroll.Get()

var remakeBucket = flag.Bool("remakeBucket", false, "remake the testing bucket on startup")

var bucketName = flag.String("bucketName", "", "trovedb bucket name")

var createBucket = flag.Bool("create", false, "create a new db bucket")

var testcases = flag.String("testcases", "", "comma separated list of testcases to run")

var dataPath = flag.String("dataPath", "", "input yaml datafile path for testing")

var dumpPath = flag.String("dumpPath", "/data/captainia/trovedb/bucketDump.txt", "system file path for bucket dump output")

// ------------------------------------------------------------------//
// getDataset
// ------------------------------------------------------------------//
func getDataset(tc_name string) map[string]interface{} {
	switch tc_name {
	case "tc_AsMap":
		return map[string]interface{}{
			"Code":  float64(400),
			"Error": "container restart failed",
			"Data": map[string]interface{}{
				"Code": float64(402),
				"Key":  "application-xyz",
				"Metric": []interface{}{
					"apple", "orange", "banana",
				},
			},
		}
	}
	return map[string]interface{}{}
}

// ------------------------------------------------------------------//
// getErrorString
// ------------------------------------------------------------------//
func getErrorString(keys ...string) string {
	errtxt := " does not exist"
	var x string
	for i, key := range keys {
		if i == 0 {
			x = key + errtxt
			continue
		}
		x += "\n" + key + errtxt
	}
	return x
}

// ------------------------------------------------------------------//
// MarkupToRunware
// ------------------------------------------------------------------//
func MarkupToRunware(filePath string, ignoreNilValue bool, key ...string) (*stx.Strucex, error) {

	ext := filepath.Ext(filePath)
	switch ext {
	case ".yaml", ".yml", ".json":
	default:
		return nil, fmt.Errorf("unsupported markup type : %s", filepath.Base(filePath))
	}

	frame, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	switch ext {
	case ".yaml", ".yml":
		frame, err = yaml.YAMLToJSON(frame)
		if err != nil {
			return nil, fmt.Errorf("YAMLToJson conversion failed : %v", err)
		}
	}

	v, _ := spb.NewValue(nil)
	// logger.Debugf("about unmarshal file input into structpb.Value ...")

	err = v.UnmarshalJSON(frame)
	if err != nil {
		return nil, fmt.Errorf("structpb.Value unmarshaling failed : %v", err)
	}
	rw, err := stx.NewRunware(v, ignoreNilValue)
	if err != nil {
		return nil, err
	} else if key != nil {
		return rw.SubNode(key[0], false), nil
	}
	return rw, nil
}

// ----------------------------------------------------------------//
// utils
// ----------------------------------------------------------------//
// ----------------------------------------------------------------//
// getIntSlice
// ----------------------------------------------------------------//
func getIntSlice(size int) []interface{} {
	x := make([]interface{}, size)
	for i := range x {
		y := 0
		x[i] = &y
	}
	return x
}

// ------------------------------------------------------------------//
// getTestbookKeys
// ------------------------------------------------------------------//
func getTestbookKeys(rw *stx.Strucex, bookId, tc_desc string) ([]string, error) {
	if tc_desc == "__all__" {
		p := rw.Parameter(bookId + "/all")
		return p.StringList(), p.Unwrap()
	}
	return strings.Split(tc_desc, ","), nil
}

// ----------------------------------------------------------------//
// stringify
// ----------------------------------------------------------------//
func stringify(x interface{}, darg ...string) string {
	d := ","
	if darg != nil {
		d = darg[0]
	}
	switch y := x.(type) {
	case []string:
		return strings.Join(y, d)
	case []interface{}:
		return strings.Join(toStringList(y), d)
	default:
		return ""
	}
}

// ----------------------------------------------------------------//
// toStringList
// ----------------------------------------------------------------//
func toStringList(x []interface{}) []string {
	result := make([]string, len(x))
	for i, ival := range x {
		result[i], _ = ival.(string)
	}
	return result
}

// ----------------------------------------------------------------//
// toByteSlice
// ----------------------------------------------------------------//
func toByteSlice(x []string) []interface{} {
	var y []interface{}
	for i, v := range x {
		if i == 0 {
			y = append(y, []byte(v))
		} else {
			y = append(y, []byte(v))
		}
	}
	return y
}

// ----------------------------------------------------------------//
// sumIntSlice
// ----------------------------------------------------------------//
func sumIntSlice(args ...interface{}) int {
	x := 0
	for _, y := range args {
		switch z := y.(type) {
		case *int:
			x += *z
		case float64:
			x += int(z)
		}
	}
	return x
}
