package test

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/246859/codis/redis/resproto2"
	"testing"
)

func testData(t *testing.T, datas [][]byte) {
	for i, data := range datas {
		next := resproto2.ParseRespProto(bytes.NewReader(data))
		for {
			data, err := next()
			if err != nil && !errors.Is(err, resproto2.EOF) {
				t.Error(err)
			}
			fmt.Println(fmt.Sprintf("<=== DATA %d ===>", i))
			fmt.Println(string(data.Bytes()))
			if errors.Is(err, resproto2.EOF) {
				break
			}
		}
	}
}

func TestSimpleString(t *testing.T) {
	datas := [][]byte{
		[]byte("+123\r\n"),
		[]byte("+OK\r\n"),
		[]byte("+asjdlakjsdljalsd\r\n"),
		[]byte("+akapopda123\r\n"),
		[]byte("+🐕\r\n"),
	}
	testData(t, datas)
}

func TestInteger(t *testing.T) {
	datas := [][]byte{
		[]byte(":1\r\n"),
		[]byte(":1024\r\n"),
		[]byte(":9223372036854775807\r\n"),
		[]byte(":-9223372036854775808\r\n"),
	}

	testData(t, datas)
}

func TestError(t *testing.T) {
	datas := [][]byte{
		[]byte("-Error message\r\n"),
		[]byte("-ERR unknown command 'foobar'\r\n"),
		[]byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"),
	}
	testData(t, datas)
}

func TestBulkString(t *testing.T) {
	datas := [][]byte{
		[]byte(
			"$6\r\n" +
				"1234\r\n" +
				"1234\r\n" +
				"1234\r\n" +
				"\r\n",
		),
		[]byte(
			"$-1\r\n",
		),
		[]byte(
			"$0\r\n\r\n",
		),
	}

	testData(t, datas)
}

func TestArray(t *testing.T) {
	datas := [][]byte{
		[]byte(
			"*2\r\n" +
				"+1st\r\n" +
				"-2error\r\n" +
				":316\r\n" +
				"$1\r\n" +
				"4\r\n",
		),
		[]byte(
			"*0\r\n",
		),
	}
	testData(t, datas)
}
