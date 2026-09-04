---
title: Git DAR dependencies
---

# Git DAR dependencies: design and resolution

Git is a second remote source for pre-built `.dar` dependencies, next to OCI. A project can point at a file in a repository or an asset on a GitHub Release; `dpm` fetches it and pins it so later builds use the same bytes.

## The problem

`dpm` already treats remote DARs as first-class dependencies. The existing remote source is OCI. Teams also publish pre-built `.dar` files in Git: a path inside a repository, or an asset on a GitHub Release.

`damlc` does not speak Git (or OCI). It compiles from local files. If a project lists a remote location in `daml.yaml`, something in front of the compiler must:

1. turn that location into a file on disk; and
2. hand the compiler only those local paths.

That front-end is `dpm`. Git support reuses the same lifecycle OCI already uses — `add`, `install`, `update`, `resolve` — rather than inventing a Git-only workflow.

## What we are not doing

This feature fetches a **pre-built** `.dar`. It does not clone a Daml project and build it. If the file is missing, empty, or not a `.dar` at the chosen revision, install fails with that fact. The author of the dependency is responsible for committing or releasing the artifact.

We also do not teach `damlc` the `git:` syntax. The compiler keeps a single input shape: absolute paths in a resolution file written by `dpm`.

## The split that everything else follows

Two different jobs share `daml.yaml`. Mixing them is what makes builds non-reproducible.

**Materialize** (`dpm add`, `dpm install`, `dpm update`) may use the network. It fetches bytes, writes the cache, and may rewrite `daml.yaml` so a moving name becomes a fixed identity.

**Resolve** (`dpm resolve`, and the same lookup `dpm` does before it starts `damlc`) must not use the network and must not rewrite the project. It only answers: given what is declared and what is already in the cache, which local files should the compiler see?

That split already exists for OCI. Git follows it on purpose:

- A branch name is not a build input. The commit it pointed at on the last install is.
- A cold machine that has never run `install` must fail `resolve` in a way that says "run install", not silently clone `main` as it exists today.
- A warm machine with a populated cache should be able to resolve and compile offline.

If resolve fetched, two checkouts of the same unpinned `#main` could compile different bytes, and every `dpm build` would depend on Git being reachable.

The diagram is the relationship between those jobs, not the internals of resolve. Resolve is a lookup. The design is which step is allowed to talk to Git.

<pre class="mermaid">
flowchart LR
  yaml["daml.yaml<br/>git:…#main?path=foo.dar"]
  mat["Materialize<br/>add / install / update"]
  pinned["daml.yaml<br/>git:…#commit?path=foo.dar"]
  cache["Local cache<br/>…/commit/foo.dar"]
  res["Resolve<br/>no network, no rewrite"]
  file["Resolution file<br/>absolute .dar paths"]
  damlc["damlc"]

  yaml --> mat
  mat -->|"fetch + pin"| pinned
  mat -->|"copy .dar"| cache
  pinned --> res
  cache --> res
  res --> file
  file --> damlc
</pre>

<script type="module">
  import mermaid from "https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.esm.min.mjs";
  mermaid.initialize({ startOnLoad: true });
</script>

`damlc` never appears on the left. It only consumes the resolution file. A missing pin or a missing cache file stops at Resolve and tells the operator to materialize; it does not clone as a side effect of asking "what should we compile?"

## How a Git dependency is written

The project file stays a list of strings, in `dependencies` or `data-dependencies`. A Git DAR is one string with a `git:` prefix. There are two shapes, because there are two places a `.dar` actually lives.

**Repository file** — the artifact is a path in the tree at a Git revision:

```text
git:github.com/org/repo#main?path=packages/foo.dar
```

The revision may be a branch, a tag, or a 40-character commit. The path must be a repository-relative `.dar`. Absolute paths, parent segments, and non-`.dar` paths are rejected.

**GitHub Release asset** — the artifact is attached to a GitHub Release, not necessarily present in the tree:

```text
git:github.com/org/repo?release=v1.0.0&asset=foo.dar
```

Omitting `asset` means "every `.dar` on that release". That umbrella form is expanded into one line per asset the first time materialize runs, so later runs have a concrete list.

The two shapes must not be mixed on one line (a release plus a `#ref?path=`). A line is either "file in a repo" or "asset on a release".

Pasted browser URLs (GitHub `/blob/` / `/raw/`, GitLab `/-/blob/`) and a few host-first shorthands are accepted as input and rewritten to the canonical `git:` one-liner. That is authoring convenience, not a second storage format.

`artifact-locations` may hold a **bare** repository URL (host and repo only). The revision and path stay on the dependency line (`@alias#main?path=foo.dar`). A location that already includes a revision or query is rejected: the alias is a repo nickname, not a hidden full dependency.

## Technical decisions

- Use single-line format only
- Pin in the field the author used
- Use HTTPS Git only
- Keep the two `daml.yaml` fields

## How materialize works (install / update / add)

Think of this as turning a name into bytes plus a fixed name.

