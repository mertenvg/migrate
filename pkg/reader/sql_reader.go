package reader

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

// FailNext will return the specified error  when calling Next() on you SQLReader.
// This is intended for testing purposes unless you specifically want to break your application
var FailNext error

type SQLReader struct {
	source *bufio.Reader
}

func NewSQLReader(source io.Reader) *SQLReader {
	return &SQLReader{
		source: bufio.NewReader(source),
	}
}

func (r *SQLReader) Next() (query string, err error) {
	if FailNext != nil {
		err := FailNext
		FailNext = nil
		return "", err
	}
	buf := bytes.Buffer{}
	var quote bool
	var escape bool
	for {
		v, _, err := r.source.ReadRune()
		if err != nil {
			if err == io.EOF {
				break // End of file
			}
			return "", fmt.Errorf("failed to read source: %w", err)
		}
		if v == '\\' {
			escape = !escape
		}
		if !escape && (v == '\'' || v == '"') {
			quote = !quote
		}
		if !quote && v == ';' {
			break
		}
		_, err = buf.WriteRune(v)
		if err != nil {
			return "", fmt.Errorf("failed to write rune to buffer: %w", err)
		}
	}
	return strings.TrimSpace(buf.String()), nil
}
