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

| Tech  | Stack    |
|-------|----------|
| GO    | Backend  |
| Htmx  | Frontend |
| Turso | Database |

We use [templ](https://github.com/a-h/templ) for templating and [tailwindcss](https://github.com/tailwindlabs/tailwindcss)
for styles. So it is technically more like the GoTTTH stack...

## Building + Running

The nix flake has four derivations:
- #default: this produces the webserver binary
- #container: docker image containing the webserver binary
- #uploader: simple CLI to upload my markdown blogs
- #blob: a WIP blob storage service I plan to use for my images

## Developing

The [make file](./Makefile) in root is setup for running air w/ livereload.
It will run tailwindcss, templ, and air. Simply run `make live`.

> [!TIP]
> You can also locally deply the docker image using `make docker`. 

The webserver port is set in the [flake](./flake.nix). You also need a dotenv file.
It should contain:
- TURSO_DATABASE_URL
- TURSO_AUTH_TOKEN
