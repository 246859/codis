package resproto2

import (
	"strconv"
)

type Data interface {
	Bytes() []byte
}

type StatusMsg struct {
	status string
}

func (s StatusMsg) Bytes() []byte {
	return []byte(s.status)
}

func (s StatusMsg) Status() string {
	return s.status
}

type IntegerMsg struct {
	i int64
}

func (i IntegerMsg) Bytes() []byte {
	return []byte(strconv.FormatInt(i.i, 10))
}

func (i IntegerMsg) Int64() int64 {
	return i.i
}

type ErrorMsg struct {
	err error
}

func (e ErrorMsg) Bytes() []byte {
	return []byte(e.err.Error())
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

	for _, data := range a.arr {
		b = append(b, append(data.Bytes(), []byte(CRLF)...)...)
	}
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
	return b.data
}
