package damlpackage

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"daml.com/x/assistant/pkg/githubrelease"
	"daml.com/x/assistant/pkg/yamledit"
	"github.com/goccy/go-yaml"
)

// DescribeGitFetch returns a short human-readable summary for progress output.
func DescribeGitFetch(dep *ParsedDarDependency) string {
	if dep == nil || dep.Git.CloneURL == nil {
		return "git dependency"
	}
	if dep.Git.Release {
		return fmt.Sprintf("%s release %q asset %q", dep.Git.CloneURL.String(), dep.Git.Ref, dep.Git.DarPath)
	}
	return fmt.Sprintf("%s @ %q path %q", dep.Git.CloneURL.String(), dep.Git.Ref, dep.Git.DarPath)
}

// FormatGitReleaseLine builds git:...?release=TAG&asset=NAME (asset may be empty).
func FormatGitReleaseLine(cloneURL, releaseTag, asset string) string {
	u, err := url.Parse(cloneURL)
	if err != nil {
		return "git:" + cloneURL
	}
	q := u.Query()
	q.Set("release", releaseTag)
	if strings.TrimSpace(asset) != "" {
		q.Set("asset", asset)
	} else {
		q.Del("asset")
	}
	u.RawQuery = q.Encode()
	return "git:" + u.String()
}

// FormatGitStructuredLine builds a git: URI from structured YAML fields.
func FormatGitStructuredLine(fields *GitStructuredFields) (string, error) {
	if err := validateGitStructuredFields(fields); err != nil {
		return "", err
	}
	if fields.URL == "" {
		return "", fmt.Errorf("git dependency: url is required")
	}
	if fields.Release != "" {
		return FormatGitReleaseLine(fields.URL, fields.Release, fields.Asset), nil
	}
	if fields.Ref == "" {
		return "", fmt.Errorf("git dependency: ref or release is required")
	}
	if fields.Path == "" {
		return "", fmt.Errorf("git dependency: path is required for repo-file dependencies")
	}
	return fmt.Sprintf("git:%s#%s?path=%s", fields.URL, fields.Ref, escapeGitDarPathQuery(fields.Path)), nil
}

// FormatGitReleaseBaseLine returns the git release dependency line without an asset.
func FormatGitReleaseBaseLine(dep *ParsedDarDependency) string {
	if dep == nil || dep.Git.CloneURL == nil {
		return ""
	}
	return FormatGitReleaseLine(gitDependencyCloneURLString(dep.Git.CloneURL), dep.Git.Ref, "")
}

// ExpandGitReleaseDependenciesInYaml expands release entries without assets in a daml.yaml field.
func ExpandGitReleaseDependenciesInYaml(ctx context.Context, yamlPath, fieldName string, rawDeps []*RawDependency) (bool, error) {
	if len(rawDeps) == 0 {
		return false, nil
	}

	expansionInputs, err := resolveGitAliasInputs(yamlPath, fieldName, rawDeps, func(dep *ParsedDarDependency) (string, bool) {
		if dep == nil || !dep.Git.Release || strings.TrimSpace(dep.Git.DarPath) != "" {
			return "", false
		}
		return FormatGitReleaseBaseLine(dep), true
	})
	if err != nil {
		return false, err
	}
	expanded, sourceIndices, err := expandReleaseGitDependenciesRaw(ctx, expansionInputs)
	if err != nil {
		return false, err
	}
	return rewriteYamlDependencyField(yamlPath, fieldName, rawDeps, expanded, sourceIndices)
}

// CanonicalizeGitDependenciesInYaml rewrites git dependency entries in a daml.yaml field
// to canonical git: one-liners (flattening structured git objects).
func CanonicalizeGitDependenciesInYaml(yamlPath, fieldName string, rawDeps []*RawDependency) (bool, error) {
	if len(rawDeps) == 0 {
		return false, nil
	}

	aliasResolved, err := resolveGitAliasInputs(yamlPath, fieldName, rawDeps, func(dep *ParsedDarDependency) (string, bool) {
		if dep == nil || dep.Scheme() != "git" {
			return "", false
		}
		return FormatGitYamlLine(dep), true
	})
	if err != nil {
		return false, err
	}
	canonicalized, changed, err := CanonicalizeRawGitDependencies(aliasResolved)
	if err != nil {
		return false, err
	}
	if !changed && rawDependenciesEqual(rawDeps, aliasResolved) {
		return false, nil
	}
	return rewriteYamlDependencyField(yamlPath, fieldName, rawDeps, canonicalized, nil)
}

