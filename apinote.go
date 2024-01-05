// The MIT License
//
// Copyright (c) 2022 Peter A McGill
package transport

import (
	"encoding/base64"
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
		panic(fmt.Errorf("ApiNote timestamp value is not a valid structpb type : %v", err))
	}

	switch d := n.data.(type) {
	case *spb.ListValue:
		x["Data"] = spb.NewListValue(d)
	case *Strucex:
		x["Data"] = spb.NewStructValue(d.AsStruct())
	case *spb.Struct:
		x["Data"] = spb.NewStructValue(d)
	case *spb.Value:
		x["Data"] = d
	default:
		x["Data"], err = spb.NewValue(d)
	}

	if err != nil {
		panic(fmt.Errorf("ApiNote data value is not a valid structpb type : %v", err))
	}

	return x
}

// ----------------------------------------------------------------//
// Body
// ----------------------------------------------------------------//
func (n *ApiNote) Bytes() ([]byte, error) {
	switch x := n.data.(type) {
	case *spb.Value:
		return base64.StdEncoding.DecodeString(x.GetStringValue())
	case []byte:
		return x, nil
	default:
		return nil, fmt.Errorf("Record data type |%T| is not []byte", n.data)
	}
}

// ----------------------------------------------------------------//
// Code
// ----------------------------------------------------------------//
func (n *ApiNote) Code() int {
	return n.code
}

// ----------------------------------------------------------------//
// copyData
// ----------------------------------------------------------------//
func copyData(ival interface{}) (interface{}, error) {
	switch d := ival.(type) {
	case *stx.Strucex:
		frame, err := d.AsStruct().MarshalJSON()
		if err != nil {
			return nil, err
		}
		return stx.NewRunware(frame)
	case *spb.Struct:
		frame, err := d.MarshalJSON()
		if err != nil {
			return nil, err
		}
		data := spb.Struct{}
		err = data.UnmarshalJSON(frame)
		return &data, err
	case *spb.Value:
		data := spb.Value{}
		frame, err := d.MarshalJSON()
		if err != nil {
			return nil, err
		}
		err = data.UnmarshalJSON(frame)
		return &data, err
	}
	return nil, fmt.Errorf("data type %T is not structpb compatible", ival)
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
					key:  y.Fields["Key"].GetStringValue(),
					err:  err,
				}
			}
		}
	}
	n.data = vm["Data"]
}

// ------------------------------------------------------------------//
// Error
// ------------------------------------------------------------------//
func (n ApiNote) Error() string {
	if n.err != nil {
		return n.err.Error()
	}
	return ""
}

// -------------------------------------------------------------- //
// Empty
// ---------------------------------------------------------------//
func (n ApiNote) Empty() bool {
	return n.code == 0 && n.err == nil && n.data == nil
}

// ----------------------------------------------------------------//
// Hardcopy
// ----------------------------------------------------------------//
func (n *ApiNote) Hardcopy() *ApiNote {
	c := &ApiNote{
		code: n.code,
		err:  n.err,
	}
	var err error
	c.data, err = copyData(n.data)
	if err != nil {
		if c.err != nil {
			c.err = fmt.Errorf("data error : %v : %w", err, c.err)
		}
	}
	return c
}

// ------------------------------------------------------------------//
// Is
// ------------------------------------------------------------------//
func (n ApiNote) Is(target error) bool {
	if errors.Is(n.err, target) {
		return true
	} else if n.err != nil && target != nil {
		return n.err.Error() == target.Error()
	}
	return false
}

// -------------------------------------------------------------- //
// Parameter
// ---------------------------------------------------------------//
func (n ApiNote) Parameter() *Parametric {
	return stx.NewParameter("-", n.data)
}

// -------------------------------------------------------------- //
// Runware
// ---------------------------------------------------------------//
func (n ApiNote) Runware() *Strucex {
	rw, _ := stx.NewRunware(n.data)
	return rw
}

// ----------------------------------------------------------------//
// SetData
// ----------------------------------------------------------------//
func (n *ApiNote) SetData(data interface{}) error {
	return n.setData(data)
}

// ----------------------------------------------------------------//
// setData
// ----------------------------------------------------------------//
func (n *ApiNote) setData(ival interface{}) error {
	var err error
	switch d := ival.(type) {
	case nil:
		n.data = nil
	case []string:
		data := make([]interface{}, len(d))
		for i, v := range d {
			data[i] = v
		}
		n.data, err = spb.NewValue(data)
	case *stx.Strucex:
		n.data = d.AsStruct()
	case *spb.Struct, *spb.Value:
		n.data = d
	case stx.Strucex:
		n.data = &d
	default:
		n.data, err = spb.NewValue(d)
	}
	if err != nil {
		return fmt.Errorf("%T is not structpb compatible : %v", ival, err)
	}
	return nil
}

// ----------------------------------------------------------------//
// SetTimestamp
// ----------------------------------------------------------------//
func (n *ApiNote) SetTimestamp() {
	n.timestamp = time.Now()
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
	switch v := n.data.(type) {
	case *spb.Value:
		return v.AsInterface()
	default:
		return v
	}
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
		err := n.setData(data)
		if err != nil {
			if n.err != nil {
				n.err = fmt.Errorf("%v\ndata error : %v", n.err, err)
			} else {
				n.err = err
			}
		}
	}
	return n
}

// ----------------------------------------------------------------//
// WithErr
// ----------------------------------------------------------------//
func (n *ApiNote) WithErr(err error, code ...int) *ApiNote {
	n.code = 200
	if err != nil {
		n.err = err
		n.code = 400
		if code != nil && code[0] > 400 {
			n.code = code[0]
		}
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
// CheckErr
// ----------------------------------------------------------------//
func (ApiResult) CheckErr(err error, code ...int) *ApiNote {
	x := &ApiNote{code: 200}
	if err != nil {
		x.err = err
		if code != nil {
			x.code = code[0]
		}
	}
	return x
}

// ----------------------------------------------------------------//
// WithCode
// ----------------------------------------------------------------//
func (ApiResult) With(code int, data interface{}) *ApiNote {
	x := ApiNote{}
	return x.With(code, data)
}

// ----------------------------------------------------------------//
// WithErr
// ----------------------------------------------------------------//
func (ApiResult) WithErr(err error, code ...int) *ApiNote {
	x := ApiNote{}
	return x.WithErr(err, code...)
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
