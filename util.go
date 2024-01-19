// The MIT License
//
// Copyright (c) 2022 Peter A McGill
package transport

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"runtime/debug"

	stx "github.com/pcbuildpluscoding/strucex/std"
	"github.com/sirupsen/logrus"
	spb "google.golang.org/protobuf/types/known/structpb"
)

// ================================================================//
// Deque
// ================================================================//
type Deque struct {
	this []interface{}
}

// -------------------------------------------------------------- //
// Delete
// ---------------------------------------------------------------//
func (dq *Deque) Delete(obj interface{}) error {
	found := false
	i := 0
	var item interface{}
	for i, item = range dq.this {
		if item == obj {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("delete failed, obj %v is not found", obj)
	}
	switch i {
	case 0:
		dq.this = dq.this[1:]
	default:
		dq.this = append(dq.this[:i], dq.this[i+1:]...)
	}
	return nil
}

// -------------------------------------------------------------- //
// Destroy
// ---------------------------------------------------------------//
func (dq *Deque) Destroy() {
	dq.this = nil
}

// -------------------------------------------------------------- //
// Empty
// ---------------------------------------------------------------//
func (dq *Deque) Empty() bool {
	return len(dq.this) == 0
}

// -------------------------------------------------------------- //
// Get
// ---------------------------------------------------------------//
func (dq *Deque) Get(i int) (interface{}, error) {
	if dq.Empty() {
		return nil, fmt.Errorf("list is empty")
	} else if i < 0 || i >= len(dq.this) {
		return nil, fmt.Errorf("invalid index")
	}
	return dq.this[i], nil
}

// -------------------------------------------------------------- //
// Insert
// ---------------------------------------------------------------//
func (dq *Deque) Insert(i int, item interface{}) error {
	if dq.Empty() && i != 0 {
		return fmt.Errorf("invalid index")
	} else if i < 0 || i > len(dq.this) {
		return fmt.Errorf("invalid index")
	}
	_item := []interface{}{item}
	temp := append(_item, dq.this[i:]...)
	dq.this = append(dq.this[:i], temp...)
	return nil
}

// -------------------------------------------------------------- //
// PopLeft
// ---------------------------------------------------------------//
func (dq *Deque) PopLeft() interface{} {
	if dq.Empty() {
		return nil
	}
	item := dq.this[0]
	dq.this = dq.this[1:]
	return item
}

// -------------------------------------------------------------- //
// PopRight
// ---------------------------------------------------------------//
func (dq *Deque) PopRight() interface{} {
	if dq.Empty() {
		return nil
	}
	last := len(dq.this) - 1
	item := dq.this[last]
	dq.this = dq.this[:last]
	return item
}

// -------------------------------------------------------------- //
// PushLeft
// ---------------------------------------------------------------//
func (dq *Deque) PushLeft(item interface{}) {
	_item := []interface{}{item}
	dq.this = append(_item, dq.this...)
}

// -------------------------------------------------------------- //
// PushRight
// ---------------------------------------------------------------//
func (dq *Deque) PushRight(item interface{}) {
	dq.this = append(dq.this, item)
}

// -------------------------------------------------------------- //
// Reset
// ---------------------------------------------------------------//
func (dq *Deque) Reset() {
	dq.this = []interface{}{}
}

// -------------------------------------------------------------- //
// Set
// ---------------------------------------------------------------//
func (dq *Deque) Set(i int, item interface{}) error {
	if dq.Empty() {
		return fmt.Errorf("list is empty")
	}
	if i < 0 || i >= len(dq.this) {
		return fmt.Errorf("invalid index")
	}
	dq.this[i] = item
	return nil
}

// -------------------------------------------------------------- //
// Size()
// ---------------------------------------------------------------//
func (dq *Deque) Size() int {
	return len(dq.this)
}

// -------------------------------------------------------------- //
// String
// ---------------------------------------------------------------//
func (dq *Deque) String() string {
	return fmt.Sprintf("%v", dq.this)
}

// -------------------------------------------------------------- //
// ErrCatcher
// ---------------------------------------------------------------//
func ErrCatcher(logger *logrus.Logger, errCh chan error) func(error) {
	f := func(err error) {
		logger.Errorf("panic error caught => %v\n%s", err, string(debug.Stack()))
		errCh <- err
	}
	return f
}

// -------------------------------------------------------------- //
// hasErrIface -
// ---------------------------------------------------------------//
func hasErrIface(v reflect.Value) (error, bool) {
	// CanInterface reports whether Interface can be used without panicking
	if !v.CanInterface() {
		return nil, false
	}
	// Interface panics if the Value was obtained by accessing unexported struct fields
	switch v.Kind() {
	case reflect.Ptr:
		err, ok := v.Interface().(*error)
		return *err, ok
	}
	err, ok := v.Interface().(error)
	return err, ok
}

// -------------------------------------------------------------- //
// HasErrKind
// ---------------------------------------------------------------//
func HasErrKind(r interface{}) (err error, isErr bool) {
	v := reflect.ValueOf(r)
	switch v.Kind() {
	case reflect.Struct:
		errtype := reflect.TypeOf((*error)(nil)).Elem()
		if v.Type().Implements(errtype) {
			err, isErr = v.Interface().(error)
		}
	case reflect.Ptr:
		err, isErr = hasErrIface(v)
	case reflect.Interface:
		err, isErr = hasErrIface(v)
	}
	return
}

// -------------------------------------------------------------- //
// EvalErrKind
// ---------------------------------------------------------------//
func EvalErrKind(r interface{}) (errval error) {
	err, isErr := HasErrKind(r)
	if !isErr {
		errtxt := "Unknown system error - %v :\n%v"
		v := reflect.ValueOf(r)
		return fmt.Errorf(errtxt, v.Type(), v)
	}
	return err
}

// -------------------------------------------------------------- //
// PanicHandler
// ---------------------------------------------------------------//
func PanicHandler(errHandler func(error), fd ...*os.File) func() {
	return func() {
		if r := recover(); r != nil {
			logfd := os.Stdout
			if fd != nil && fd[0] != nil {
				logfd = fd[0]
			}
			if errHandler != nil {
				errHandler(EvalErrKind(r))
			}
			fmt.Fprintf(logfd, "%s\n\n%s\n", r, debug.Stack())
		}
	}
}

// ------------------------------------------------------------------//
// ErrorAs
// ------------------------------------------------------------------//
func ErrorAs(err error, i interface{}) bool {
	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Ptr:
		return errors.As(err, i)
	}
	return false
}

