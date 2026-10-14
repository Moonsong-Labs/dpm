package damlpackage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitDependency_inlineConflictingFieldsInYaml(t *testing.T) {
	raw := "git:github.com/org/repo.git?release=v1.0.0&asset=bar.dar#main?path=dist/foo.dar"
	contents := []byte(`sdk-version: 3.4.5
dependencies:
  - ` + raw + `
`)
	_, err := ReadFromContents(contents, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "release cannot be combined with ref or path")
}

func TestGitDependency_inDataDependencies(t *testing.T) {
	raw := "git:https://github.com/example-org/example-repo.git#main?path=pkg/foo.dar"
	contents := []byte(`sdk-version: 3.4.5
dependencies:
  - daml-script
data-dependencies:
  - ` + raw + `
`)
	p, err := ReadFromContents(contents, "")
	require.NoError(t, err)
	dep, ok := p.ParsedDarDependencies.DataDependencies[raw]
	require.True(t, ok)
	assert.Equal(t, "main", dep.Git.Ref)
	assert.Equal(t, "pkg/foo.dar", dep.Git.DarPath)
}

func TestGitDependency_inDependencies(t *testing.T) {
	raw := "git:https://github.com/example-org/example-repo.git#main?path=pkg/foo.dar"
	contents := []byte(`sdk-version: 3.4.5
dependencies:
  - ` + raw + `
`)
	p, err := ReadFromContents(contents, "")
	require.NoError(t, err)
	dep, ok := p.ParsedDarDependencies.Dependencies[raw]
	require.True(t, ok)
	assert.Equal(t, "main", dep.Git.Ref)
	assert.Equal(t, "pkg/foo.dar", dep.Git.DarPath)
}

func TestParseGitStructuredDependency(t *testing.T) {
	contents := []byte(`sdk-version: 3.6.0
dependencies:
  - git:
      url: https://github.com/org/repo.git
      ref: main
      path: pkg/foo.dar
data-dependencies:
  - git:
      url: github.com/org/repo.git
      release: v1.0.0
      asset: foo.dar
`)
	p, err := ReadFromContents(contents, "")
	require.NoError(t, err)

	require.Len(t, p.Dependencies, 1)
	dep := p.Dependencies[0]
	require.NotNil(t, dep.GitStructured)
	assert.Equal(t, "main", dep.GitStructured.Ref)
	assert.Equal(t, "pkg/foo.dar", dep.GitStructured.Path)
	parsed, ok := p.ParsedDarDependencies.Dependencies["git:https://github.com/org/repo.git#main?path=pkg/foo.dar"]
	require.True(t, ok)
	assert.Equal(t, "main", parsed.Git.Ref)
	assert.Equal(t, "pkg/foo.dar", parsed.Git.DarPath)
	assert.Equal(t, "https://github.com/org/repo.git", parsed.Git.CloneURL.String())

	require.Len(t, p.DataDependencies, 1)
	release := p.DataDependencies[0]
	require.NotNil(t, release.GitStructured)
	assert.Equal(t, "v1.0.0", release.GitStructured.Release)
	assert.Equal(t, "foo.dar", release.GitStructured.Asset)
}

func TestParseGitStructuredDependency_requiresSDK36(t *testing.T) {
	t.Parallel()

	body := `
dependencies:
  - git:
      url: github.com/org/repo
      ref: main
      path: pkg/foo.dar
`
	cases := []struct {
		name    string
		version string
		ok      bool
	}{
		{name: "unset", version: "", ok: false},
		{name: "3.5", version: "3.5.2", ok: false},
		{name: "3.6 snapshot", version: "3.6.0-snapshot.20250901", ok: true},
		{name: "3.6", version: "3.6.0", ok: true},
		{name: "3.7", version: "3.7.1", ok: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			header := ""
			if tc.version != "" {
				header = "sdk-version: " + tc.version + "\n"
			}
			_, err := ReadFromContents([]byte(header+body), "")
			if tc.ok {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), "SDK 3.6 and later")
			assert.Contains(t, err.Error(), "not on earlier SDKs")
			assert.Contains(t, err.Error(), "git: one-liner")
		})
	}
}

func TestParseGitStructuredDependency_conflictingFields(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		yaml string
		err  string
	}{
		{
			name: "release with ref and path",
			yaml: `sdk-version: 3.4.5
dependencies:
  - git:
      url: https://github.com/org/repo.git
      ref: main
      path: dist/foo.dar
      release: v1.0.0
      asset: bar.dar
`,
			err: "release cannot be combined with ref or path",
		},
		{
			name: "asset without release",
			yaml: `sdk-version: 3.4.5
dependencies:
  - git:
      url: https://github.com/org/repo.git
      ref: main
      path: pkg/foo.dar
      asset: bar.dar
`,
			err: "asset requires release",
		},
		{
			name: "path without ref",
			yaml: `sdk-version: 3.4.5
dependencies:
  - git:
      url: https://github.com/org/repo.git
      path: pkg/foo.dar
`,
			err: "path requires ref",
		},
		{
			name: "missing url",
			yaml: `sdk-version: 3.4.5
dependencies:
  - git:
      ref: main
      path: pkg/foo.dar
`,
			err: "url is required",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ReadFromContents([]byte(tc.yaml), "")
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.err)
		})
	}
}

func TestGitDependencyHostsAgreeAcrossLayers(t *testing.T) {
	t.Parallel()

	lines := []string{
		"git:gitlab.com/calvogenerico/daml-temp-coso#3000cf452734676e9c87d0a92d1bad38a3f16ec3?path=foo.dar",
		"git:github.com/org/repo#main?path=dist/foo.dar",
		"git:bitbucket.org/org/repo#main?path=foo.dar",
		"git:git.internal.example.com/team/repo#v1.2.3?path=out/foo.dar",
		"git:https://gitlab.example.com/group/subgroup/repo.git#main?path=foo.dar",
	}

	for _, line := range lines {
		t.Run(line, func(t *testing.T) {
			contents := []byte("sdk-version: 3.4.5\ndependencies:\n  - " + line + "\n")
			p, err := ReadFromContents(contents, "")
			require.NoError(t, err, "package parse should accept any https host")
			require.Len(t, p.Dependencies, 1)

			_, _, coerceErr := CanonicalizeRawGitDependencies([]*RawDependency{rawDependencyFromValue(line)})
			require.NoError(t, coerceErr, "canonicalization must accept whatever parsing accepts")
		})
	}
}

func TestParseGitDependency_gitLabReleaseIsRejectedWithGuidance(t *testing.T) {
	t.Parallel()

	_, err := ExpandReleaseGitDependenciesRaw(t.Context(), []*RawDependency{
		rawDependencyFromValue("git:gitlab.com/org/repo?release=v1.0.0"),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "only supported for github.com")
	assert.Contains(t, err.Error(), "?path=")
}
