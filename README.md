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
    <img src="https://img.shields.io/github/actions/workflow/status/ethanthoma/personal-website/docker.yml?style=for-the-badge&labelColor=%231f1d2e&color=%239ccfd8">
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
- #uploader: simple CLI to upload my markdown blogs

## Developing

The [make file](./Makefile) in root is setup for running air w/ livereload. It
will run tailwindcss, templ, and air. Simply run `make live`.

> [!TIP]
> You can also locally deploy the docker image using `make docker`.

The webserver port is set in the [flake](./flake.nix). You also need a dotenv
file. It should contain:

- TURSO_DATABASE_URL
- TURSO_AUTH_TOKEN
