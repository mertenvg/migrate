package reader

import (
	"bufio"
	"bytes"
	"errors"
	"reflect"
	"testing"
)

func TestNewSQLReader(t *testing.T) {
	src := bytes.NewBufferString("")
	want := &SQLReader{
		source: bufio.NewReader(src),
	}
	if got := NewSQLReader(src); !reflect.DeepEqual(got, want) {
		t.Errorf("NewSQLReader() = %v, want %v", got, want)
	}
}

func validateNext(t *testing.T, r *SQLReader, want string, wantErr bool) {
	q, err := r.Next()
	if (err != nil) != wantErr {
		t.Errorf("Next() unexpected error = %v, want nil", err)
	}
	if q != want {
		t.Errorf("Next() got '%v', want '%v'", q, want)
	}
}

func TestSQLReader_Next(t *testing.T) {
	src := bytes.NewBufferString(`Query 1; Query 2; Insert 'it''s; ok', "weird""; name" into "table;"; done `)
	r := NewSQLReader(src)

	validateNext(t, r, "Query 1", false)
	validateNext(t, r, "Query 2", false)
	validateNext(t, r, `Insert 'it''s; ok', "weird""; name" into "table;"`, false)
	validateNext(t, r, "done", false)
	validateNext(t, r, "", false)
}

func TestSQLReader_Next_EString(t *testing.T) {
	// Backslash escapes only apply inside E'...'. The `\'` should not close the
	// literal, so the `;` inside must not split.
	src := bytes.NewBufferString(`SELECT E'a\'; b\\c'; SELECT 2;`)
	r := NewSQLReader(src)

	validateNext(t, r, `SELECT E'a\'; b\\c'`, false)
	validateNext(t, r, "SELECT 2", false)
}

func TestSQLReader_Next_PlainStringBackslashLiteral(t *testing.T) {
	// In a regular '...' literal, `\` is just a character — only `''` escapes.
	// So `\'` ends the string, and the following `;` splits.
	src := bytes.NewBufferString(`SELECT 'a\'; SELECT 2;`)
	r := NewSQLReader(src)

	validateNext(t, r, `SELECT 'a\'`, false)
	validateNext(t, r, "SELECT 2", false)
}

func TestSQLReader_Next_EStringPrefixMustBeStandalone(t *testing.T) {
	// `somE'x'` is identifier `somE` followed by string `'x'`, NOT an E-string.
	// So the `\` inside is literal and the next `'` closes the string.
	src := bytes.NewBufferString(`SELECT somE'a\'; SELECT 2;`)
	r := NewSQLReader(src)

	validateNext(t, r, `SELECT somE'a\'`, false)
	validateNext(t, r, "SELECT 2", false)
}

func TestSQLReader_Next_UnicodeEscapeString(t *testing.T) {
	// U&'...' uses `''` as the quote escape (default `\` escapes hex codepoints
	// and never produces a `'`), so splitting works the same as a normal string.
	src := bytes.NewBufferString(`SELECT U&'d\0061t\+000061; ok'; SELECT 2;`)
	r := NewSQLReader(src)

	validateNext(t, r, `SELECT U&'d\0061t\+000061; ok'`, false)
	validateNext(t, r, "SELECT 2", false)
}

func TestSQLReader_Next_CopyFromStdin(t *testing.T) {
	src := bytes.NewBufferString("COPY t (a, b) FROM stdin;\n1\tfoo;bar\n2\tbaz\n\\.\nSELECT 1;")
	r := NewSQLReader(src)

	validateNext(t, r, "COPY t (a, b) FROM stdin", false)
	validateNext(t, r, "SELECT 1", false)
	validateNext(t, r, "", false)
}

func TestSQLReader_Next_CopyFromStdinUnterminated(t *testing.T) {
	src := bytes.NewBufferString("COPY t FROM stdin;\n1\tfoo\n2\tbar\n")
	r := NewSQLReader(src)

	if _, err := r.Next(); err == nil {
		t.Errorf("Next() expected error for unterminated COPY data block")
	}
}

func TestSQLReader_Next_DollarQuoted(t *testing.T) {
	src := bytes.NewBufferString(`DO $$ BEGIN RAISE NOTICE 'hi;there'; END $$; SELECT 1;`)
	r := NewSQLReader(src)

	validateNext(t, r, `DO $$ BEGIN RAISE NOTICE 'hi;there'; END $$`, false)
	validateNext(t, r, "SELECT 1", false)
	validateNext(t, r, "", false)
}