// ExpandReleaseGitDependenciesRaw expands release entries with no asset, preserving other entry shapes.
func ExpandReleaseGitDependenciesRaw(ctx context.Context, rawDeps []*RawDependency) ([]*RawDependency, error) {
	expanded, _, err := expandReleaseGitDependenciesRaw(ctx, rawDeps)
	return expanded, err
}

func expandReleaseGitDependenciesRaw(ctx context.Context, rawDeps []*RawDependency) ([]*RawDependency, []int, error) {
	existingAssets := map[string]struct{}{}
	for _, rawDep := range rawDeps {
		line, err := rawDep.Value()
		if err != nil {
			return nil, nil, err
		}
		if !IsGitDependencyLine(line) {
			continue
		}
		dep, err := ParseGitDependency(line)
		if err != nil {
			return nil, nil, err
		}
		if !dep.Git.Release || strings.TrimSpace(dep.Git.DarPath) == "" {
			continue
		}
		key, err := GitLockKeyForDep(dep)
		if err != nil {
			return nil, nil, err
		}
		existingAssets[key] = struct{}{}
	}

	var out []*RawDependency
	var sourceIndices []int
	for sourceIndex, rawDep := range rawDeps {
		line, err := rawDep.Value()
		if err != nil {
			return nil, nil, err
		}
		if !IsGitDependencyLine(line) {
			out = append(out, rawDep)
			sourceIndices = append(sourceIndices, sourceIndex)
			continue
		}
		dep, err := ParseGitDependency(line)
		if err != nil {
			return nil, nil, err
		}
		if !dep.Git.Release {
			out = append(out, rawDep)
			sourceIndices = append(sourceIndices, sourceIndex)
			continue
		}
		if err := githubrelease.ValidateReleaseHost(dep.Git.CloneURL); err != nil {
			return nil, nil, err
		}
		if strings.TrimSpace(dep.Git.DarPath) != "" {
			out = append(out, rawDep)
			sourceIndices = append(sourceIndices, sourceIndex)
			continue
		}
		_, _ = fmt.Fprintf(os.Stderr, "Resolving git release: listing .dar assets for %s release %q\n",
			dep.Git.CloneURL.String(), dep.Git.Ref)
		assets, err := githubrelease.ListDarAssets(ctx, dep.Git.CloneURL, dep.Git.Ref)
		if err != nil {
			return nil, nil, fmt.Errorf("git release %q: %w", dep.Git.Ref, err)
		}
		for _, asset := range assets {
			assetDep := *dep
			assetDep.Git.DarPath = asset
			key, err := GitLockKeyForDep(&assetDep)
			if err != nil {
				return nil, nil, err
			}
			if _, exists := existingAssets[key]; exists {
				continue
			}
			existingAssets[key] = struct{}{}

			expandedLine := FormatGitReleaseLine(gitDependencyCloneURLString(dep.Git.CloneURL), dep.Git.Ref, asset)
			expanded, err := rawDependencyWithValue(rawDep, expandedLine)
			if err != nil {
				return nil, nil, err
			}
			out = append(out, expanded)
			sourceIndices = append(sourceIndices, sourceIndex)
		}
	}
	return out, sourceIndices, nil
}

