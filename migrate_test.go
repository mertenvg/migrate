package migrate

import (
	"testing"
)

func TestWithFileSource(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("unexpected panic: %v", r)
		}
	}()

	o := WithFileSource("./testdata/file_source")
	m := &Migrate{}
	o(m)
}

func TestNew_Postgres(t *testing.T) {
	t.Skip()
}
