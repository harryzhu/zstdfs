package potato

type options struct {
	i int
	s string
	b bool
	m map[string]string
}

func NewOption(opts ...ConfigOption) *options {
	r := new(options)
	for _, opt := range opts {
		opt(r)
	}
	return r
}

type ConfigOption func(*options)

func WriteInt(s int) ConfigOption {
	return func(o *options) {
		o.i = s
	}
}

func WriteString(s string) ConfigOption {
	return func(o *options) {
		o.s = s
	}
}

func WriteBool(s bool) ConfigOption {
	return func(o *options) {
		o.b = s
	}
}

func WriteMap(s map[string]string) ConfigOption {
	return func(o *options) {
		o.m = s
	}
}
