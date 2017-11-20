/**
 *
 * @author  chosen0ne(louzhenlin86@126.com)
 * @date    2017-03-17 18:49:53
 */

package redisproxy

import (
	"bytes"
	"io"
	"strconv"
)

var (
	newLine      = []byte{'\r', '\n'}
	OkResponse   = NewSimpleString("OK")
	NullResponse = BytesResponse([]byte("$-1\r\n"))
	pongResponse = NewSimpleString("PONG")
)

type Response interface {
	encode() ([]byte, error)
}

type BytesResponse []byte

// raw string as a Response
func (b BytesResponse) encode() ([]byte, error) {
	return b, nil
}

func responseTo(o io.Writer, r Response) (int, error) {
	oBytes, err := r.encode()
	if err != nil {
		return 0, err
	}

	if n, err := o.Write(oBytes); err != nil {
		return n, err
	}

	return 0, nil
}

type IntResponse struct {
	Val int
}

func NewInt(v int) Response {
	return &IntResponse{v}
}

func (r *IntResponse) encode() ([]byte, error) {
	var o bytes.Buffer
	if err := o.WriteByte(':'); err != nil {
		return nil, err
	}
	if _, err := o.WriteString(strconv.Itoa(r.Val)); err != nil {
		return nil, err
	}
	if _, err := o.Write(newLine); err != nil {
		return nil, err
	}

	return o.Bytes(), nil
}

type SimpleStringResponse struct {
	Val string
}

func NewSimpleString(v string) Response {
	return &SimpleStringResponse{v}
}

func (r *SimpleStringResponse) encode() ([]byte, error) {
	var o bytes.Buffer
	if err := o.WriteByte('+'); err != nil {
		return nil, err
	}
	if _, err := o.WriteString(r.Val); err != nil {
		return nil, err
	}
	if _, err := o.Write(newLine); err != nil {
		return nil, err
	}

	return o.Bytes(), nil
}

type ErrResponse struct {
	Val string
}

func NewError(v string) Response {
	return &ErrResponse{v}
}

func (r *ErrResponse) encode() ([]byte, error) {
	var o bytes.Buffer
	if err := o.WriteByte('-'); err != nil {
		return nil, err
	}
	if _, err := o.WriteString(r.Val); err != nil {
		return nil, err
	}
	if _, err := o.Write(newLine); err != nil {
		return nil, err
	}

	return o.Bytes(), nil
}

type BulkStringResponse struct {
	Val string
}

func NewBulkString(v string) Response {
	return &BulkStringResponse{v}
}

func (r *BulkStringResponse) encode() ([]byte, error) {
	var o bytes.Buffer

	if err := o.WriteByte('$'); err != nil {
		return nil, err
	}
	if _, err := o.WriteString(strconv.Itoa(len(r.Val))); err != nil {
		return nil, err
	}
	if _, err := o.Write(newLine); err != nil {
		return nil, err
	}
	if _, err := o.WriteString(r.Val); err != nil {
		return nil, err
	}
	if _, err := o.Write(newLine); err != nil {
		return nil, err
	}

	return o.Bytes(), nil
}

type ArrayResponse struct {
	elements []Response
}

func NewArray(responses ...Response) *ArrayResponse {
	array := &ArrayResponse{make([]Response, 0)}
	for _, r := range responses {
		array.elements = append(array.elements, r)
	}

	return array
}

func (r *ArrayResponse) Add(e Response) *ArrayResponse {
	r.elements = append(r.elements, e)
	return r
}

func (r *ArrayResponse) encode() ([]byte, error) {
	var o bytes.Buffer

	// Element count
	if err := o.WriteByte('*'); err != nil {
		return nil, err
	}
	if _, err := o.WriteString(strconv.Itoa(len(r.elements))); err != nil {
		return nil, err
	}
	if _, err := o.Write(newLine); err != nil {
		return nil, err
	}

	for _, e := range r.elements {
		eBytes, err := e.encode()
		if err != nil {
			return nil, err
		}
		if _, err := o.Write(eBytes); err != nil {
			return nil, err
		}
	}

	return o.Bytes(), nil
}
