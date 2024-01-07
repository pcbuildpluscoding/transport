package test

import (
	"fmt"
	"log"
	"net"
	"time"

	stx "github.com/pcbuildpluscoding/strucex/std"
	tpt "github.com/pcbuildpluscoding/transport"
	rdt "github.com/pcbuildpluscoding/types/apirecord"
	rwt "github.com/pcbuildpluscoding/types/runware"
)

// ================================================================//
// NewTestClient
// ================================================================//
func NewTestClient(rw *stx.Strucex) (*TestClient, error) {
	desc := "TestClient-" + time.Now().Format("150405.000000")

	logger.Debugf("creating %s ...", desc)

	timeout := rw.Int("ReadTimeout")
	if timeout <= 0 {
		return nil, fmt.Errorf("read timeout |%d| is not > 0", timeout)
	}
	readTimeout := time.Duration(timeout) * time.Second

	timeout = rw.Int("WriteTimeout")
	if timeout <= 0 {
		return nil, fmt.Errorf("write timeout |%d| is not > 0", timeout)
	}
	writeTimeout := time.Duration(timeout) * time.Second

	c := &TestClient{
		Desc:         desc,
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
	}

	return c, nil
}

// ==================================================================//
// LineCache
// ==================================================================//
type LineCache struct {
	this []byte
}

func (c *LineCache) add(line string) {
	c.this = append(c.this, []byte(line)...)
}

func (c *LineCache) flush() []byte {
	x := c.this
	c.this = []byte{}
	return x
}

func (c *LineCache) size() int {
	return len(c.this)
}

// ==================================================================//
// TestClient
// ==================================================================//
type TestClient struct {
	tpt.ApiResult
	Desc        string
	conn        net.Conn
	readTimeout time.Duration
	// state        cwt.Activity
	writeTimeout time.Duration
	writeCount   int
	writeTotal   int
}

// -------------------------------------------------------------- //
// AsyncRequest
// ---------------------------------------------------------------//
func (c *TestClient) AsyncRequest(req rwt.Runware) error {
	if c.conn == nil {
		return fmt.Errorf("connection is closed")
		// return net.ErrClosed
	}
	frame, err := req.Encode()
	if err != nil {
		return fmt.Errorf("request encoding failed : %v", err)
	}
	c.writeCount += 1
	c.writeTotal += len(frame)
	logger.Debugf("sending frame with size, running total and writeCount : %d, %d, %d", len(frame), c.writeTotal, c.writeCount)
	err = c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
	if err != nil {
		return fmt.Errorf("%s setWriteDeadline error : %v", c.Desc, err)
	}
	return tpt.WriteWH(c.conn, frame, 0)
}

// -------------------------------------------------------------- //
// Close
// ---------------------------------------------------------------//
func (c *TestClient) Close(jobId string) error {
	logger.Debugf("%s closing ...", c.Desc)
	if c.conn != nil {
		if err := c.conn.SetDeadline(time.Time{}); err == nil {
			logger.Debugf("%s closing the connection ...", c.Desc)
			req, _ := stx.NewRunware(nil)
			_ = req.Set("JobId", jobId)
			_ = req.Set("Action", "Complete")
			if err := c.AsyncRequest(req); err != nil {
				logger.Errorf("close request error : %v", err)
			}
			return c.conn.Close()
		}
	}
	return nil
}

// ----------------------------------------------------------------//
// checkResponse
// ----------------------------------------------------------------//
func (c *TestClient) checkResponse(resp rdt.ApiRecord, expected string) error {
	x := toApiNote(resp)
	if err := x.Unwrap(); err != nil {
		return err
	} else if advice, _ := x.Value().(string); advice == "" {
		return fmt.Errorf("response is empty")
	} else if advice != expected {
		return fmt.Errorf("response advice != %s - got %s instead", expected, advice)
	}
	return nil
}

// -------------------------------------------------------------- //
// getReply
// ---------------------------------------------------------------//
func (c *TestClient) getReply() rdt.ApiRecord {
	err := c.conn.SetReadDeadline(time.Now().Add(c.readTimeout))
	if err != nil {
		return c.Withf(500, "%s SetReadDeadline error : %v", c.Desc, err)
	}
	_, frame, err := tpt.ReadWH(c.conn)
	if err != nil {
		// first set the data field then the code and error
		return c.With(400, err)
	}
	y, err := tpt.NewApiRecord(frame)
	if err != nil {
		return c.Withf(500, "response decode error : %v", err)
	}
	return y
}

// -------------------------------------------------------------- //
// Init
// ---------------------------------------------------------------//
func (c *TestClient) Init(conn net.Conn, req *stx.Strucex) error {
	c.conn = conn
	if bsize := req.Int("WriteBufferSize"); bsize > 0 {
		err := c.setWriteBuffer(bsize)
		if err != nil {
			return err
		}
	}
	resp := c.Request(req.SubNode("Request", false))
	logger.Debugf("%s got handshake response : %v", c.Desc, resp.Unwrap())
	if err := c.checkResponse(resp, "ok"); err != nil {
		return err
	}
	err := c.conn.SetReadDeadline(time.Time{})
	if err != nil {
		return err
	}
	return nil
}

// -------------------------------------------------------------- //
// Request
// ---------------------------------------------------------------//
func (c *TestClient) Request(req rwt.Runware) rdt.ApiRecord {
	err := c.AsyncRequest(req)
	if err != nil {
		// first set the data field, that the code and error
		return c.With(400, err)
	}
	// request is a synchronous operation
	return c.getReply()
}

// -------------------------------------------------------------- //
// ResetVars
// ---------------------------------------------------------------//
func (c *TestClient) ResetVars() {
	c.writeCount = 0
	c.writeTotal = 0
}

// -------------------------------------------------------------- //
// setWriteBuffer
// ---------------------------------------------------------------//
func (c *TestClient) setWriteBuffer(bsize int) error {
	if c.conn == nil {
		return fmt.Errorf("conn is undefined")
	}
	logger.Infof("%s setting write buffer size : %d", c.Desc, bsize)
	err := c.conn.(*net.TCPConn).SetWriteBuffer(bsize)
	if err != nil {
		return fmt.Errorf("%s setWriteBuffer error : %v", c.Desc, err)
	}
	return nil
}

// -------------------------------------------------------------- //
// toApiNote
// ---------------------------------------------------------------//
func toApiNote(rcd rdt.ApiRecord) *tpt.ApiNote {
	switch x := rcd.(type) {
	case *tpt.ApiNote:
		return x
	default:
		log.Panicf("ApiRecord value type is not *ApiNote")
	}
	return nil
}

// -------------------------------------------------------------- //
// toStrucex
// ---------------------------------------------------------------//
func toStrucex(req rwt.Runware) *stx.Strucex {
	switch x := req.(type) {
	case *stx.Strucex:
		return x
	default:
		log.Panicf("Runware value type is not *Strucex")
	}
	return nil
}
