---
title: Declaring Git dependencies
layout: default
nav_order: 2
---

# Declaring Git dependencies

To use a Git dependency in your project, **declare it in `daml.yaml` under `dependencies` or `data-dependencies`**. The dependency is a DAR file (`.dar`) stored in a **public Git repository**. Git is another remote source, like OCI.

The canonical form is **one string**:

```text
git:<host>/<owner>/<repo>#<ref>?path=<path/inside/the/repo.dar>
```

- `<ref>` is a **branch, a tag, or a commit SHA**.
- `?path=` is **required**, relative to the repo root, and must end in `.dar`.
- A trailing `.git` on the repo is optional. `dpm` drops it.
- **HTTPS only.** SSH clone URLs (`git:ssh://…` and `git@github.com:org/repo`) are rejected. The repository **must be public**.
- Any public Git host reachable over HTTPS works for this form: GitHub, GitLab, Bitbucket, Codeberg, or a self-hosted server.

Add it with `dpm add dar`. **Exactly one** of `--dependencies` or `--data-dependencies` is required:

```shell
dpm add dar --data-dependencies 'git:example.com/my/dars#1.2.3?path=my-package-1.2.3.dar'
```

After install, `daml.yaml` **contains the commit SHA instead of the branch or tag**:

```yaml
data-dependencies:
  - git:example.com/my/dars#82a5467ac5bf4ed78415dee71f7af587a9e7a8f5?path=my-package-1.2.3.dar
```

`dpm add dar` and `dpm install` rewrite a branch or tag to a commit SHA **in the same field you used**. The pin stays under `dependencies` or `data-dependencies`. It is not moved.

## GitHub release assets

For a DAR file attached to a GitHub release, which need not be present in the repository tree:

```text
git:github.com/example/my-dars?release=1.2.3&asset=my-package-1.2.3.dar
```

{: .note }
`?release=` **works only on github.com**. Omit `&asset=` to expand the entry into **one line per `.dar` asset** on that release. On other hosts, use `#<ref>?path=` instead. Release lines are **not rewritten** to a commit SHA.

A line should be *either* a file in a repo *or* an asset on a release. **Using both at the same time returns an error.**

## Repository aliases

To **reuse one repository URL for several DARs**, name it under `artifact-locations` and refer to it as `@alias`. The alias is *only the repo URL*. `#ref` and `?path=` or `?release=` stay on the dependency line. After `dpm install`, the alias is **replaced by a full `git:` line**.

```yaml
artifact-locations:
  "@my-dars":
    url: git:example.com/my/dars

data-dependencies:
  - "@my-dars#1.2.3?path=my-package-1.2.3.dar"
```

A location URL that already carries `#` or `?` is rejected.

## Adding from a browser URL

`dpm add dar` also accepts a browser `raw` or `blob` URL to a `.dar` file (GitHub `…/blob|raw/<ref>/…`, GitLab `…/-/blob/<ref>/…`) and **writes the canonical `git:` line**.

```shell
dpm add dar --data-dependencies \
  'https://github.com/example/my-dars/raw/refs/tags/1.2.3/my-package-1.2.3.dar'
```

{: .note }
**Do not paste a browser URL into `daml.yaml` yourself.** `dpm` only normalizes those URLs on `add`.

## Verify pins

`dpm update --check` verifies that Git dependencies are **installed and match the commit pins** in `daml.yaml`. **It does not fetch and does not edit the file.** `dpm update` re-resolves branch and tag refs and rewrites those pins.

See [Technical design]({{ '/technical-design.html' | relative_url }}) for why resolve does not fetch, and [Testing with a demo project]({{ '/testing.html' | relative_url }}) to try the forms against a demo project.
