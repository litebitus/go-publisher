package described

// MockSettings are a mock section, described over
// two lines.
type MockSettings struct {
	// The mock directory, as `dir`.
	Directory   string `yaml:"dir"`
	Undescribed string `yaml:"undescribed"`
	unexported  string
}

type mockHidden struct {
	// Never read: the type is not exported.
	Field string
}
