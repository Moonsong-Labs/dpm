---
title: Overview
layout: default
nav_order: 1
permalink: /
---

# Overview

Git-Based DAR Dependencies for dpm ([Dev Fund Proposal #105](https://github.com/canton-foundation/canton-dev-fund/pull/105#issuecomment-5411478050)) adds support for **resolving Daml dependencies from Git-hosted prebuilt DAR files (`.dar`)** in the `dpm` CLI.

![Demo of Git DAR dependencies in dpm]({{ '/assets/demo.gif' | relative_url }})

The main idea behind the feature is that `dpm` can **take a DAR file from Git the same way it already takes one from OCI**. Adding a Git-based DAR file to `daml.yaml` makes `dpm` fetch it and **pin it** so later builds use the cached file.

## Content

- **[Declaring Git dependencies]({{ '/git-references.html' | relative_url }})**: how to write the `git:` lines, use repository aliases and GitHub release assets, and verify pins.
- **[Technical design]({{ '/technical-design.html' | relative_url }})**: how fetch and pin work, and why resolve never uses the network.
- **[Testing with a demo project]({{ '/testing.html' | relative_url }})**: a demo you can build and run against real Git DAR dependencies.

## Deliverables

This page compiles technical documentation and design for the functionality. Because the functionality is a natural addition to `dpm` and its existing documentation, **the delivery expands the current solution and documentation**:

1. **[PR to `dpm`](https://github.com/digital-asset/dpm/pull/311)**: Feature development, technical documentation, and tests for each new functionality added.
2. **[PR to Canton Foundation Docs](https://github.com/canton-network/cf-docs/pull/1463)**: Public-facing documentation for `dpm` users.

**Milestone 1** and evidence comments were posted in issue [#666 from `canton-dev-fund`](https://github.com/canton-foundation/canton-dev-fund/issues/666), and the full [proposal is in #105](https://github.com/canton-foundation/canton-dev-fund/pull/105#issuecomment-5411478050).
