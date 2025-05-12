package files

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mertenvg/migrate"
)

type Provider struct {
	position   int
	names      []string
	migrations map[string]*Migration
}

func NewProvider(path string) *Provider {
	files, err := os.ReadDir(path)
	if err != nil {
		panic(fmt.Errorf("cannot read dir '%v': %w", path, err))
	}
	var names []string
	var downFiles []os.DirEntry
	migrations := make(map[string]*Migration)

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fileName := file.Name()
		if strings.HasSuffix(fileName, ".down.sql") {
			downFiles = append(downFiles, file)
			continue
		}
		name := fileName
		if strings.HasSuffix(name, ".sql") {
			name = strings.TrimSuffix(name, ".sql")
		}
		if strings.HasSuffix(name, ".up") {
			name = strings.TrimSuffix(name, ".up")
		}
		names = append(names, name)
		migrations[name] = &Migration{
			name:   name,
			upPath: filepath.Join(path, fileName),
		}
	}
	for _, file := range downFiles {
		fileName := file.Name()
		name := strings.TrimSuffix(fileName, ".down.sql")
		if migration, ok := migrations[name]; ok {
			migration.downPath = filepath.Join(path, fileName)
		} else {
			panic(fmt.Errorf("no matching 'up' migration found for '%v'", fileName))
		}
	}
	return &Provider{}
}

func (p *Provider) Next() (migrate.Migration, error) {
	if p.position >= len(p.names) {
		return nil, nil
	}
	name := p.names[p.position]
	migration, ok := p.migrations[name]
	p.position++
	if !ok {
		return nil, fmt.Errorf("no migration found for '%v'", name)
	}
	return migration, nil
}

type Migration struct {
	name     string
	upPath   string
	downPath string
	close    []io.Closer
}

func (m *Migration) Name() string {
	return m.name
}

func (m *Migration) Up() io.Reader {
	file, err := os.Open(m.upPath)
	if err != nil {
		panic(fmt.Errorf("cannot open migration file '%v': %w", m.upPath, err))
	}
	m.close = append(m.close, file)
	return file
}

func (m *Migration) Down() io.Reader {
	file, err := os.Open(m.downPath)
	if err != nil {
		panic(fmt.Errorf("cannot open migration file '%v': %w", m.downPath, err))
	}
	m.close = append(m.close, file)
	return file
}

func (m *Migration) Close() {
	for _, file := range m.close {
		err := file.Close()
		if err != nil {
			panic(fmt.Errorf("cannot close migration file '%v': %w", m.name, err))
		}
	}
}
