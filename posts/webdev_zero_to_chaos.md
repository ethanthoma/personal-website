# Webdev: Zero to Chaos

> "I am not a webdev. I started at zero (relative). It only goes down from there."

## Attempt One: Static Site, Labyrinth Infra

I hate overly-complicated slop. I just wanted a simple website. A little bit of 
HTML and CSS should be all I really needed. The site will have some little blogs 
about my research and projects and that's that. All that's required is showing 
the research world that I earnestly care about what I do. And I do.

So, if you are making a static website, a smarter person would choose a simple 
hosting service like Cloudflare or even GitHub pages. Unfortunately, in my last 
term of my undergrad, did a Cloud-based project and became intimate with 
everyone's favourite, AWS.

### Lost in the Clouds

Using AWS sucks. This is not a surprise to anyone. Entire companies exist 
because of how difficult it is to use. Luckily, we can use better ways to 
interact with the cloud ecosystem: Infrastructure as Code (IaC). Does this make 
it more complex? Probably. Does it make the dev experience better? Probably not.
But I did it anyway in case, you know, I wanted to redeploy my website multiple 
times or something.

Originally, I had tried the "great" [AWS CDK](https://aws.amazon.com/cdk/). 
Although better than the haphazard button clicking, the AWS CDK is no saviour. 
Instead, I opted to use [Pulumi](https://www.pulumi.com/). A little `pulumi up` 
and it was up. However, like all things, nothing is free. I wrote it in Go so 
now my build-free code has a whole ass Go compiler and more. But IaC is a one 
time op, so it shouldn't matter?

### Adding Building to a Static Site

Now, this is where I fall prey to the beautiful world of fake complexity. 
Instead of using modern CSS, I decided to use SCSS. What does SCSS do that CSS 
doesn't? IDK. I clearly didn't care either. Of course, now I need my HTML to 
point to the CSS. I also wanted my `pulumi up` to hoist a whole dir to S3 and 
blah blah blah. Now my build step copies my HTML and compiles my SCSS to a build 
dir.

Running scripts imperatively sucks. That's why C has the awful choice of 500 
different Makefile DSLs for building. And I too, don't want to remember the 
commands (or, at that time, learn to use flakes correctly). So I overengineered 
a BASH build script. I even paired it up with a neat `config.json` that my 
script read to know what to copy or not, what to compile or not, etc. At the end 
my config looked like:

```json
{
    "main": "html/index.html",
    "root": "./frontend/",
    "scripts": {
        "build": "./bin/build",
        "deploy": "mirai build && cd infra && aws-admin pulumi up && cd ..",
        "deploy-y": "mirai build && cd infra && aws-admin pulumi up -y && cd .."
    },
    "include": [
        "src/*",
        "vendor"
    ],
    "exclude": [ ]
}
```

And on and on I added more crap to it. It could fetch libraries from URLs into a
vendor directory and a command that wrapped the files in a Docker image with 
nginx to serve it...which I cannot recall why since I hosting it in a bucket 
anyways.

### From Bash to Bazel to Crash

Obviously, building a custom Bash-based build tool to SERVE STATIC FILES is 
unbelievably dumb. Despite the obvious, I persevered. I was smart enough to see 
that Bash was a dumb idea. Not smart enough to see the whole thing was too.

Next on my plan was [Bazel](https://bazel.build/). Bazel + Nix offers a great 
way to do [hermetic builds](https://bazel.build/basics/hermeticity), a real 
benefit to an actual production-piece of software. And so, I spent a couple more 
weeks to wack this crap into my landfill of bad decisions. At the end of this 
multi-week escaped, all I had was a really shit website and an overengineered, 
time-consuming build process. What was the lesson? THINK??? LITERALLY ONE SECOND 
AFTER IMPLEMENTING THE SEVENTH 3RD PARTY DEPENDENCY TOOL OF DOOM!!! JUST THINK.

But yeah, keep it simple. Please.

## Attempt Two: Climbing a Dune

> "Surmounting obstacles is a good thing if you are heading in the right 
direction..."

A couple months post the first tragedy (much like the world wars, it wasn't 
called the first at this point), my friend decided to make their own website 
too. Unfairly, he is a webdev so he knew what to do. I on the other hand, lost 
in some side quest to make myself suffer, I decided I too should try again. But 
like always, not using a decent ecosystem with great tooling. No, that'd be too 
easy...

Like all great websites, I opted to use [OCaml](https://ocaml.org/). So like all 
great engineers, I went down the thorny path with only suffering at the end.
 
### From Opium to Dream

My first Ocaml webserver attempt started with [Opium](https://github.com/rgrinberg/opium), 
a simple, banger of a library. The router a was literally a couple lines of 
code. I had my ugly ass components in ReasonML with their version of JSX 
(TYXML). Luckily, the Ocaml build tool uses Dune which can compile ReasonML 
just as easily.

As I am a Nix user, I used 
[opam-nix](https://www.tweag.io/blog/2023-02-16-opam-nix/). It's great except it 
has to rebuild every dependency every time so unless you are deploying it, it 
sucks ass to use for development. To be fair to Tweag, it's really Opam's fault,
no theirs. At some point, I switched to using [Dream](https://aantron.github.io/dream/) 
instead of Opium. This I regret heavily since it just adds so much bloat and my 
GitHub workflow for deployment takes 13 minutes...
 
### The HOT Stack: HTMX Ocaml Turso

Everyone xitter users' favourite way to webdev: [Htmx](https://htmx.org/); it is 
a small JS library that lets you do fun AJAX stuff in HTML tags instead of being 
forced to use JS to do anything. With a little bit of kleptomancy in the HTMX 
examples and TYXML adjustment, I got a decently presentable, modern webapp. Yay.

At this point, my little blog posts were markdown files on GitHub. This was 
great except my deploy-via-docker image meant I had to redeploy my site to 
change or add or remove a post, incurring a 13 minute cost. Luckily (or not), I
read this [ThePrimeagen tweet](https://twitter.com/ThePrimeagen/status/1686482867809894400)
and give [Turso](https://turso.tech/) a shot. It comes with a bunch of free tier 
goodies so seems like a win to me.
 
### How to Climb a Dune: You Shouldn't

Turso doesn't have an Ocaml SDK.
 
In fact, there isn't even an SDK for SQLite. But, there is a Go SDK...which you 
compile to C...which Ocaml can read...

And so my taks is now to
- get Dune to compile my Go code to C
- get Dune to link my C code
- call the linked C code from Ocaml

However, Dune is hard to use. The docs are kinda hard to parse and this comes 
from my slogging it out with AWS and GCP and every other cloud system's empire 
of unversioned docs.

The first part, compiling Go to C, was pretty easy. Little Makefile, a couple 
lines of Ocaml ctypes, and some Dune rules and we got it. Woot. Getting the 
linking to work probably took me two weeks. I may be an idiot but even then 
NEVER DO THIS. IT WAS IMPOSSIBLE TO FIND DOCUMENTATION FOR IT AND EVERY EXAMPLE 
SIMPLY DIDN'T WORK. Finally, the third step, piss easy. Took maybe 10 minutes 
once the linking worked.

## Attempt Three: GO-d's chosen Language

> "If you already use Go for one thing, why not both."

When I moved from WSL to Nixos on my surface, my ocaml-webserver would no longer 
build. Why? Some platform specific mirage crap in opam-nix that I was too lazy 
to solve. Plus, my website looked so polished for a simple blog. I didn't like 
that. It was time. Time for [grug brain](https://grugbrain.dev/).

### Libraries Suck

What do I mean they suck, I mean it in the [suckless](https://suckless.org/philosophy/) 
sense. My attempts were filled with bloated code that did too much and took too 
long to do anything. So I went with Go. Not only is my Turso database-reading 
code already in Go, Go standard library comes with a simple HTTP webserver and 
HTML string templates.

The last two attempts took a combined four months of inconsistent effort. This 
one? One day. I set up Go using [Gomod2nix](https://www.tweag.io/blog/2021-03-04-gomod2nix/).
I went over what [ThePrimeagen](https://github.com/ThePrimeagen/fem-htmx/tree/main) 
did in his Go + HTMX tutorial (although he uses a library for HTTP) and I wrote 
some simple template strings based off of my old Ocaml templates. Little bit of 
barebones CSS and I had a working site with my Turso hosted blogs in maybe an 
hour. In 30 minutes, I got my Nix flake to package my Go app into a Docker 
Image, setup a GitHub Action to build and host the image, and call the [Render](https://render.com/) 
deploy hook.

### But Why Write This

Simply, I don't know. This was more for me than it is for you. I don't think I 
know more about webdev than I did before I started. What I do know is that good 
engineering is hard. It's so easy to get trapped in solving abstractions and not 
real valued things. At the end of the day, I needed a medium to communicate 
with. I spent most of my time solving how to integrate this tool with this 
library for this cloud service and so on.

And so, from ivory tower (ivory from the research papers I still have to read, 
sorry supervisor), I preach to you, don't build it simply, be simple. I didn't 
like fixing abstractions, I liked shit that worked. It's why I opt for TUIs and 
delete electron apps. I believe the fundamental goal for any long running 
project is ease of maintenance. I shouldn't try to be clever or smart, just be 
grug, I'm sure his stuff is still running and easily maintained to this day.
