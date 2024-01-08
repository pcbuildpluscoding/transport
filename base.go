// The MIT License
//
// Copyright (c) 2022 Peter A McGill
package transport

import (
	"net"
	"reflect"
	"strings"
	"time"

	"github.com/pcbuildpluscoding/logroll"
	stx "github.com/pcbuildpluscoding/strucex/std"
	fwt "github.com/pcbuildpluscoding/types/flowware"
	"github.com/sirupsen/logrus"
	spb "google.golang.org/protobuf/types/known/structpb"
)

// TO DO - delete these - they are moved to types/appware
var (
	logger                = logroll.Get()
	NanoDatestamp  string = "2006-01-02-15-04-05.000000"
	MilliDatestamp string = "2006-01-02-15-04-05.000"
	SecDatestamp   string = "2006-01-02-15-04-05"
	DayTimestamp   string = "02-15-04-05.000000"
)

// -------------------------------------------------------------- //
// SetLogger
// ---------------------------------------------------------------//
func SetLogger(super *logrus.Logger) {
	logger = super
}

type Parametric = stx.Parametric
type Strucex = stx.Strucex
type ConnType = fwt.ConnType

// ------------------------------------------------------------------//
// NewApiConnex
// ------------------------------------------------------------------//
func NewApiConnex(conn net.Conn, ctype fwt.ConnType, cid string) ApiConnex {
	return ApiConnex{
		Conn:  conn,
		ctype: ctype,
		cid:   cid,
	}
}

// ------------------------------------------------------------------//
// NewPipeConnex
// ------------------------------------------------------------------//
func NewPipeConnex(args ...string) (ApiConnex, ApiConnex) {
	suffix := time.Now().Format("150405.000000")
	if args != nil {
		suffix = args[0]
	}
	readcid := "PIPE-R-" + suffix
	writecid := "PIPE-W-" + suffix
	readconn, writeconn := net.Pipe()
	return ApiConnex{
			readconn,
			fwt.Stream_Client,
			readcid,
		},
		ApiConnex{
			writeconn,
			fwt.Stream_Service,
			writecid,
		}
}

// ------------------------------------------------------------------//
// HostIPAddr
// ------------------------------------------------------------------//
func HostIPAddr(ipdesc string) string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic("stdlib.net failed to return local network interface info")
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			switch strings.ToLower(ipdesc) {
			case "ip6", "ipv6":
				if ipnet.IP.To4() == nil {
					return ipnet.IP.String()
				}
			default:
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	}
	return "127.0.0.1"
}

// ----------------------------------------------------------------//
// Constructors
// ----------------------------------------------------------------//
// ----------------------------------------------------------------//
// NewApiResult
// ----------------------------------------------------------------//
func NewApiResult() ApiResult {
	return ApiResult{}
}

// ----------------------------------------------------------------//
// TypeName
// ----------------------------------------------------------------//
func TypeName(obj interface{}) string {
	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Ptr {
		return "*" + t.Elem().Name()
	}
	return t.Name()
}

// -------------------------------------------------------------- //
// NewApiRecord
// ---------------------------------------------------------------//
func NewApiRecord(ival interface{}) (*ApiNote, error) {
	x := ApiNote{}
	if ival == nil {
		return &x, nil
	}
	err := stx.Importe(ival, func(y map[string]*spb.Value) error {
		x.fromMap(y)
		return nil
	})
	return &x, err
}
