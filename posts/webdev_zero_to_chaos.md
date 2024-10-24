# WTF is Webdev

> "I am not a webdev. I started at zero (relatively). It only goes down from there..."

## Attempt One: Static Site, Labyrinth Infra

I hate overly-complicated slop. Every time I have the misfortune of engaging with
webdev content, my more-than-perfect vision glazes over. I just see dependency, 
dependency, framework, library, dependency...

What I want is simple: a website. A little bit of HTML, a little bit of CSS, 
some markdown, and I'm done. It's all I really need. 

And so, if you are smarter than me, choose a simple service that turns your 
markdown into HTML and vamos. Unfortunately, I am not smarter than me.

### Lost in the Clouds

At first, I did make a little HTML and CSS (go me). Unfortunately, it stops 
there. My first bad idea was, of course, manually hosting my stuff but

> using AWS sucks. 

Entire companies exist because of its UI/UX hell. 

But we are in luck (RUN), we can abstract the crap out of it and use the shiny 
paradigm Infrastructure as Code (IaC). It comes with the wonderful perks of 
all declarative languages: it works perfectly until it doesn't work at all.

Originally, I tried [AWS CDK](https://aws.amazon.com/cdk/). Better than 
haphazard button clicking but not by much. Eventually, I opted to use [Pulumi](https://www.pulumi.com/). 
A much better experience and I got an opportunity to write it in Golang. So far,
not a lost cause...

### Adding Building to a Static Site

Don't. Just don't. However, I did.

Back in good-old 2016 when I was barely 15, I tried making my own website using 
CSS and jQuery. All I recall was that CSS sucked. A lot. So instead of seeing if 
modern CSS was less shit, I did the usual thing where we assume more complexity 
is better: I decided to use SCSS. 

What does SCSS do that CSS doesn't? No clue. Didn't know then, still don't.

Now I add the fun step of compiling/transpiling/translating/preprocessing my 
SCSS into CSS. Of course a simple Makefile would do? WRONG. Clearly I need to 
overengineer a Bash build script. It came with a neat `config.json` that looked 
like this:

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

(mirai is the name of my bash build-tool I made just for this...)

I just kept adding more crap to it. It would fetch files from URLs into the
vendor directory and had a command that wrapped the files in a Docker image with 
nginx to serve it...which I cannot recall why I did that since I hosting it in 
a s3 bucket anyways.

### Bash to Bazel to Crash

Obviously, building a custom Bash-based build tool to SERVE STATIC FILES is 
unbelievably dumb. 

Clearly, I needed something more advanced: [Bazel](https://bazel.build/). 
Bazel + Nix offers a great way to do [hermetic builds](https://bazel.build/basics/hermeticity), 
something I don't need. And so, I went to add this crap into my collection of bad 
decisions.

At the end of this multi-week escapade, I did it. I made the ugliest, 
overengineered, STATIC website. A mess of complex, custom Bazel files and multiple
steps to deploy. Pretty sure I abandoned the website right after I got Bazel to 
work...

## Attempt Two: Climbing a Dune

> "Surmounting obstacles is a good thing if you are heading in the right 
direction..."

A solid six months past my first webdev trauma, my friend decided to make their 
own website. Despite my poor gaslighting attempts for him to make it in [OCaml](https://ocaml.org/), 
I failed...but I did convince myself...and so back to suffering ago, yippee

### I like Opium but I can Dream too

My first webserver (ever technically) was made with [Opium](https://github.com/rgrinberg/opium), 
a simple, banger of a library. The router a was literally a couple lines of code. 

For components, I opted to use ReasonML's version of JSX (TYXML). Dune, the Ocaml
build-tool, handled the integration for me. People tried to convince me the 
function composition way is better but I disagree. HTML is recognizable and easier
to translate across frameworks and libraries.

Eventually I moved to [Dream](https://aantron.github.io/dream/) as it was better 
maintained. What a mistake. Dream added something like 700 dependencies. This was 
a "bit of an issue" as in the NixOS world I was using [opam-nix](https://www.tweag.io/blog/2023-02-16-opam-nix/) 
which is great except it has to rebuild every dependency every time. This took a 
solid 13 minutes for when I was deploying via GitHub workflow. What the fuck...

### The HOT Stack: HTMX Ocaml Turso

Everyone xitter users' favorite way to webdev: [Htmx](https://htmx.org/); it is 
a small JS library that lets you do fun AJAX stuff in HTML tags instead of being 
forced to use JS to do anything. With a little bit of kleptomancy in the HTMX 
examples and TYXML adjustment, I got a decently presentable, modern webapp. Yay.

Originally, I was just updating my blogs by pushing them as files to GitHub but 
with the 13 minute cost, it sucked a lot. Luckily (or not), I read this 
[ThePrimeagen tweet](https://twitter.com/ThePrimeagen/status/1686482867809894400)
and decided to use [Turso](https://turso.tech/) to store my blogs and metadata. 
It comes with a bunch of free tier goodies so seems like a win to me.
 
### How to Climb a Dune: You Shouldn't

Turso doesn't have an Ocaml SDK.
 
This is the sign to stop and look for a better solution. However, there is a Go 
SDK...which you can compile to C...which Ocaml can read...

And so, much like Bazel, I embarked on a multi-week task to
- get Dune to compile my Go code to C,
- get Dune to link my C code, and
- call the linked C code from Ocaml.

However, Dune is hard to use. The docs are about as easy and useful as Nix docs,
so not at all. Getting Dune to compile Go to C was super easy. Makefile here, 
Ocaml ctypes there, done.

Now, linking, what the fuck is linking. Getting the linking to work probably 
took me about two weeks. I may be an idiot (for which this post serves as proof) 
but even then NEVER DO THIS. IT WAS IMPOSSIBLE TO FIND DOCUMENTATION FOR IT AND 
EVERY EXAMPLE SIMPLY DIDN'T WORK. 

Finally, the third step. Took maybe 10 minutes.

## Attempt Three: GO-d's chosen Language

> "If you already use Go for one thing, why not both."

Now, I didn't mind the slow, Ocaml website of doom. However, when I moved from 
WSL to NixOS on my surface, I just couldn't get it build anymore. Something to 
do with some platform specific mirage crap in opam-nix that I was too lazy 
to solve. Plus, my website looked a little too polished for a simple blog. I 
didn't like that. 

This time, however, it was time for [grug brain](https://grugbrain.dev/).

### Libraries Suck

They just do. Libraries should be as low level as possible. Fundamentally, they 
are made to be generic to be more useful, but generic stuff normally means more 
stuff. More stuff I have to learn, maintain, debug, etc. Every line of code I 
write is technical debt, every library I import is no different, except that one 
import line now represents hundreds to thousands to tens of thousands of debt.

Part of this, less is more ideal is Golang. An overly simple, slightly verbose 
language (I was on my Zig and Odin arc so it fits).

Not only is my Turso database-reading code already in Go, Go standard library 
comes with a simple HTTP webserver and HTML string templates which is awesome.

The first two attempts took a combined four months of inconsistent effort. This 
one? One day. 

It took ONE DAY to do what took me months of wrangling complexity...

Go works wonderfully on Nix with [Gomod2nix](https://www.tweag.io/blog/2021-03-04-gomod2nix/). 
I wrote some simple template strings based off of my old Ocaml templates. Little 
bit of barebones CSS (nested and layers FTW) and I had a working site with my 
Turso hosted blogs in maybe an hour. 30 minutes later my Nix flake could package 
my Go webserver into a Docker Image and I setup a GitHub Action to build and 
host the image and call the [Render](https://render.com/) deploy hook.

> **_NOTE:_** I eventually moved to templ and TailwindCSS to give them try. They are awesome tbqh!

## But Why Write This

Making something is about as hard as you make it. There is a lower bound due to 
the complexity of the task but anything above that is YOU, not the problem.

People love complexity. They hate the cost it. It's why Typescript library 
authors wrangle types all day and we have whatever the hell a meta-framework is.
Eventually it becomes ingrained. People will reach for React and whatever else 
to make a static site just cause that's the complexity they've grown to accept.

**I believe the fundamental goal for any long running project is ease of maintenance.**

Stop trying to be clever or smart, just be grug. I'm sure his stuff will still 
run and be maintained for the next 10, 50, 100 years.
