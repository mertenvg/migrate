package reader

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
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

// copyFromStdinRe matches a `COPY ... FROM stdin` statement so the inline data
// block following it can be skipped.
var copyFromStdinRe = regexp.MustCompile(`(?is)^\s*COPY\b.*\bFROM\s+STDIN\b`)

func (r *SQLReader) Next() (query string, err error) {
	if FailNext != nil {
		err := FailNext
		FailNext = nil
		return "", err
	}
	buf := bytes.Buffer{}
	var quote bool
	var quoteChar rune
	var eString bool  // current `'...'` literal is an E-string (allows `\` escapes)
	var bsEscape bool // previous char was a backslash inside an E-string
	// When inDollar we are inside a Postgres dollar-quoted block. dollarTag is
	// the tag between the dollar signs (empty string for `$$`).
	var inDollar bool
	var dollarTag string
	for {
		if !quote && !inDollar {
			// Dollar-quote delimiters.
			if tag, n, ok := r.peekDollarTag(); ok {
				inDollar = true
				dollarTag = tag
				for range n {
					b, rerr := r.source.ReadByte()
					if rerr != nil {
						return "", fmt.Errorf("failed to read source: %w", rerr)
					}
					buf.WriteByte(b)
				}
				continue
			}
			// `--` line comment.
			if b, _ := r.source.Peek(2); len(b) == 2 && b[0] == '-' && b[1] == '-' {
				for {
					b, rerr := r.source.ReadByte()
					if rerr != nil {
						if rerr == io.EOF {
							break
						}
						return "", fmt.Errorf("failed to read source: %w", rerr)
					}
					if b == '\n' {
						break
					}
				}
				continue
			}
			// `/* ... */` block comment (nestable).
			if b, _ := r.source.Peek(2); len(b) == 2 && b[0] == '/' && b[1] == '*' {
				if _, rerr := r.source.Discard(2); rerr != nil {
					return "", fmt.Errorf("failed to read source: %w", rerr)
				}
				depth := 1
				for depth > 0 {
					b, _ := r.source.Peek(2)
					if len(b) == 2 && b[0] == '/' && b[1] == '*' {
						r.source.Discard(2)
						depth++
						continue
					}
					if len(b) == 2 && b[0] == '*' && b[1] == '/' {
						r.source.Discard(2)
						depth--
						continue
					}
					if len(b) < 2 {
						return "", fmt.Errorf("unterminated block comment")
					}
					if _, rerr := r.source.ReadByte(); rerr != nil {
						return "", fmt.Errorf("failed to read source: %w", rerr)
					}
				}
				continue
			}
		}

		if inDollar {
			// Inside a dollar block: look for the matching closing tag first.
			if tag, n, ok := r.peekDollarTag(); ok && tag == dollarTag {
				inDollar = false
				dollarTag = ""
				for range n {
					b, rerr := r.source.ReadByte()
					if rerr != nil {
						return "", fmt.Errorf("failed to read source: %w", rerr)
					}
					buf.WriteByte(b)
				}
				continue
			}
		}

		v, _, err := r.source.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("failed to read source: %w", err)
		}

		if inDollar {
			buf.WriteRune(v)
			continue
		}

		if quote {
			if bsEscape {
				buf.WriteRune(v)
				bsEscape = false
				continue
			}
			if eString && v == '\\' {
				buf.WriteRune(v)
				bsEscape = true
				continue
			}
			if v == quoteChar {
				// Doubled quote (`''` / `""`) is an escape — stay inside the literal.
				if next, _ := r.source.Peek(1); len(next) == 1 && rune(next[0]) == quoteChar {
					buf.WriteRune(v)
					r.source.ReadByte()
					buf.WriteRune(quoteChar)
					continue
				}
				buf.WriteRune(v)
				quote = false
				eString = false
				quoteChar = 0
				continue
			}
			buf.WriteRune(v)
			continue
		}

		// Not in any quoted context.
		if v == '\'' || v == '"' {
			quote = true
			quoteChar = v
			if v == '\'' && bufEndsWithEPrefix(&buf) {
				eString = true
			}
			buf.WriteRune(v)
			continue
		}
		if v == ';' {
			break
		}
		buf.WriteRune(v)
	}

	stmt := strings.TrimSpace(buf.String())

	// `COPY ... FROM stdin;` is followed by an inline data block terminated by
	// a line containing only `\.`. That data is not SQL and would otherwise be
	// parsed as further statements — skip it.
	if copyFromStdinRe.MatchString(stmt) {
		if err := r.skipCopyData(); err != nil {
			return "", err
		}
	}

	return stmt, nil
}

// skipCopyData consumes a Postgres COPY data block up to and including the
// terminating `\.` line.
func (r *SQLReader) skipCopyData() error {
	for {
		line, err := r.source.ReadString('\n')
		trimmed := strings.TrimRight(line, "\r\n")
		if trimmed == `\.` {
			return nil
		}
		if err == io.EOF {
			return fmt.Errorf("unterminated COPY data block")
		}
		if err != nil {
			return fmt.Errorf("failed to read copy data: %w", err)
		}
	}
}

// bufEndsWithEPrefix reports whether the buffer ends with a standalone `E`/`e`
// token, which marks the following `'...'` as a Postgres escape string.
func bufEndsWithEPrefix(buf *bytes.Buffer) bool {
	b := buf.Bytes()
	if len(b) == 0 {
		return false
	}
	last := b[len(b)-1]
	if last != 'e' && last != 'E' {
		return false
	}
	if len(b) == 1 {
		return true
	}
	return !isIdentCont(b[len(b)-2])
}

func isIdentCont(b byte) bool {
	return b == '_' || (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
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