### 1. Prepare the file

Both fields are walked.

- An umbrella `?release=` line (no asset) is replaced by one line per `.dar` asset on that release. Wrapper entries that carry a main package id keep that wrapper. Assets already listed are not duplicated.
- If listing the release fails, the project file is left as it was (on `add` of a new umbrella line, the temporary line is removed). If listing succeeds and a later download fails, the per-asset lines stay, so a retry does not list the release again.
- Canonical form is written: aliases become full `git:` URIs; recognized pasted URLs become the one-liner.

### 2. Fetch what is missing

For a **repository file**:

- If the revision is already a full commit and a non-empty cached `.dar` exists at the expected path, nothing is cloned.
- Otherwise `dpm` reuses or creates a working clone of that repository (one clone per host/org/repo, kept next to the cache). It fetches the requested revision, checks out the commit that revision currently names, and copies the `.dar` out of the worktree.
- The copy is atomic. The source must be a regular, non-empty `.dar`. A symlink that leaves the worktree is rejected. An empty file in the repository is an error (the publisher did not ship an artifact). An empty file already in the cache is treated as "not cached" and fetched again.

For a **release asset**:

- If the cached file is present and non-empty, it is reused.
- Otherwise the named asset is downloaded from the GitHub Releases API into the cache.

### 3. Pin moving names

After a successful repository fetch, if the declared revision was a branch or tag, that list entry is rewritten to the commit that was just checked out. Already-pinned commits are not re-resolved; `update` only fetches them when the cache file is missing. Release tags are not rewritten.

`dpm update --check` is the read-only counterpart: it succeeds when every repository Git DAR is commit-pinned and cached, and every release asset is cached. It fails on a moving ref, a missing cache file, or an umbrella release that was never expanded. It does not fetch and does not edit `daml.yaml`.

## How resolve works

Resolve is a cache lookup driven by the **already-pinned** project file.

1. Read `daml.yaml`. Do not clone, do not call GitHub, do not edit the file.
2. Walk `dependencies`, then `data-dependencies`.
3. For each Git repository-file line:
   - if the revision is not a 40-character commit, fail — the project is not installed (a branch name is not enough);
   - if the cache does not contain a non-empty file for `(repository, commit, path)`, fail — install (or update) has not materialized it;
   - otherwise emit that file's absolute path.
4. For each Git release line:
   - if there is no asset (still an umbrella), fail — expand first via install or update;
   - if the cache file for `(repository, tag, asset)` is missing or empty, fail;
   - otherwise emit that absolute path.
5. Write those paths into the resolution document, in the matching resolved list. `dpm` then starts `damlc` with that document. The compiler never sees a `git:` string.

That is the whole resolution implementation: **parse the declaration, demand a pin, demand a cache hit, return a path**. The interesting work (network, clone, rewrite) already happened in materialize.

## How the cache is addressed

The cache is content-addressed enough that two projects asking for the same artifact share one file.

- Repository file: `…/cache/git/<host>/<org>/<repo>/<commit>/<path/in/repo.dar>`  
  The commit is the identity. `#main` never appears in this path after a successful install.
- Release asset: `…/cache/git/<host>/<org>/<repo>/<hash-of-tag-and-asset>/<asset>`  
  The tag is not a commit, so the directory name is a hash of tag plus asset name, which keeps two assets on the same release from colliding and keeps the path free of raw tag characters.

Path segments are sanitized so a hostile repository name or `..` in a path cannot write outside the cache root.

The working clone (`…/<repo>/.work/…`) is an implementation detail of fetch. Resolve never looks at it. Only the copied `.dar` under the commit (or release hash) is a resolve input.

## What `dpm update` does differently

`install` is "make the current declaration real". If the line already says a commit and the file is cached, it is a no-op.

`update` is "re-evaluate names that are allowed to move":

- a branch or tag is fetched again and the pin is overwritten if the commit changed;
- a commit pin is left alone, and fetched only to fill a missing cache;
- a release tag is left alone (after any missing umbrella expansion).

That matches OCI: floating tags are refreshed on update; digest pins are not.

## Lockfile

When the existing lockfile switch is on, Git DARs in `dependencies` participate with a stable identity key (repository plus path, or release plus asset) the same way OCI DARs do. The lockfile still does **not** record `data-dependencies`. That is why pinning in `daml.yaml` is the reproducibility mechanism that covers both fields.

## Failure model

Materialize fails closed: unknown host for releases, SSH, mixed shapes, missing path, missing file at that revision, empty source `.dar`, symlink escape, unknown release or asset.

Resolve fails closed: unpinned ref, missing or empty cache, unexpanded umbrella release. The error tells the operator to install or update. It does not repair the project as a side effect of asking "what should we compile?"

## Why this is enough for the compiler

From `damlc`'s point of view nothing Git-specific happened. It receives the same kind of resolution document it already receives for OCI and for local paths: two lists of absolute `.dar` files. Git is a new way for `dpm` to **fill** those lists, not a new compiler feature.
