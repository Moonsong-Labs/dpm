---
title: Git references
---

{% include nav.html %}

# Git references

You can depend on a `.dar` stored in a public Git repository. Git is another remote source, like OCI. Put the location under `dependencies` or `data-dependencies`.

The canonical form is one string:

```text
git:<host>/<owner>/<repo>#<ref>?path=<path/inside/the/repo.dar>
```

- `<ref>` is a branch, a tag, or a commit SHA.
- `?path=` is required, relative to the repo root, and must end in `.dar`.
- A trailing `.git` on the repo is optional; `dpm` drops it.
- HTTPS only. SSH clone URLs (`git:ssh://…` and `git@github.com:org/repo`) are rejected. The repository must be public.
- Any public Git host reachable over HTTPS works for this form — GitHub, GitLab, Bitbucket, Codeberg, or a self-hosted server.

Add it with `dpm add dar`. Exactly one of `--dependencies` or `--data-dependencies` is required:

```shell
dpm add dar --data-dependencies 'git:example.com/my/dars#1.2.3?path=my-package-1.2.3.dar'
```

Resulting `daml.yaml` after install — the branch or tag is rewritten to a commit SHA:

```yaml
data-dependencies:
  - git:example.com/my/dars#82a5467ac5bf4ed78415dee71f7af587a9e7a8f5?path=my-package-1.2.3.dar
```

`dpm add dar` and `dpm install` rewrite a branch or tag to a commit SHA in the same field you used. The pin stays under `dependencies` or `data-dependencies`; it is not moved.

## GitHub releases

For a `.dar` attached to a GitHub Release (not necessarily present in the tree):

```text
git:github.com/example/my-dars?release=1.2.3&asset=my-package-1.2.3.dar
```

`?release=` is **github.com only**. Omit `&asset=` to expand the entry into one line per `.dar` on that release. On other hosts, use `#<ref>?path=` instead. Release lines are not rewritten to a commit SHA.

The two shapes must not be mixed on one line. A line is either a file in a repo or an asset on a release.

## Aliases

To reuse one repository URL for several DARs, name it under `artifact-locations` and refer to it as `@alias`. The alias is only the repo URL; `#ref` and `?path=` / `?release=` stay on the dependency line. After `dpm install`, the alias is replaced by a full `git:` line.

```yaml
artifact-locations:
  "@my-dars":
    url: git:example.com/my/dars

data-dependencies:
  - "@my-dars#1.2.3?path=my-package-1.2.3.dar"
```

A location URL that already carries `#` or `?` is rejected.

## Browser URLs

`dpm add dar` also accepts a browser `raw` / `blob` URL to a `.dar` (GitHub `…/blob|raw/<ref>/…`, GitLab `…/-/blob/<ref>/…`) and writes the canonical `git:` line. Do not paste a browser URL into `daml.yaml` yourself — `dpm` only normalizes those URLs on `add`.

```shell
dpm add dar --data-dependencies \
  'https://github.com/example/my-dars/raw/refs/tags/1.2.3/my-package-1.2.3.dar'
```

## Checking pins

`dpm update --check` verifies that Git dependencies are installed and match the commit pins in `daml.yaml`. It does not fetch and does not edit the file. `dpm update` re-resolves branch and tag refs and rewrites those pins.

See [Technical design]({{ '/' | relative_url }}) for why resolve does not fetch, and [Testing]({{ '/testing.html' | relative_url }}) to try the forms against a demo project.