func resolveGitAliasInputs(yamlPath, fieldName string, rawDeps []*RawDependency, rewrite func(*ParsedDarDependency) (string, bool)) ([]*RawDependency, error) {
	hasAlias := false
	for _, rawDep := range rawDeps {
		line, err := rawDep.Value()
		if err != nil {
			return nil, err
		}
		if _, ok := ArtifactLocationAlias(line); ok {
			hasAlias = true
			break
		}
	}
	if !hasAlias {
		return rawDeps, nil
	}

	pkg, err := Read(yamlPath)
	if err != nil {
		return nil, err
	}
	_, parsedDeps := pkg.RawAndParsed(fieldName)

	out := append([]*RawDependency(nil), rawDeps...)
	for i, rawDep := range rawDeps {
		line, err := rawDep.Value()
		if err != nil {
			return nil, err
		}
		if _, ok := ArtifactLocationAlias(line); !ok {
			continue
		}
		value, ok := rewrite(parsedDeps[line])
		if !ok {
			continue
		}
		out[i], err = rawDependencyWithValue(rawDep, value)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func rewriteYamlDependencyField(yamlPath, fieldName string, original, rewritten []*RawDependency, sourceIndices []int) (bool, error) {
	if rawDependenciesEqual(original, rewritten) {
		return false, nil
	}
	if sourceIndices == nil {
		sourceIndices = make([]int, len(rewritten))
		for i := range sourceIndices {
			sourceIndices[i] = i
		}
	}

	b, err := os.ReadFile(yamlPath)
	if err != nil {
		return false, err
	}

	out, err := yamledit.ReplaceListAnyPreservingComments(b, fieldName, rewritten, sourceIndices)
	if err != nil {
		return false, err
	}

	if err := os.WriteFile(yamlPath, []byte(out), 0644); err != nil {
		return false, err
	}

	return true, nil
}

// CanonicalizeRawGitDependencies rewrites git entries to canonical one-liners.
// Structured git objects are flattened; {value, main-package-id} wrappers are kept.
func CanonicalizeRawGitDependencies(rawDeps []*RawDependency) ([]*RawDependency, bool, error) {
	out := make([]*RawDependency, 0, len(rawDeps))
	changed := false
	for _, rawDep := range rawDeps {
		canonical, itemChanged, err := canonicalizeRawGitDependency(rawDep)
		if err != nil {
			return nil, false, err
		}
		if itemChanged {
			changed = true
		}
		out = append(out, canonical)
	}
	return out, changed, nil
}

func canonicalizeRawGitDependency(raw *RawDependency) (*RawDependency, bool, error) {
	switch {
	case raw.GitStructured != nil:
		line, err := FormatGitStructuredLine(raw.GitStructured)
		if err != nil {
			return nil, false, err
		}
		if !IsGitDependencyLine(line) {
			return raw, false, nil
		}
		canonical, err := CoerceGitDependencyInput(line, GitInputOptions{RequireGitPrefix: true})
		if err != nil {
			return nil, false, err
		}
		return rawDependencyFromValue(canonical), true, nil
	case raw.WithPackageId != nil:
		canonical, itemChanged, err := canonicalizeGitLineValue(raw.WithPackageId.Value)
		if err != nil || !itemChanged {
			return raw, itemChanged, err
		}
		return &RawDependency{WithPackageId: &withPackageId{
			Value:         canonical,
			MainPackageId: raw.WithPackageId.MainPackageId,
		}}, true, nil
	case raw.ValueOnly != nil:
		canonical, itemChanged, err := canonicalizeGitLineValue(*raw.ValueOnly)
		if err != nil || !itemChanged {
			return raw, itemChanged, err
		}
		return rawDependencyFromValue(canonical), true, nil
	default:
		return raw, false, nil
	}
}

func canonicalizeGitLineValue(line string) (string, bool, error) {
	if !IsGitDependencyLine(line) {
		return line, false, nil
	}
	canonical, err := CoerceGitDependencyInput(line, GitInputOptions{RequireGitPrefix: true})
	if err != nil {
		return "", false, err
	}
	return canonical, canonical != line, nil
}

func rawDependencyFromValue(value string) *RawDependency {
	v := value
	return &RawDependency{ValueOnly: &v}
}

func rawDependencyWithValue(raw *RawDependency, value string) (*RawDependency, error) {
	if raw != nil && raw.WithPackageId != nil {
		return &RawDependency{WithPackageId: &withPackageId{
			Value:         value,
			MainPackageId: raw.WithPackageId.MainPackageId,
		}}, nil
	}
	return rawDependencyFromValue(value), nil
}

// MarshalDependencyWithValue writes value, keeping {value, main-package-id} wrappers
// and flattening structured git objects to a one-liner.
func MarshalDependencyWithValue(raw *RawDependency, value string) (string, error) {
	updated, err := rawDependencyWithValue(raw, value)
	if err != nil {
		return "", err
	}
	out, err := yaml.Marshal(updated)
	return string(out), err
}

func rawDependenciesEqual(a, b []*RawDependency) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		ya, err := yaml.Marshal(a[i])
		if err != nil {
			return false
		}
		yb, err := yaml.Marshal(b[i])
		if err != nil {
			return false
		}
		if !bytes.Equal(ya, yb) {
			return false
		}
	}
	return true
}
