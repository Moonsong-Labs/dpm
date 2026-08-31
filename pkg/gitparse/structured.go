package gitparse

// GitStructuredFields is the structured git dependency form in daml.yaml.
type GitStructuredFields struct {
	URL     string `yaml:"url"`
	Ref     string `yaml:"ref"`
	Path    string `yaml:"path"`
	Release string `yaml:"release"`
	Asset   string `yaml:"asset"`
}
