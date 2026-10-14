package gitparse

// GitStructuredFields is the structured git dependency form in daml.yaml.
// damlc accepts this mapping on SDK 3.6 and later, not on earlier SDKs.
//
//	git:
//	  url: github.com/org/repo
//	  ref: main
//	  path: pkg/foo.dar
//
// Release assets use release and asset instead of ref and path.
type GitStructuredFields struct {
	URL     string `yaml:"url,omitempty"`
	Ref     string `yaml:"ref,omitempty"`
	Path    string `yaml:"path,omitempty"`
	Release string `yaml:"release,omitempty"`
	Asset   string `yaml:"asset,omitempty"`
}

// StructuredFieldsFromGit builds the structured YAML fields for a parsed git source.
func StructuredFieldsFromGit(git GitSource) *GitStructuredFields {
	fields := &GitStructuredFields{
		URL: gitDependencyCloneURLString(git.CloneURL),
	}
	if git.Release {
		fields.Release = git.Ref
		fields.Asset = git.DarPath
		return fields
	}
	fields.Ref = git.Ref
	fields.Path = git.DarPath
	return fields
}

// StructuredFieldsFromLine parses a git: line into structured YAML fields.
func StructuredFieldsFromLine(line string) (*GitStructuredFields, error) {
	dep, err := ParseGitDependency(line)
	if err != nil {
		return nil, err
	}
	return StructuredFieldsFromGit(dep.Git), nil
}
