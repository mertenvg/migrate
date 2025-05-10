package migrate

func WithProvider(p Provider) Option {
	return func(m *Migrate) {
		m.p = p
	}
}

func WithAdapter(a Adapter) Option {
	return func(m *Migrate) {
		m.a = a
	}
}
