---
title: Intro
layout: default
nav_order: 1
permalink: /
---

# Intro

These are the pull toghether docs of the feature implemented in https://github.com/digital-asset/dpm/pull/311.
The idea behind the feature is that dpm could take a `.dar` from Git the same way it already takes one from OCI.
Adding a Git-based `.dar` to `daml.yaml` makes `dpm` fetch it and pin it so later builds use the cached file.

![Demo of Git DAR dependencies in dpm]({{ '/assets/demo.gif' | relative_url }})

Pinning, install, resolve, and the `git:` syntax are covered in these pages: [Git references]({{ '/git-references.html' | relative_url }}) for how to write the lines, [Technical design]({{ '/technical-design.html' | relative_url }}) for how fetch and pin work, and [Testing]({{ '/testing.html' | relative_url }}) for a demo you can run.
