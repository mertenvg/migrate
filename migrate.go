package migrate

import (
	"fmt"
	"os"
)

type Migrate struct {
}

type Option func(m *Migrate)

func New(opts ...Option) *Migrate {
	m := &Migrate{}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func WithFileSource(path string) Option {
	files, err := os.ReadDir(path)
	if err != nil {
		panic(err)
	}
	return func(m *Migrate) {
		for _, file := range files {
			fmt.Println(file.Name(), file.IsDir())
		}
	}
}
