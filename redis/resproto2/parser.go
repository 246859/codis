package resproto2

import (
	"bufio"
	"errors"
	errors2 "github.com/pkg/errors"
	"io"
	"net/textproto"
	"strconv"
)

const (
	CRLF = "\r\n"

	CR = '\r'
	LF = '\n'

	simpleStringMsg = '+'
	errorMsg        = '-'
	integerMsg      = ':'
	bulkStringMsg   = '$'
	arrayMsg        = '*'
)

var (
	ErrInvalidReader      = errors.New("invalid reader")
	ErrInvalidHeader      = errors.New("invalid header")
	ErrMismatchHeaderType = errors.New("mismatch header type")
	// EOF represent an individual message parse completed
	EOF = errors.New("RESP EOF")
)

type RespIterator func() (Data, error)

type RespProtocolParser func(firstLine []byte, reader textproto.Reader) (Data, error)

// ParseRespProto RESP Protocol parser
func ParseRespProto(reader io.Reader) RespIterator {
	respReader := textproto.NewReader(bufio.NewReaderSize(reader, 10240))
	return parse(respReader)
}

func parse(reader *textproto.Reader) RespIterator {

	return func() (Data, error) {
		// it will clear the CRLF default
		header, err := reader.ReadLineBytes()
		if err != nil {
			return nil, err
		}

		var (
			data     Data
			parseErr error
		)

		switch header[0] {
		case simpleStringMsg:
			data, parseErr = parseSimpleString(header, reader)
		case bulkStringMsg:
			data, parseErr = parseBulkString(header, reader)
		case integerMsg:
			data, parseErr = parseInteger(header, reader)
		case arrayMsg:
			data, parseErr = parseArray(header, reader)
		case errorMsg:
			data, parseErr = parseError(header, reader)
		}

		return data, parseErr
	}
}

// simpleMsg is start with prefix '+', like +OK\r\n
func parseSimpleString(header []byte, reader *textproto.Reader) (StatusMsg, error) {

	if header == nil {
		return StatusMsg{}, ErrInvalidHeader
	}

	if header[0] != simpleStringMsg {
		return StatusMsg{}, errors2.Wrap(ErrMismatchHeaderType, string(header[0]))
	}

	return StatusMsg{status: string(header[1:])}, EOF
}

func parseError(header []byte, reader *textproto.Reader) (ErrorMsg, error) {

	if header == nil {
		return ErrorMsg{}, ErrInvalidHeader
	}

	if header[0] != errorMsg {
		return ErrorMsg{}, errors2.Wrap(ErrMismatchHeaderType, string(header[0]))
	}

	return ErrorMsg{err: errors.New(string(header[1:]))}, EOF
}

func parseInteger(header []byte, reader *textproto.Reader) (IntegerMsg, error) {

	if header == nil {
		return IntegerMsg{}, ErrInvalidHeader
	}

	if header[0] != integerMsg {
		return IntegerMsg{}, errors2.Wrap(ErrMismatchHeaderType, string(header[0]))
	}

	integerStr := string(header[1:])

	i, err := strconv.ParseInt(integerStr, 10, 64)
	if err != nil {
		return IntegerMsg{}, err
	}

	return IntegerMsg{i: i}, EOF
}

func parseBulkString(header []byte, reader *textproto.Reader) (BulkStringMsg, error) {

	if header == nil {
		return BulkStringMsg{}, ErrInvalidHeader
	}

	if header[0] != bulkStringMsg {
		return BulkStringMsg{}, errors2.Wrap(ErrMismatchHeaderType, string(header[0]))
	}

	// parse bulk string content length
	contentLengthData := string(header[1:])
	contentLen, err := strconv.ParseInt(contentLengthData, 10, 64)
	if err != nil {
		return BulkStringMsg{}, err
	}

	if contentLen < 0 {
		return BulkStringMsg{len: contentLen}, EOF
	} else if contentLen == 0 {
		// read rest CRLF bytes
		var crlf [2]byte
		_, err := reader.R.Read(crlf[:])
		if err != nil && !errors.Is(err, io.EOF) {
			return BulkStringMsg{}, err
		}
		return BulkStringMsg{data: make([]byte, 0)}, EOF
	}

	// read data with specific content length
	content := make([]byte, contentLen+int64(len(CRLF)))

	if _, err := io.ReadFull(reader.R, content); err != nil {
		return BulkStringMsg{}, err
	}

	data := content[:contentLen]

	return BulkStringMsg{
		data: data,
		len:  contentLen,
	}, EOF
}

func parseArray(header []byte, reader *textproto.Reader) (ArrayMsg, error) {
	if header == nil {
		return ArrayMsg{}, ErrInvalidHeader
	} else if header[0] != arrayMsg {
		return ArrayMsg{}, errors2.Wrap(ErrMismatchHeaderType, string(header[0]))
	}

	// parse arr content length
	contentLengthData := string(header[1:])
	contentLen, err := strconv.ParseInt(contentLengthData, 10, 64)
	if err != nil {
		return ArrayMsg{}, err
	}

	var dataList []Data

	if contentLen < 0 {
		return ArrayMsg{
			arr: dataList,
			len: contentLen,
		}, EOF
	}

	next := parse(reader)

	for i := int64(0); i < contentLen; i++ {
		data, err := next()
		if err != nil && !errors.Is(err, EOF) {
			return ArrayMsg{}, err
		}

		dataList = append(dataList, data)
	}

	return ArrayMsg{
		arr: dataList,
		len: contentLen,
	}, EOF
}
