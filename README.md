<h3 align="center">
    <img 
        src="https://raw.githubusercontent.com/ethanthoma/personal-website/main/services/webserver/public/favicon/android-chrome-512x512.png" 
        width="100"
        alt="Logo"/>
    <br/>
    <a href="https://www.ethanthoma.com/">Personal Website</a>
</h3>

<p align="center">
    <img src="https://img.shields.io/github/last-commit/ethanthoma/personal-website/main?style=for-the-badge&labelColor=%231f1d2e&color=%23c4a7e7">
    <img src="https://img.shields.io/github/actions/workflow/status/ethanthoma/personal-website/deploy.yml?style=for-the-badge&labelColor=%231f1d2e&color=%239ccfd8">
    <img src="https://img.shields.io/github/languages/count/ethanthoma/personal-website?style=for-the-badge&labelColor=%231f1d2e&color=%23ebbcba">
</p>

## GoTH Stack

The backend is written in Go using [templ](https://github.com/a-h/templ) and
using stdlib HTTP.

Styling is done via [tailwindcss](https://github.com/tailwindlabs/tailwindcss).

Reactivity is done thanks to [htmx](https://htmx.org/) and
[surreal](https://github.com/gnat/surreal).

## Building + Running

All building is managed via nix.

Use `nix build .#<name>` to run a build command. The names are:

- #default: this produces the webserver binary
- #container: docker image containing the webserver binary

## Developing

The [make file](./Makefile) in root is setup for running air w/ livereload. It
will run tailwindcss, templ, and air. Simply run `make live`.

The webserver port is set via environment variable `WEBSERVER_PORT`.

## Blog Configuration

The website fetches blog posts from configurable sources using the `BLOG_SOURCE`
environment variable:

Local Directory:

```bash
BLOG_SOURCE="~/projects/blogs/"
BLOG_SOURCE="/absolute/path/to/blogs/"
BLOG_SOURCE="./relative/path/"
```

Git Repositories (Nix-style format):

```bash
BLOG_SOURCE="github:owner/repo"
BLOG_SOURCE="gitlab:owner/repo"
```

Default: If no `BLOG_SOURCE` is set, defaults to my blogs
(`github:ethanthoma/blogs`).

## Blog Post Format

Blog posts are in markdown. They support YAML frontmatter:

```yaml
---
title: "Post Title"
date: "2024-01-15T10:30:00Z"
slug: "post-slug"
---

# Post Title
Content here...
```
