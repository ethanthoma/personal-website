# My Personal Website

This is my repo for my personal website. Feel free to use it.

The stack is planned to be:
- FE: HTMX
- Server: Golang
- Edge-deployed workers: Golang or zig or rust idk 
- DB: Turso (to be)

The goal of this project is to learn lots of things. I know nothing about FE and
I'd like to change that. Some tools I use for running my code are
- Nix flakes and direnv for dev shells
- Pulumi in Golang for cloud deployments

## 1: Building

At the moment, there is nothing to build, everything is pure html and css.

## 2: Running

You can run the code locally through any local server provider as there are only
static files at the moment.

The other option, running through Pulumi, is a lot harder. You WILL need a 
domain name. My DNS is via CloudFlare and so, the pulumi code assumes that.
If you have it going through CloudFlare and want to store the static images on 
AWS, you can do the following:
- `nix develop`
- `cd ./deploy`
- `pulumi up`
You will probably be missing secrets required when running `pulumi up` so follow
the intructions of the cli tool. You will likely also have to setup more stuff 
via the dashboard on Cloudflare for their DNS depending on your hostname 
provider.

