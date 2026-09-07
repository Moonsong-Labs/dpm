---
title: Overview
layout: default
nav_order: 1
permalink: /
---

# Overview

Git-Based DAR Dependencies for dpm (Dev Fund Proposal #105) adds support for resolving Daml dependencies from Git-hosted prebuilt DARs in the `dpm` CLI.

The main idea behind the feature is that `dpm` can take a `.dar` file from Git the same way it already takes one from OCI. Adding a Git-based `.dar` file to `daml.yaml` makes `dpm` fetch it and pin it so later builds use the cached file.

It removes friction in Canton/Daml development caused by manual handling of external DAR dependencies. Today, Daml projects can depend on SDK-provided packages, packages published in an OCI registry, and local DAR files. Custom, external DAR dependencies are often handled through manual downloads, checked-in artifacts, or project-specific scripts.

![Demo of Git DAR dependencies in dpm]({{ '/assets/demo.gif' | relative_url }})

Pinning, install, resolve, and the `git:` syntax are covered in these pages: [Declaring Git dependencies]({{ '/git-references.html' | relative_url }}) for how to write the lines, [Technical design]({{ '/technical-design.html' | relative_url }}) for how fetch and pin work, and [Demo project]({{ '/testing.html' | relative_url }}) for a demo you can run.

## Deliverables

This page compiles technical documentation and design for the functionality. Because the functionality is a natural addition to `dpm` and its existing documentation, the delivery expands the current solution and documentation:

1. [PR to `dpm`](https://github.com/digital-asset/dpm/pull/311): Feature development, technical documentation, and tests for each new functionality added.
2. [PR to Canton Foundation Docs](https://github.com/canton-network/cf-docs/pull/1463): Public-facing documentation for `dpm` users.

Milestone 1 and evidence comments were posted in issue [#666 from `canton-dev-fund`](https://github.com/canton-foundation/canton-dev-fund/issues/666), and the full [proposal is in #105](https://github.com/canton-foundation/canton-dev-fund/pull/105#issuecomment-5411478050).
