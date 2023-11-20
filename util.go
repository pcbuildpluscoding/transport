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

	"github.com/sirupsen/logrus"
)

//==================================================================//
//  Deque
//==================================================================//
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
  err, ok := v.Interface().(error)
  return err, ok
}

// -------------------------------------------------------------- //
// HasErrKind
// ---------------------------------------------------------------//
func HasErrKind(r interface{}) (err error, isErr bool) {
  err = nil
  isErr = false
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
      fmt.Fprintf(logfd,"%s\n\n%s\n", r, debug.Stack())
    }
  }
}

//==================================================================//
//  ApiError
//==================================================================//
type ApiError struct {
  code int
  appKey string
  err error
}

//------------------------------------------------------------------//
// As
//------------------------------------------------------------------//
func (e ApiError) As(target interface{}) bool {
  return ErrorAs(e.err, target)
}

//------------------------------------------------------------------//
// Code
//------------------------------------------------------------------//
func (e ApiError) Code() int {
  return e.code
}

//------------------------------------------------------------------//
// AppKey
//------------------------------------------------------------------//
func (e ApiError) AppKey() string {
  return e.appKey
}

//------------------------------------------------------------------//
// Error
//------------------------------------------------------------------//
func (e ApiError) Error() string {
  return fmt.Sprintf("%s error[%d] : %s", e.appKey, e.code, e.err.Error())
}

//------------------------------------------------------------------//
// Is
//------------------------------------------------------------------//
func (e ApiError) Is(target error) bool {
  x, ok := target.(*ApiError)
  if !ok {
    return false
  }
  return e.code == x.code && e.appKey == x.appKey && errors.Is(e.err, x.err)
}

//------------------------------------------------------------------//
// Unwrap
//------------------------------------------------------------------//
func (n ApiError) Unwrap() error {
  return n.err
}

//------------------------------------------------------------------//
// ErrorAs
//------------------------------------------------------------------//
func ErrorAs(err error, i interface{}) bool {
  v := reflect.ValueOf(i)
  switch v.Kind() {
    case reflect.Ptr:
      // terr, isErr := hasErrIface(v)
      _, isErr := hasErrIface(v)
      if isErr {
        if reflect.TypeOf(err).Name() == v.Elem().Type().Name() {
          // logger.Infof("################### err.TypeOf.Name comparison matches ####################")
          if v.CanSet() {
            v.Set(reflect.ValueOf(err))
            return true
          }
          return false
        }
        // logger.Infof("################### errors.As matches ####################")
        return errors.As(err, i)
      }
  }
  return false
}
