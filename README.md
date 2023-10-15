# My Personal Website

This is my repo for my personal website. Feel free to use it.

The stack is planned to be:
- FE: HTMX
- Server: Golang
- Edge-deployed workers: Zig
- DB: Turso

The goal of this project is to learn lots of things. I know nothing about FE and
I'd like to change that. Some tools I use for building/running my code are
- Bazel for building my Golang code
- Nix flakes for isolated, hermetic builds
- Pulumi in Golang for cloud deployments

## 1: Building

You can build the code by running the following commands:
- `nix develop -i` in the root dir of the project
- `bazel run //:gazelle deploy/`

As of now, only the IaC in Golang is built like this. I have yet to add the
Golang BE or Zig edge functions.

## 2: Running

You can run the code locally through any local server provider as it is all 
static files at the moment. I personally use `devd` which is a Golang based
http cli tool with live reloading. It works for now.

The other option, running through Pulumi, is a lot harder. You WILL need a 
domain name. To do so, do the following:
- `nix develop -i`
- `nix-shell -p awscli2`
- `cd ./deploy`
- `pulumi up`
You will probably be missing secrets required when running `pulumi up` so follow
the intructions of the cli tool.

You will likely have to setup more stuff via the dashboard on Cloudflare for
their DNS depending on your hostname provider.

