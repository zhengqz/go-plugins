/*
 *
 * Copyright 2014, Google Inc.
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are
 * met:
 *
 *     * Redistributions of source code must retain the above copyright
 * notice, this list of conditions and the following disclaimer.
 *     * Redistributions in binary form must reproduce the above
 * copyright notice, this list of conditions and the following disclaimer
 * in the documentation and/or other materials provided with the
 * distribution.
 *     * Neither the name of Google Inc. nor the names of its
 * contributors may be used to endorse or promote products derived from
 * this software without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
 * "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
 * LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
 * A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
 * OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
 * SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
 * LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
 * DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
 * THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
 * OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 *
 */

package grpc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/transport"
)

// The format of the payload: compressed or not?
// recvMsg reads a complete gRPC message from the stream.
//
// It returns a flag set to true if message was compressed,
// the message as a byte slice or error if so.
// The caller owns the returned msg memory.
//
// If there is an error, possible values are:
//   * io.EOF, when no messages remain
//   * io.ErrUnexpectedEOF
//   * of type transport.ConnectionError
//   * of type transport.StreamError
// No other error values or types must be returned.
func recvMsg(s *transport.Stream, maxRecvMsgSize int) (bool, []byte, error) {
	isCompressed, msg, err := s.Read(maxRecvMsgSize)
	if err != nil {
		return false, nil, err
	}
	return isCompressed, msg, nil
}

// encode serializes msg and returns a buffer of msg.
// If msg is nil, it generates an empty buffer.
// TODO(ddyihai): eliminate extra Compressor parameter.
func encode(c encoding.Codec, msg interface{}, cp grpc.Compressor, outPayload *stats.OutPayload, compressor encoding.Compressor) ([]byte, error) {
	var (
		b    []byte
		cbuf *bytes.Buffer
	)
	if msg != nil {
		var err error
		b, err = c.Marshal(msg)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "grpc: error while marshaling: %v", err.Error())
		}
		if outPayload != nil {
			outPayload.Payload = msg
			// TODO truncate large payload.
			outPayload.Data = b
			outPayload.Length = len(b)
		}
		if compressor != nil || cp != nil {
			cbuf = new(bytes.Buffer)
			// Has compressor, check Compressor is set by UseCompressor first.
			if compressor != nil {
				z, _ := compressor.Compress(cbuf)
				if _, err := z.Write(b); err != nil {
					return nil, status.Errorf(codes.Internal, "grpc: error while compressing: %v", err.Error())
				}
				z.Close()
			} else {
				// If Compressor is not set by UseCompressor, use default Compressor
				if err := cp.Do(cbuf, b); err != nil {
					return nil, status.Errorf(codes.Internal, "grpc: error while compressing: %v", err.Error())
				}
			}
			b = cbuf.Bytes()
		}
	}
	if uint(len(b)) > math.MaxUint32 {
		return nil, status.Errorf(codes.ResourceExhausted, "grpc: message too large (%d bytes)", len(b))
	}

	if outPayload != nil {
		// A 5 byte gRPC-specific message header will added to this message
		// before it's put on wire.
		outPayload.WireLength = 5 + len(b)
	}
	return b, nil
}

func checkRecvPayload(recvCompress string, haveCompressor bool) *status.Status {
	if recvCompress == "" || recvCompress == encoding.Identity {
		return status.New(codes.Internal, "grpc: compressed flag set with identity or empty encoding")
	}
	if !haveCompressor {
		return status.Newf(codes.Unimplemented, "grpc: Decompressor is not installed for grpc-encoding %q", recvCompress)
	}
	return nil
}