func TestSQLReader_Next_DollarQuotedTagged(t *testing.T) {
	src := bytes.NewBufferString(`CREATE FUNCTION f() RETURNS void AS $body$ BEGIN PERFORM 'x;$$y'; END $body$ LANGUAGE plpgsql; SELECT 2;`)
	r := NewSQLReader(src)

	validateNext(t, r, `CREATE FUNCTION f() RETURNS void AS $body$ BEGIN PERFORM 'x;$$y'; END $body$ LANGUAGE plpgsql`, false)
	validateNext(t, r, "SELECT 2", false)
}

func TestSQLReader_Next_LineComment(t *testing.T) {
	src := bytes.NewBufferString("INSERT INTO plates VALUES ('A1') -- dev/test data; license plates\n, ('B2');\nSELECT 1;")
	r := NewSQLReader(src)

	validateNext(t, r, "INSERT INTO plates VALUES ('A1') , ('B2')", false)
	validateNext(t, r, "SELECT 1", false)
	validateNext(t, r, "", false)
}

func TestSQLReader_Next_LineCommentInString(t *testing.T) {
	src := bytes.NewBufferString(`SELECT '-- not a comment; really'; SELECT 2;`)
	r := NewSQLReader(src)

	validateNext(t, r, `SELECT '-- not a comment; really'`, false)
	validateNext(t, r, "SELECT 2", false)
}

func TestSQLReader_Next_LineCommentEOF(t *testing.T) {
	src := bytes.NewBufferString("SELECT 1; -- trailing comment no newline")
	r := NewSQLReader(src)

	validateNext(t, r, "SELECT 1", false)
	validateNext(t, r, "", false)
}

func TestSQLReader_Next_BlockComment(t *testing.T) {
	src := bytes.NewBufferString("SELECT 1 /* a; b; c */, 2; /* trailing */ SELECT 3;")
	r := NewSQLReader(src)

	validateNext(t, r, "SELECT 1 , 2", false)
	validateNext(t, r, "SELECT 3", false)
	validateNext(t, r, "", false)
}

func TestSQLReader_Next_BlockCommentNested(t *testing.T) {
	src := bytes.NewBufferString("SELECT 1 /* outer /* inner; */ still; outer */, 2; SELECT 3;")
	r := NewSQLReader(src)

	validateNext(t, r, "SELECT 1 , 2", false)
	validateNext(t, r, "SELECT 3", false)
}

func TestSQLReader_Next_BlockCommentInString(t *testing.T) {
	src := bytes.NewBufferString(`SELECT '/* not a comment; */'; SELECT 2;`)
	r := NewSQLReader(src)

	validateNext(t, r, `SELECT '/* not a comment; */'`, false)
	validateNext(t, r, "SELECT 2", false)
}

func TestSQLReader_Next_BlockCommentUnterminated(t *testing.T) {
	src := bytes.NewBufferString("SELECT 1 /* never closes")
	r := NewSQLReader(src)

	if _, err := r.Next(); err == nil {
		t.Errorf("Next() expected error for unterminated block comment")
	}
}

func TestSQLReader_Next_DoubledSingleQuote(t *testing.T) {
	src := bytes.NewBufferString(`INSERT INTO t VALUES ('it''s; fine'); SELECT 2;`)
	r := NewSQLReader(src)

	validateNext(t, r, `INSERT INTO t VALUES ('it''s; fine')`, false)
	validateNext(t, r, "SELECT 2", false)
}

func TestSQLReader_Next_DoubledDoubleQuote(t *testing.T) {
	src := bytes.NewBufferString(`SELECT "weird""; name" FROM t; SELECT 2;`)
	r := NewSQLReader(src)

	validateNext(t, r, `SELECT "weird""; name" FROM t`, false)
	validateNext(t, r, "SELECT 2", false)
}

func TestSQLReader_Next_MixedQuoteChars(t *testing.T) {
	// A `"` inside `'...'` and a `'` inside `"..."` must not toggle the quote
	// state — previously both quote chars shared a single flag.
	src := bytes.NewBufferString(`SELECT 'a"b;c', "d'e;f" FROM t; SELECT 2;`)
	r := NewSQLReader(src)

	validateNext(t, r, `SELECT 'a"b;c', "d'e;f" FROM t`, false)
	validateNext(t, r, "SELECT 2", false)
}

func TestSQLReader_Next_WithFail(t *testing.T) {
	src := bytes.NewBufferString(`Query 1; Query 2`)
	r := NewSQLReader(src)

	FailNext = errors.New("fail")

	validateNext(t, r, "", true)
	validateNext(t, r, "Query 1", false)
	validateNext(t, r, "Query 2", false)
}

type FailReader struct{}

func (r FailReader) Read(_ []byte) (n int, err error) {
	return 0, errors.New("fail")
}

func TestSQLReader_Next_WithReadFail(t *testing.T) {
	r := NewSQLReader(&FailReader{})

	validateNext(t, r, "", true)
	validateNext(t, r, "", true)
}
