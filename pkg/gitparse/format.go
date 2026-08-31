package gitparse

import (
	"fmt"
	"net/url"
	"strings"
)

var gitDarPathQueryEscaper = strings.NewReplacer("%", "%25", "+", "%2B")

func escapeGitDarPathQuery(darPath string) string {
	return gitDarPathQueryEscaper.Replace(darPath)
}

func gitDependencyCloneURLString(u *url.URL) string {
	if u == nil {
		return ""
	}
	if u.Scheme == "https" {
		return u.Host + strings.TrimSuffix(u.EscapedPath(), ".git")
	}
	return u.String()
}

// FormatGitYamlLine builds the canonical daml.yaml git dependency string.
func FormatGitYamlLine(git GitSource) string {
	if git.CloneURL == nil {
		return ""
	}
	cloneURL := gitDependencyCloneURLString(git.CloneURL)
	if git.Release {
		return FormatGitReleaseLine(cloneURL, git.Ref, git.DarPath)
	}
	return fmt.Sprintf("git:%s#%s?path=%s", cloneURL, git.Ref, escapeGitDarPathQuery(git.DarPath))
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
func FormatGitReleaseBaseLine(git GitSource) string {
	if git.CloneURL == nil {
		return ""
	}
	return FormatGitReleaseLine(gitDependencyCloneURLString(git.CloneURL), git.Ref, "")
}

// DescribeGitFetch returns a short human-readable summary for progress output.
func DescribeGitFetch(git GitSource) string {
	if git.CloneURL == nil {
		return "git dependency"
	}
	if git.Release {
		return fmt.Sprintf("%s release %q asset %q", git.CloneURL.String(), git.Ref, git.DarPath)
	}
	return fmt.Sprintf("%s @ %q path %q", git.CloneURL.String(), git.Ref, git.DarPath)
}
