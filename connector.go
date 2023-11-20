// The MIT License
//
// Copyright (c) 2022 Peter A McGill
package transport

import (
	"fmt"
	"net"
)

//===================================================================//
// Connex -
//===================================================================//
type Connex struct {
  net.Conn
  cid string
}

// -------------------------------------------------------------- //
// Cid
// ---------------------------------------------------------------//
func (cx Connex) Cid() string {
  return cx.cid
}

// -------------------------------------------------------------- //
// NewConnex
// ---------------------------------------------------------------//
func NewConnex(conn net.Conn, cid string) Connex {
  return Connex{
    Conn: conn,
    cid: cid,
  }
}

//===================================================================//
// ApiConnex - could accomodate an in-memory net.Pipe - for embedded
// app components that want to emulate tcpconn reading and writing
//===================================================================//
type ApiConnex struct {
  net.Conn
  ctype ConnType
  cid string
}

// -------------------------------------------------------------- //
// Cid
// ---------------------------------------------------------------//
func (cx *ApiConnex) ConnId() string {
  return cx.cid
}

// -------------------------------------------------------------- //
// ConnType
// ---------------------------------------------------------------//
func (cx *ApiConnex) ConnType() string {
  return cx.ctype.String()
}

//----------------------------------------------------------------//
// RecvFrame
//----------------------------------------------------------------//
func (cx *ApiConnex) RecvFrame() ([]byte, error) {
  _, frame, err := ReadWH(cx)
  return frame, err
}

//----------------------------------------------------------------//
// RecvStream
//----------------------------------------------------------------//
func (cx *ApiConnex) RecvStream(chunkCh chan []byte, errCh chan error) {
  for {
    flag, chunk, err := ReadWH(cx)
    if err != nil {
      errCh<- fmt.Errorf("%s - receive stream error : %w", cx.cid, err)
      return
    }
    chunkCh <- chunk
    if !flag.HasMore() {
      close(chunkCh)
      return
    }
  }
}

//----------------------------------------------------------------//
// SendFrame
//----------------------------------------------------------------//
func (cx *ApiConnex) SendFrame(frame []byte, flag byte) error {
  return WriteWH(cx, frame, flag)
}

// -------------------------------------------------------------- //
// Tell
// ---------------------------------------------------------------//
func (cx *ApiConnex) Tell() (string, string, string) {
  return cx.cid, cx.LocalAddr().String(), cx.RemoteAddr().String()
}