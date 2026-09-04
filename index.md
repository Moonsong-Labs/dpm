---
title: Intro
layout: default
nav_order: 1
permalink: /
---

# Intro

`dpm` can take a pre-built `.dar` from Git the same way it already takes one from OCI. Point `daml.yaml` at a file in a repository or an asset on a GitHub Release; `dpm` fetches it and pins it so later builds use the same bytes.

![Demo of Git DAR dependencies in dpm]({{ '/assets/demo.gif' | relative_url }})

How to write those lines is on [Git references]({{ '/git-references.html' | relative_url }}). How to try them is on [Testing]({{ '/testing.html' | relative_url }}). The [technical design]({{ '/technical-design.html' | relative_url }}) is why install fetches and resolve does not.
