package schema

// Option is a function that configures a Definition.
type Option func(*Definition)

// Define creates a new Definition with the given options.
func Define(opts ...Option) *Definition {
	d := &Definition{}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

// POST sets the endpoint to HTTP POST with the given path.
func POST(path string) Option {
	return func(d *Definition) {
		d.Method = "POST"
		d.Path = path
	}
}

// GET sets the endpoint to HTTP GET with the given path.
func GET(path string) Option {
	return func(d *Definition) {
		d.Method = "GET"
		d.Path = path
	}
}

// PUT sets the endpoint to HTTP PUT with the given path.
func PUT(path string) Option {
	return func(d *Definition) {
		d.Method = "PUT"
		d.Path = path
	}
}

// DELETE sets the endpoint to HTTP DELETE with the given path.
func DELETE(path string) Option {
	return func(d *Definition) {
		d.Method = "DELETE"
		d.Path = path
	}
}

// PATCH sets the endpoint to HTTP PATCH with the given path.
func PATCH(path string) Option {
	return func(d *Definition) {
		d.Method = "PATCH"
		d.Path = path
	}
}

// Request defines the request fields for this endpoint.
func Request(fields ...Field) Option {
	return func(d *Definition) {
		d.ReqFields = fields
	}
}

// Response defines the response fields for this endpoint.
func Response(fields ...Field) Option {
	return func(d *Definition) {
		d.ResFields = fields
	}
}
