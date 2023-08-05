# My Personal Portfolio Website

This is FOR ME but anyone can use it. Although idk why you would. The stack 
is/planned to be HTMX, SASS, s3 via Pulumi in GO, and zig Cloudflare workers. I
have not thought about what to do for database storage but that is a 'future me'
problem...probably Turso or some other, cheap, edge database provider.

## 0: Requirements

I use a custom build and install script for my frontend in this code. I did this
for fun and not to be practical but it works for now, lol.

The scripts are in bash and requires you to have the following installed:
- `mirai` (this can be found in my dot files repo)
- `jq`
- `rsync`
- `sass`

Once these are installed, you should give execution rights to `mirai` which can 
be done via `chmod +x`. You should do this for both files in the `./bin/` dir:
`build` and `install`.

## 1: Building

After that, you can build the code by running `mirai install` and then `mirai 
build` in the root directory of the project. This should A) download the `htmx`
javascript file for the project into `./frontend/vendors/` and compile the code 
in `./frontend/src/` to `./dist/`.

## 2: Running

There are two ways to run the static files (two ways I do, there are countless
others). Either through Docker or via Pulumi.

### 2.1: Docker

To run through Docker is easy. Once you have Docker installed, simply run the
command `mirai start`. This will build the image and deploy it to a container
named website. I don't have any experience with Docker so I am sure there are 
better ways than this. If you are interested in the command that is ran, check
the `./config.json` file which contains all the commands that `mirai` can run.

## 2.2 Pulumi

Running with Pulumi is a lot harder. Firstly, you will need to install two CLIs:
- `AWS CLI 3`
- `Pulumi`

Afterwards, setup your AWS creditials on the CLI. This can be done by following
their docs. Next, you will have to do the same for Pulumi with your Pulumi 
account. Finally, you will need two more things: a domain name and a Cloudflare
account.

When you have those, you can fill out the secrets in the 
`./infra/static-files/pulumi.dev.yaml`. Once that is setup, you should be able 
to run `mirai deploy` from the root directory. Follow the instructions and 
everything should more or less work.

It is likely you will have to setup more stuff on Cloudflare for their DNS 
depending on who your hostname provider is.

