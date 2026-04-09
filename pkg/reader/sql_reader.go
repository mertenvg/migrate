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
	// When non-nil we are inside a Postgres dollar-quoted block. The value is
	// the tag between the dollar signs (empty string for `$$`).
	var inDollar bool
	var dollarTag string
	for {
		// Detect dollar-quote delimiters outside of regular string quotes.
		if !escape && !quote {
			if tag, n, ok := r.peekDollarTag(); ok {
				if !inDollar {
					inDollar = true
					dollarTag = tag
				} else if tag == dollarTag {
					inDollar = false
					dollarTag = ""
				}
				// Consume and write the delimiter (`$tag$`) verbatim.
				for i := 0; i < n; i++ {
					b, rerr := r.source.ReadByte()
					if rerr != nil {
						return "", fmt.Errorf("failed to read source: %w", rerr)
					}
					if werr := buf.WriteByte(b); werr != nil {
						return "", fmt.Errorf("failed to write byte to buffer: %w", werr)
					}
				}
				continue
			}
		}

		v, _, err := r.source.ReadRune()
		if err != nil {
			if err == io.EOF {
				break // End of file
			}
			return "", fmt.Errorf("failed to read source: %w", err)
		}
		if inDollar {
			if _, werr := buf.WriteRune(v); werr != nil {
				return "", fmt.Errorf("failed to write rune to buffer: %w", werr)
			}
			continue
		}
		if !escape && v == '\\' {
			escape = true
			continue
		}
		if !escape && (v == '\'' || v == '"') {
			quote = !quote
		}
		if !quote && v == ';' {
			break
		}
		if escape {
			_, err = buf.WriteRune('\\')
			if err != nil {
				return "", fmt.Errorf("failed to write rune to buffer: %w", err)
			}
			escape = false
		}
		_, err = buf.WriteRune(v)
		if err != nil {
			return "", fmt.Errorf("failed to write rune to buffer: %w", err)
		}
	}
	return strings.TrimSpace(buf.String()), nil
}

// peekDollarTag checks whether the next bytes form a Postgres dollar-quote
// delimiter (`$tag$` or `$$`). If so, it returns the tag (without the dollars)
// and the total byte length of the delimiter.
func (r *SQLReader) peekDollarTag() (string, int, bool) {
	b, _ := r.source.Peek(64)
	if len(b) < 2 || b[0] != '$' {
		return "", 0, false
	}
	for i := 1; i < len(b); i++ {
		if b[i] == '$' {
			tag := string(b[1:i])
			if !validDollarTag(tag) {
				return "", 0, false
			}
			return tag, i + 1, true
		}
		if !isTagCont(b[i]) {
			return "", 0, false
		}
		if i == 1 && !isTagStart(b[i]) {
			return "", 0, false
		}
	}
	return "", 0, false
}

func isTagStart(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isTagCont(c byte) bool {
	return isTagStart(c) || (c >= '0' && c <= '9')
}

func validDollarTag(s string) bool {
	if s == "" {
		return true
	}
	if !isTagStart(s[0]) {
		return false
	}
	for i := 1; i < len(s); i++ {
		if !isTagCont(s[i]) {
			return false
		}
	}
	return true
}