// For the two compressor parameters, both should not be set, but if they are,
// dc takes precedence over compressor.
// TODO(dfawley): wrap the old compressor/decompressor using the new API?
func recv(c encoding.Codec, s *transport.Stream, dc grpc.Decompressor, m interface{}, maxReceiveMessageSize int, inPayload *stats.InPayload, compressor encoding.Compressor) error {
	isCompressed, d, err := recvMsg(s, maxReceiveMessageSize)
	if err != nil {
		return err
	}
	if inPayload != nil {
		inPayload.WireLength = len(d)
	}

	if isCompressed {
		if st := checkRecvPayload(s.RecvCompress(), compressor != nil || dc != nil); st != nil {
			return st.Err()
		}
		// To match legacy behavior, if the decompressor is set by WithDecompressor or RPCDecompressor,
		// use this decompressor as the default.
		if dc != nil {
			d, err = dc.Do(bytes.NewReader(d))
			if err != nil {
				return status.Errorf(codes.Internal, "grpc: failed to decompress the received message %v", err)
			}
		} else {
			dcReader, err := compressor.Decompress(bytes.NewReader(d))
			if err != nil {
				return status.Errorf(codes.Internal, "grpc: failed to decompress the received message %v", err)
			}
			d, err = ioutil.ReadAll(dcReader)
			if err != nil {
				return status.Errorf(codes.Internal, "grpc: failed to decompress the received message %v", err)
			}
		}
		if len(d) > maxReceiveMessageSize {
			// TODO: Revisit the error code. Currently keep it consistent with java
			// implementation.
			return status.Errorf(codes.ResourceExhausted, "grpc: received message larger than max (%d vs. %d)", len(d), maxReceiveMessageSize)
		}
	}
	if err := c.Unmarshal(d, m); err != nil {
		return status.Errorf(codes.Internal, "grpc: failed to unmarshal the received message %v", err)
	}
	if inPayload != nil {
		inPayload.RecvTime = time.Now()
		inPayload.Payload = m
		// TODO truncate large payload.
		inPayload.Data = d
		inPayload.Length = len(d)
	}
	return nil
}

// rpcError defines the status from an RPC.
type rpcError struct {
	code codes.Code
	desc string
}

func (e *rpcError) Error() string {
	return fmt.Sprintf("rpc error: code = %d desc = %s", e.code, e.desc)
}

// Code returns the error code for err if it was produced by the rpc system.
// Otherwise, it returns codes.Unknown.
func Code(err error) codes.Code {
	if err == nil {
		return codes.OK
	}
	if e, ok := err.(*rpcError); ok {
		return e.code
	}
	return codes.Unknown
}

// ErrorDesc returns the error description of err if it was produced by the rpc system.
// Otherwise, it returns err.Error() or empty string when err is nil.
func ErrorDesc(err error) string {
	if err == nil {
		return ""
	}
	if e, ok := err.(*rpcError); ok {
		return e.desc
	}
	return err.Error()
}

// Errorf returns an error containing an error code and a description;
// Errorf returns nil if c is OK.
func Errorf(c codes.Code, format string, a ...interface{}) error {
	if c == codes.OK {
		return nil
	}
	return &rpcError{
		code: c,
		desc: fmt.Sprintf(format, a...),
	}
}

// toRPCErr converts an error into a rpcError.
func toRPCErr(err error) error {
	switch e := err.(type) {
	case *rpcError:
		return err
	case transport.StreamError:
		return &rpcError{
			code: e.Code,
			desc: e.Desc,
		}
	case transport.ConnectionError:
		return &rpcError{
			code: codes.Internal,
			desc: e.Desc,
		}
	default:
		switch err {
		case context.DeadlineExceeded:
			return &rpcError{
				code: codes.DeadlineExceeded,
				desc: err.Error(),
			}
		case context.Canceled:
			return &rpcError{
				code: codes.Canceled,
				desc: err.Error(),
			}
		case grpc.ErrClientConnClosing:
			return &rpcError{
				code: codes.FailedPrecondition,
				desc: err.Error(),
			}
		}

	}
	return Errorf(codes.Unknown, "%v", err)
}

// convertCode converts a standard Go error into its canonical code. Note that
// this is only used to translate the error returned by the server applications.
func convertCode(err error) codes.Code {
	switch err {
	case nil:
		return codes.OK
	case io.EOF:
		return codes.OutOfRange
	case io.ErrClosedPipe, io.ErrNoProgress, io.ErrShortBuffer, io.ErrShortWrite, io.ErrUnexpectedEOF:
		return codes.FailedPrecondition
	case os.ErrInvalid:
		return codes.InvalidArgument
	case context.Canceled:
		return codes.Canceled
	case context.DeadlineExceeded:
		return codes.DeadlineExceeded
	}
	switch {
	case os.IsExist(err):
		return codes.AlreadyExists
	case os.IsNotExist(err):
		return codes.NotFound
	case os.IsPermission(err):
		return codes.PermissionDenied
	}
	return codes.Unknown
}

func wait(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	wait, ok := ctx.Value("wait").(bool)
	if !ok {
		return false
	}
	return wait
}
