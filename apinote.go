// The MIT License
//
// Copyright (c) 2022 Peter A McGill
package transport

import (
	"errors"
	"fmt"
	"time"

	stx "github.com/pcbuildpluscoding/strucex/std"
	spb "google.golang.org/protobuf/types/known/structpb"
)

// ================================================================//
// ApiNote
// ================================================================//
type ApiNote struct {
	code      int
	err       error
	data      interface{}
	timestamp interface{}
}

// ------------------------------------------------------------------//
// As
// ------------------------------------------------------------------//
func (n *ApiNote) ErrorAs(target interface{}) bool {
	return ErrorAs(n.err, target)
}

// ----------------------------------------------------------------//
// AsMap
// ----------------------------------------------------------------//
func (n *ApiNote) AsMap() map[string]interface{} {
	x := spb.Struct{
		Fields: n.asVMap(),
	}
	return x.AsMap()
}

// ----------------------------------------------------------------//
// asVMap
// ----------------------------------------------------------------//
func (n *ApiNote) asVMap() map[string]*spb.Value {
	x := map[string]*spb.Value{}
	x["Code"] = spb.NewNumberValue(float64(n.code))
	if n.err != nil {
		x["Error"] = spb.NewStringValue(n.err.Error())
	}

	var err error

	if n.timestamp != nil {
		x["Timestamp"], err = GetTimestampA("", n.timestamp)
	}
	if n.data == nil {
		return x
	}

	if err != nil {
		panic(fmt.Errorf("ApiNote timestamp field is not a valid structpb.Value type : %v", err))
	}

	switch n.data.(type) {
	case *Strucex:
		rw := n.data.(*Strucex)
		x["Data"], err = spb.NewValue(rw.AsMap())
	case *spb.Value:
		x["Data"] = n.data.(*spb.Value)
	default:
		x["Data"], err = spb.NewValue(n.data)
	}

	if err != nil {
		panic(fmt.Errorf("ApiNote data field is not a valid structpb.Value type : %v", err))
	}

	return x
}

// ----------------------------------------------------------------//
// Body
// ----------------------------------------------------------------//
func (n *ApiNote) Bytes() ([]byte, error) {
	switch n.data.(type) {
	case []byte:
		return n.data.([]byte), nil
	default:
		return nil, fmt.Errorf("Record data type is not []byte")
	}
}

// ----------------------------------------------------------------//
// Code
// ----------------------------------------------------------------//
func (n *ApiNote) Code() int {
	return n.code
}

// ------------------------------------------------------------ //
// Decode
// -------------------------------------------------------------//
func (n *ApiNote) Decode(frame []byte) error {
	var s spb.Struct
	err := s.UnmarshalJSON(frame)
	if err != nil {
		return err
	}
	n.fromMap(s.Fields)
	return nil
}

// ------------------------------------------------------------ //
// Encode
// -------------------------------------------------------------//
func (n *ApiNote) Encode() ([]byte, error) {
	stru := spb.Struct{
		Fields: n.asVMap(),
	}
	return stru.MarshalJSON()
}

// ------------------------------------------------------------ //
// fromMap
// -------------------------------------------------------------//
func (n *ApiNote) fromMap(vm map[string]*spb.Value) {
	if x, ok := vm["Code"]; ok {
		n.code = int(x.GetNumberValue())
	}
	if x, ok := vm["Timestamp"]; ok {
		n.timestamp = x.AsInterface()
	}
	if x, ok := vm["Error"]; ok {
		if errtxt := x.GetStringValue(); errtxt != "" {
			n.err = errors.New(errtxt)
		}
	} else if x, ok := vm["ApiError"]; ok {
		switch x.GetKind().(type) {
		case *spb.Value_StructValue:
			if y := x.GetStructValue(); y != nil {
				var err error
				if errtxt := y.Fields["Error"].GetStringValue(); errtxt != "" {
					err = errors.New(errtxt)
				}
				n.err = ApiError{
					code: int(y.Fields["Code"].GetNumberValue()),
					key:  y.Fields["AppKey"].GetStringValue(),
					err:  err,
				}
			}
		}
	}
	if x, ok := vm["Data"]; ok {
		n.data = x.AsInterface()
	}
}

// ------------------------------------------------------------------//
// Error
// ------------------------------------------------------------------//
func (n ApiNote) Error() string {
	return fmt.Sprintf("code: %d, error: %v", n.code, n.err)
}

// -------------------------------------------------------------- //
// Empty
// ---------------------------------------------------------------//
func (n ApiNote) Empty() bool {
	return n.data == nil
}

// ----------------------------------------------------------------//
// Hardcopy
// ----------------------------------------------------------------//
func (n *ApiNote) Hardcopy() ApiNote {
	return *n
}

// ------------------------------------------------------------------//
// Is
// ------------------------------------------------------------------//
func (n ApiNote) Is(target error) bool {
	return errors.Is(n.err, target)
}

