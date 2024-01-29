// The MIT License
//
// Copyright (c) 2022 Peter A McGill
package transport

import (
	"encoding/base64"
	"errors"
	"fmt"

	stx "github.com/pcbuildpluscoding/strucex/std"
	spb "google.golang.org/protobuf/types/known/structpb"
)

// ================================================================//
// ApiNote
// ================================================================//
type ApiNote struct {
	code int
	err  error
	data interface{}
}

// ------------------------------------------------------------------//
// As
// ------------------------------------------------------------------//
func (n *ApiNote) As(target interface{}) bool {
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

	if n.data == nil {
		return x
	}

	var err error
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
// Copy
// ----------------------------------------------------------------//
func (n *ApiNote) Copy() *ApiNote {
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
	if x, ok := vm["Error"]; ok {
		if errtxt := x.GetStringValue(); errtxt != "" {
			n.err = errors.New(errtxt)
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
func (n *ApiNote) SetData(ival interface{}, strict ...bool) error {
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
		n.data = d.AsStruct()
	default:
		if strict != nil && !strict[0] {
			n.data = ival
			return nil
		}
		n.data, err = spb.NewValue(d)
	}
	if err != nil {
		return fmt.Errorf("%T is not structpb compatible : %v", ival, err)
	}
	return nil
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
func (n *ApiNote) With(code int, union interface{}, data ...interface{}) *ApiNote {
	f := func(d interface{}) {
		err := n.SetData(d)
		if err != nil {
			if n.err != nil {
				n.err = fmt.Errorf("data error : %v : %w", err, n.err)
			} else {
				n.err = fmt.Errorf("data error : %v", err)
			}
		}
	}
	n.code = code
	switch v := union.(type) {
	case nil:
	case error:
		n.err = v
	default:
		f(union)
	}
	if data != nil {
		if data[0] == nil {
			n.data = nil
		}
		switch v := data[0].(type) {
		case error: // not expected but must be detected
			n.data = v.Error()
		default:
			f(data[0])
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
func (ApiResult) CheckErr(errcode int, err error, success_code ...int) *ApiNote {
	x := &ApiNote{code: 200}
	if err != nil {
		x.code = errcode
		x.err = err
	} else if success_code != nil {
		x.code = success_code[0]
	}
	return x
}

// ----------------------------------------------------------------//
// WithCode
// ----------------------------------------------------------------//
func (ApiResult) With(code int, union interface{}, data ...interface{}) *ApiNote {
	x := ApiNote{}
	return x.With(code, union, data...)
}

// ----------------------------------------------------------------//
// Withf
// ----------------------------------------------------------------//
func (ApiResult) Withf(code int, format string, args ...interface{}) *ApiNote {
	x := ApiNote{}
	return x.Withf(code, format, args...)
}
