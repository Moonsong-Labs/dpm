package damlpackage

import (
	"fmt"
	"strings"

	"daml.com/x/assistant/pkg/gitparse"
	"github.com/goccy/go-yaml"
)

var RawDependenciesSchemaErr = fmt.Errorf("dar dependencies fields must be of type string, structured git object, or '{ value: <string>, main-package-id: <string> }' object")

// RawDependency is the 'string | {...}' sum-type for
// dependencies / data-dependencies YAML fields.
// A structured git object stays a mapping in daml.yaml; pinning updates its ref.
// That mapping is only valid when sdk-version is 3.6 or later.
type RawDependency struct {
	ValueOnly     *string
	WithPackageId *withPackageId
	GitStructured *gitparse.GitStructuredFields
}

type withPackageId struct {
	Value         string `yaml:"value"`
	MainPackageId string `yaml:"main-package-id"`
}

type gitStructuredEntry struct {
	Git *gitparse.GitStructuredFields `yaml:"git"`
}

func (r *RawDependency) Value() (string, error) {
	switch {
	case r.GitStructured != nil:
		return gitparse.FormatGitStructuredLine(r.GitStructured)
	case r.WithPackageId != nil:
		return r.WithPackageId.Value, nil
	case r.ValueOnly != nil:
		return *r.ValueOnly, nil
	default:
		return "", RawDependenciesSchemaErr
	}
}

func (r *RawDependency) GetMainPackageId() *string {
	if r.WithPackageId == nil {
		return nil
	}
	return &r.WithPackageId.MainPackageId
}

func (r *RawDependency) UnmarshalYAML(b []byte) error {
	var s string
	if err := yaml.Unmarshal(b, &s); err == nil {
		r.ValueOnly = &s
		return nil
	}

	var asMap map[string]any
	if err := yaml.Unmarshal(b, &asMap); err == nil {
		if _, ok := asMap["git"]; ok {
			var gitEntry gitStructuredEntry
			if err := yaml.Unmarshal(b, &gitEntry); err != nil {
				return fmt.Errorf("git dependency: %w", err)
			}
			if gitEntry.Git == nil {
				gitEntry.Git = &gitparse.GitStructuredFields{}
			}
			r.GitStructured = gitEntry.Git
			return nil
		}
	}

	var obj withPackageId
	if err := yaml.Unmarshal(b, &obj); err == nil {
		if strings.TrimSpace(obj.Value) == "" {
			return RawDependenciesSchemaErr
		}
		r.WithPackageId = &obj
		return nil
	}

	return RawDependenciesSchemaErr
}

func (r *RawDependency) MarshalYAML() (any, error) {
	switch {
	case r.GitStructured != nil:
		return gitStructuredEntry{Git: r.GitStructured}, nil
	case r.WithPackageId != nil:
		return *r.WithPackageId, nil
	case r.ValueOnly != nil:
		return r.ValueOnly, nil
	default:
		return nil, nil
	}
}