// -------------------------------------------------------------- //
// Parameter
// ---------------------------------------------------------------//
func (n ApiNote) Parameter() *Parametric {
	v, _ := stx.NewParameter("-", n.data)
	return v
}

// ----------------------------------------------------------------//
// SetData
// ----------------------------------------------------------------//
func (n *ApiNote) SetData(data interface{}) error {
	_, err := spb.NewValue(data)
	n.data = data
	return err
}

// ----------------------------------------------------------------//
// SetTimestamp
// ----------------------------------------------------------------//
func (n *ApiNote) SetTimestamp() {
	n.timestamp = time.Now()
}

// -------------------------------------------------------------- //
// SubNode
// ---------------------------------------------------------------//
func (n ApiNote) SubNode() *Strucex {
	rw, _ := stx.NewRunware(n.data)
	return rw
}

// ----------------------------------------------------------------//
// Timestamp
// ----------------------------------------------------------------//
func (n *ApiNote) Timestamp(format ...string) string {
	timestamp, err := GetTimestamp(n.timestamp, format...)
	if err != nil {
		panic(err)
	}
	return timestamp
}

// ------------------------------------------------------------------//
// Unwrap
// ------------------------------------------------------------------//
func (n ApiNote) Unwrap() error {
	return n.err
}

// ------------------------------------------------------------ //
// Value
// -------------------------------------------------------------//
func (n *ApiNote) Value() interface{} {
	return n.data
}

// ------------------------------------------------------------ //
// With
// -------------------------------------------------------------//
func (n *ApiNote) With(code int, data interface{}) *ApiNote {
	n.code = code
	switch v := data.(type) {
	case error:
		n.err = v
	default:
		n.data = data
	}
	return n
}

// ----------------------------------------------------------------//
// Withf
// ----------------------------------------------------------------//
func (n *ApiNote) Withf(code int, format string, args ...interface{}) *ApiNote {
	n.code = code
	if code >= 400 {
		n.err = fmt.Errorf(format, args...)
	} else {
		n.data = fmt.Sprintf(format, args...)
	}
	return n
}

// ----------------------------------------------------------------//
// Wrapf - wraps an existing error
// ----------------------------------------------------------------//
func (n *ApiNote) Wrapf(code int, format string, args ...interface{}) *ApiNote {
	if n.err == nil {
		return n.Withf(code, format, args...)
	}
	errTxt := fmt.Sprintf(format, args...) + " : %w"
	n.code = code
	n.err = fmt.Errorf(errTxt, n.err)
	return n
}

// ================================================================//
// ApiResult
// ================================================================//
type ApiResult struct{}

// ----------------------------------------------------------------//
// WithCode
// ----------------------------------------------------------------//
func (ApiResult) With(code int, data interface{}) *ApiNote {
	x := ApiNote{}
	return x.With(code, data)
}

// ----------------------------------------------------------------//
// Withf
// ----------------------------------------------------------------//
func (ApiResult) Withf(code int, format string, args ...interface{}) *ApiNote {
	x := ApiNote{}
	return x.Withf(code, format, args...)
}

// ------------------------------------------------------------------//
// Utils
// ------------------------------------------------------------------//
// ------------------------------------------------------------------//
// GetTimestamp
// ------------------------------------------------------------------//
func GetTimestamp(ival interface{}, format ...string) (string, error) {
	switch v := ival.(type) {
	case map[string]interface{}:
		xa := v
		if xb, ok := xa["TimeAt"]; ok {
			if xc, ok := xb.(int64); ok {
				return GetTimestampB(time.UnixMicro(xc), format...), nil
			}
			return "", fmt.Errorf("ApiNote timestamp error, unexpected TimeAt datatype : %T", xb)
		}
		return "", fmt.Errorf("ApiNote timestamp error, required TimeAt field is undefined")
	case time.Time:
		return GetTimestampB(v, format...), nil
	case string:
		return v, nil
	}
	return "", fmt.Errorf("ApiNote timestamp error, unexpected datatype : %T", ival)
}

// ------------------------------------------------------------------//
// GetTimestampA
// ------------------------------------------------------------------//
func GetTimestampA(formatArg string, i ...interface{}) (*spb.Value, error) {
	if i == nil || i[0] == nil {
		timestamp := GetTimestampB(time.Now(), formatArg)
		return spb.NewStringValue(timestamp), nil
	}
	timestamp := map[string]interface{}{
		"Context": i[0],
		"TimeAt":  time.Now().UnixMicro(),
	}
	return spb.NewValue(timestamp)
}

// ------------------------------------------------------------------//
// GetTimestampB
// ------------------------------------------------------------------//
func GetTimestampB(t time.Time, formatArg ...string) string {
	format := "2006-01-02_15:04:05.000000"
	if formatArg != nil && formatArg[0] != "" {
		format = formatArg[0]
	}
	return t.Format(format)
}
