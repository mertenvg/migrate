package files

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestNewProvider(t *testing.T) {
	p := NewProvider("./testdata")
	fmt.Println(p.names)
	wanted := []string{
		"00001",
		"00002",
		"00003",
		"00004",
		"00005.some-description",
	}
	hasDown := map[string]bool{
		"00001":                  true,
		"00004":                  true,
		"00005.some-description": true,
	}
	for _, want := range wanted {
		m, err := p.Next()
		if err != nil {
			t.Errorf("Next() unexpected error %v", err)
		}
		if m == nil {
			t.Errorf("Next() expected non nil value")
			continue
		}
		if m.Name() != want {
			t.Errorf("Next() Name() wanted %s, got %s", want, m.Name())
		}
		upData, err := io.ReadAll(m.Up())
		if err != nil {
			t.Errorf("ReadAll(Up()) unexpected error %v", err)
		}
		upData = bytes.TrimSpace(upData)
		if string(upData) != fmt.Sprintf("%s.up", want) {
			t.Errorf("Up() wanted %s, got %s", fmt.Sprintf("%s.up", want), string(upData))
		}
		downData, err := io.ReadAll(m.Down())
		if err != nil {
			t.Errorf("ReadAll(Down()) unexpected error %v", err)
		}
		downData = bytes.TrimSpace(downData)
		if has := hasDown[m.Name()]; has {
			if string(downData) != fmt.Sprintf("%s.down", want) {
				t.Errorf("Up() wanted %s, got %s", fmt.Sprintf("%s.down", want), string(downData))
			}
		}
		m.Close()
	}
}
