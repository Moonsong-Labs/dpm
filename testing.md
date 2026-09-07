---
title: Testing with a demo project
layout: default
nav_order: 4
---

# Testing with a demo project

For functional testing, use the [dpm-git-links-demo](https://github.com/Moonsong-Labs/dpm-git-links-demo) project. It is a small Daml app created for testing purposes. Its `daml.yaml` already declares **two Git DARs**, and its source imports a symbol from each. **If either link fails to resolve, the build fails.**

```yaml
dependencies:
  - git:github.com/Moonsong-Labs/test-daml-hello#master?path=dist/test-daml-hello-sdk-3.5.2-lf-2.2.dar

data-dependencies:
  - git:github.com/canton-network/splice#release-line-0.6.8?path=daml/dars/splice-amulet-0.1.19.dar
```

`Hello.greeting` comes from the first line. `Splice.Amulet.Amulet` comes from the second. The demo README is the longer version of this page.

## Prerequisites

Git DAR support is **not in a released `dpm` yet**. Build `dpm-dev` from the Moonsong Labs fork, branch `proposal/git-dependencies-support`. You need Go 1.25.0 or later.

```bash
git clone --branch proposal/git-dependencies-support git@github.com:Moonsong-Labs/dpm.git
cd dpm
go build -o bin/dpm-dev ./cmd/dpm/
export PATH="$PWD/bin:$PATH"
```

The first fetch needs **network access to `github.com`**. After that the DARs live under `~/.dpm/cache`.

## Quickstart

From the demo checkout:

```bash
dpm-dev install package
dpm-dev build
dpm-dev test
```

`install package` **fetches both Git DARs and pins the movable refs** (`master`, `release-line-0.6.8`) to commit SHAs. A later install on already-pinned lines skips the fetch if the cache is warm.

{: .warning }
**The rewrite of `daml.yaml` is expected.** Restore the unpinned file before you push the demo.

`build` should produce `.daml/dist/dpm-git-links-demo-0.0.1.dar`. `test` runs the scripts in `daml/Tests.daml`.

To prove the links are doing real work, **delete either Git line and build again**. Without splice you get `Could not find module 'Splice.Amulet'`. Without Hello you get `Could not find module 'Hello'`.

## Sample declarations

The demo `samples/` folder has the other declaration forms. Quoted refs there are *inputs* (`master`, `0.6.10`, `main`). **They are not frozen SHAs.**

| Sample | Form |
| --- | --- |
| [browser-url-tag.yaml](https://github.com/Moonsong-Labs/dpm-git-links-demo/blob/main/samples/browser-url-tag.yaml) | GitHub `raw` URL at a tag, via `dpm-dev add dar` |
| [browser-url-branch.yaml](https://github.com/Moonsong-Labs/dpm-git-links-demo/blob/main/samples/browser-url-branch.yaml) | GitHub `raw` URL on a branch |
| [browser-url-commit.yaml](https://github.com/Moonsong-Labs/dpm-git-links-demo/blob/main/samples/browser-url-commit.yaml) | GitHub `blob` URL at a commit |
| [artifact-location-alias.yaml](https://github.com/Moonsong-Labs/dpm-git-links-demo/blob/main/samples/artifact-location-alias.yaml) | `@alias` under `artifact-locations` |
| [release-single-asset.yaml](https://github.com/Moonsong-Labs/dpm-git-links-demo/blob/main/samples/release-single-asset.yaml) | GitHub release, one named asset |
| [release-all-assets.yaml](https://github.com/Moonsong-Labs/dpm-git-links-demo/blob/main/samples/release-all-assets.yaml) | GitHub release, every `.dar` asset |

Release samples use `Moonsong-Labs/daml-finance` (`test-release-0.0.6`). Splice publishes DARs in the tree, not as release assets.

## Known limitations

- **Paste a browser URL into `dpm-dev add dar`, not into `daml.yaml`.** Reading the file expects a `git:` line. A raw `https://` dependency fails with `http dependencies not yet supported`.
- **`dependencies` is stricter than `data-dependencies`.** It requires the same SDK version and LF target across the closure. Splice belongs under `data-dependencies` in this demo for that reason.
- **To re-resolve from scratch, delete `.daml/`.** For a cold cache, also delete `~/.dpm/cache/git`.

The syntax itself is on [Declaring Git dependencies]({{ '/git-references.html' | relative_url }}). Why install fetches and resolve does not is on [Technical design]({{ '/technical-design.html' | relative_url }}).
