package resproto2

import (
	"fmt"
	"strconv"
)

type Data interface {
	Bytes() []byte
}

type StatusMsg struct {
	status string
}

func (s StatusMsg) Bytes() []byte {
	return []byte(string(simpleStringMsg) + s.status + CRLF)
}

func (s StatusMsg) Status() string {
	return s.status
}

type IntegerMsg struct {
	i int64
}

func (i IntegerMsg) Bytes() []byte {
	return []byte(string(integerMsg) + strconv.FormatInt(i.i, 10) + CRLF)
}

func (i IntegerMsg) Int64() int64 {
	return i.i
}

type ErrorMsg struct {
	err error
}

func (e ErrorMsg) Bytes() []byte {
	return []byte(string(errorMsg) + e.err.Error() + CRLF)
}

func (e ErrorMsg) Error() error {
	return e.err
}

type ArrayMsg struct {
	arr []Data
	len int64
}

func (a ArrayMsg) Len() int64 {
	return a.len
}

func (a ArrayMsg) Array() []Data {
	return a.arr
}

func (a ArrayMsg) Bytes() []byte {
	var b []byte
	b = append(b, fmt.Sprintf("%c%d\r\n", arrayMsg, a.len)...)
	for _, data := range a.arr {
		b = append(b, data.Bytes()...)
	}
	b = append(b, CRLF...)
	return b
}

type BulkStringMsg struct {
	data []byte
	len  int64
}

func (b BulkStringMsg) Len() int64 {
	return b.len
}

func (b BulkStringMsg) Bytes() []byte {
	var bs []byte
	bs = append(bs, fmt.Sprintf("%c%d\r\n", bulkStringMsg, b.len)...)
	bs = append(bs, b.data...)
	bs = append(bs, CRLF...)
	return b.data
}