// ==================================================================//
// FlowRule
// ==================================================================//
type FlowRule map[string]interface{}

// ------------------------------------------------------------------//
// Add
// ------------------------------------------------------------------//
func (r FlowRule) Add(key string, value interface{}) {
	r[key] = value
}

// ----------------------------------------------------------------//
// Runware
// ----------------------------------------------------------------//
func (r FlowRule) AsMap() map[string]interface{} {
	return r
}

// ----------------------------------------------------------------//
// AsRunware
// ----------------------------------------------------------------//
func (r FlowRule) AsRunware() (*Strucex, error) {
	x := stx.Strucex{}
	err := x.Decode(r.AsMap())
	return &x, err
}

// ----------------------------------------------------------------//
// AsStruct
// ----------------------------------------------------------------//
func (r FlowRule) AsStruct() (*spb.Struct, error) {
	s, err := spb.NewValue(r.AsMap())
	return s.GetStructValue(), err
}

// ------------------------------------------------------------------//
// Bool
// ------------------------------------------------------------------//
func (r FlowRule) Bool(key string) bool {
	x, _ := r[key].(bool)
	return x
}

// ------------------------------------------------------------------//
// Copy
// ------------------------------------------------------------------//
func (r FlowRule) Copy() FlowRule {
	x := FlowRule{}
	for k, v := range r {
		x.Add(k, v)
	}
	return x
}

// ------------------------------------------------------------------//
// Float
// ------------------------------------------------------------------//
func (r FlowRule) Float(key string) float64 {
	switch x := r[key].(type) {
	case nil:
	case float64:
		return x
	case int:
		return float64(x)
	}
	return 0
}

// ------------------------------------------------------------------//
// Int
// ------------------------------------------------------------------//
func (r FlowRule) Int(key string) int {
	switch x := r[key].(type) {
	case nil:
	case float64:
		return int(x)
	case int:
		return x
	}
	return 0
}

// ------------------------------------------------------------------//
// ParamList
// ------------------------------------------------------------------//
func (r FlowRule) ParamList(key string) ([]*Parametric, error) {
	w, found := r[key]
	if !found {
		return []*Parametric{}, fmt.Errorf("missing key : " + key)
	}
	return toParamList(key, w)
}

// ------------------------------------------------------------------//
// Pop
// ------------------------------------------------------------------//
func (r FlowRule) Pop(key string) interface{} {
	x := r[key]
	delete(r, key)
	return x
}

// ----------------------------------------------------------------//
// Runware
// ----------------------------------------------------------------//
func (r FlowRule) Runware(key string) (*Strucex, error) {
	x, found := r[key]
	if !found {
		return nil, fmt.Errorf("missing key : " + key)
	}
	s := stx.Strucex{}
	err := s.Decode(x)
	return &s, err
}

// ------------------------------------------------------------------//
// String
// ------------------------------------------------------------------//
func (r FlowRule) String(key string) string {
	x, _ := r[key].(string)
	return x
}

// ------------------------------------------------------------------//
// StringList
// ------------------------------------------------------------------//
func (r FlowRule) StringList(key string) []string {
	w, found := r[key]
	if !found {
		return []string{}
	}
	switch x := w.(type) {
	case []string:
		return x
	case []interface{}:
		return toStringList(x)
	}
	return []string{}
}

// ------------------------------------------------------------------//
// Value
// ------------------------------------------------------------------//
func (r FlowRule) Value(key string) interface{} {
	x, _ := r[key]
	return x
}

// ----------------------------------------------------------------//
// utils
// ----------------------------------------------------------------//
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
// toParamList
// ----------------------------------------------------------------//
func toParamList(key string, ival interface{}) ([]*Parametric, error) {
	var args []interface{}
	switch val := ival.(type) {
	case []interface{}:
		args = val
	case *spb.Value:
		switch x := val.GetKind().(type) {
		case *spb.Value_ListValue:
			args = validList(x.ListValue)
		default:
			return []*Parametric{}, fmt.Errorf("structpb.List type is required")
		}
	case *spb.ListValue:
		args = validList(val)
	default:
		return []*Parametric{}, fmt.Errorf("structpb.List type is required")
	}
	x := make([]*Parametric, len(args))
	for i, value := range args {
		id := fmt.Sprintf("%s[%d]", key, i)
		x[i] = stx.NewParameter(id, value)
	}
	return x, nil
}

// -------------------------------------------------------------- //
// validList
// ---------------------------------------------------------------//
func validList(x *spb.ListValue) []interface{} {
	if x != nil {
		return x.AsSlice()
	}
	return []interface{}{}
}
