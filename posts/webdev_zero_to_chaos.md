# WTF is Webdev

> "I am not a webdev. How hard could it be?"
>
> "As hard as you make it..."

## Attempt One: Static Site, Labyrinth Infra

I hate overly-complicated slop. Every time I have the misfortune of engaging
with webdev content, my more-than-perfect vision glazes over. I just see
dependency, dependency, framework, library, dependency...

What I want is simple: a website. A little bit of HTML, a little bit of CSS,
some markdown, and I'm done. It's all I really need.

And so, if you are smarter than me, choose a simple service that turns your
markdown into HTML and vamos. Unfortunately, I am not smarter than me.

### Lost in the Clouds

At first, I did make a little HTML and CSS (go me). Unfortunately, it stops
there. My first bad idea was, of course, manually hosting my stuff but

**AWS sucks**

Entire companies exist because of its UI/UX hell.

But we are in luck (lie), we can abstract the crap out of it and use the cool,
shiny paradigm known as Infrastructure as Code (IaC). It comes with the
wonderful perks of all declarative languages: it works perfectly (lie).

Originally, I tried [AWS CDK](https://aws.amazon.com/cdk/). Better than
haphazard button clicking! But not by much. Eventually, I ditched it and opted
to use [Pulumi](https://www.pulumi.com/). A much, much better experience and it
came with the added perk (lie) to write it in Golang. So far, not a lost cause
(lie)...

### Adding Building to a Static Site

A wise man would probably ask why. Unfortunately, I am also not wise.

Back in good-old 2016 when I was barely 15, I tried making my own website. I
used CSS and jQuery, the good old days.

Although, my memory is barely functional, I do recall that CSS sucked. A lot.

So my first mission was to see what I could do instead of CSS, under the
assumption that CSS did not get better (it did though) which eventually led me
to more complexity...SCSS.

Now, what can SCSS do that CSS can't? No idea. Literally not a clue.

Anyways, now I add the fun step of compiling (or transpiling or translating or
preprocessing or what you prefer to call it) my SCSS into CSS. Of course a
simple Makefile would be crazy (lie). Clearly, I need an overengineered Bash
build script (lie).

I even made it accepct a "neat" `config.json` that looked like this:

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

And I just kept adding more crap to it:

- It would fetch files from URLs into the vendor directory.
- It had a command that wrapped the files in a Docker image with nginx to serve
  it...which I cannot recall why I did that since I hosting it in a s3 bucket
  anyways.
- The exlude part let me not copy all my code over, which I'm pretty sure I left
  empty...

I'm sure there's more...

### Bash to Bazel to Crash

Obviously, building a custom Bash-based build tool to **SERVE STATIC FILES** is
unbelievably dumb.

Clearly, I needed something more advanced (lie): [Bazel](https://bazel.build/).

Bazel + Nix offers a great way to do
[hermetic builds](https://bazel.build/basics/hermeticity), something I don't
need! So, I went to add this anyway.

At the end of this multi-week escapade, I did it. I made the ugliest,
overengineered, STATIC website. A mess of complex, custom Bazel files and
multiple steps to deploy.

And then, I abandoned it. Yippee.

## Attempt Two: Climbing a Dune

> "Surmounting obstacles is good IF you are heading in the right direction..."

A solid six months past my first webdev trauma, I was chatting away with my
friend.

My friend was thinking of making his own personal website + blog. And despite my
poor gaslighting attempts for him to make it in [OCaml](https://ocaml.org/), I
failed to do (he did try it for a little bit).

Unfortunately, as I am unwise, I fell to my own tricks and convinced myself...

### I like Opium but I can Dream too

My first webserver (ever technically) was made with
[Opium](https://github.com/rgrinberg/opium), a simple, banger of a library. The
router was literally a couple lines of code.

For components, I opted to use ReasonML's version of JSX (TYXML). Twitter and
abroad said the function way is better. Don't worry, I tried it. It's not.

Dune, the Ocaml build-tool, handled the integration of Ocaml and ReasonML for
me.

Eventually I moved to [Dream](https://aantron.github.io/dream/) as it was better
maintained. What a mistake that was. Dream added something like 700
dependencies.

This was a "bit of an issue" as in the Nix world I was using
[opam-nix](https://www.tweag.io/blog/2023-02-16-opam-nix/) which was great
except it had to rebuild every dependency every time. This took a solid 13
minutes for when I was deploying via GitHub workflow. What the...

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

Of course, SQL not really made to hold blobs but not my problem.

### How to Climb a Dune

Turso doesn't have an Ocaml SDK.

We call this a warning sign.

However, there is a Go SDK...which you can compile to C.

Ocaml can read C code...

And so, much like Bazel, I embarked on a multi-week task to ~take a break from
bad decisions~:

- Get Dune to compile my Go code to C.
- Get Dune to link my C code.
- Call the linked C code from Ocaml.

Dune was hard to use. The docs are about as easy and useful as Nix docs. Which
means they're not easy or useful.

I could get Dune to compile my Go code to C easyily. Makefile here, OCaml ctypes
there, done.

Now, linking. What the fuck is linking.

Getting the linking to work took me about two suffering weeks. I may be an idiot
(this post serves as proof) but even then NEVER DO THIS. IT WAS IMPOSSIBLE TO
FIND DOCUMENTATION FOR IT AND EVERY EXAMPLE SIMPLY DIDN'T WORK.

Finally, the third step. Took maybe 10 minutes.

## Attempt Three: GO-d's chosen Language

> "If you already use Go for one thing, why not both."

To be honest, I didn't mind my website. Sure it took 13 minutes to deploy and
every time I ran it locally, but that's the cost of doing business. It wasn't
even all that ugly with my solid 0 years of design experience.

I just had one small, little problem. I couldn't build it.

When I moved from WSL on my surface to a true, NixOS experience. Something,
somewhere broke and it just never worked. More importantly, I was too lazy to
solve it. I also wanted to update the styling a bit and honestly, I was tired of
700 dependencies.

This time, it was time for [Grug brain](https://grugbrain.dev/).

### Libraries Suck

They do.

That could be the whole section. Fundamentally, they are made to be generic
because that's how they are actually useful and generic stuff normally means
more stuff.

More stuff I have to learn, more stuff I have to maintain, more stuff I have to
debug.

**Every line of code I write is technical debt.**

Every library I import is no different. Less is more. That sometimes means
langauges too. So i chose the overly simple (to a fault) and slightly verbose
language of Go.

Not only is my Turso database-reading code already in Go, the Go standard
library comes with a simple HTTP webserver and HTML string templates, so I don't
even have to vendor anything.

What about dev time? My first two webdev attempts took a combined four months of
inconsistent effort. My Golang website? One day.

It took **ONE WHOLE DAY** to do what took me MONTHS.

Furthermore, Go works wonderfully with Nix thanks to
[Gomod2nix](https://www.tweag.io/blog/2021-03-04-gomod2nix/). I translated my
old ReasonML templates to Go Little bit of barebones CSS (nested and layers) and
I had a working site with my Turso hosted blogs in maybe an hour. A mere 30
minutes later, my Nix flake was updated to package my Go webserver into a Docker
image with a GitHub Action to build it and call the
[Render](https://render.com/) deploy hook to deploy it. CD done.

**_NOTE:_** I eventually moved to templ and TailwindCSS to give them try. They
are awesome tbqh!

## To be or not to be Grug

Making something is about as hard as you choose to make it. Obviously, there is
a lower bound (due to the inherent complexity of the task) but anything above
that is what **YOU** chose to add.

And I find that People, subconsciously, love complexity. They just hate the cost
it. It's why TypeScript library authors wrangle types all day and why someone
invented whatever the hell a meta-framework is. Eventually it becomes
_ingrained_. People will reach for React and whatever else to make a simple
static site just because that's the complexity they've grown to accept.

And I think that's bad.

Sure, it's fine as a learning tool but **I believe that the fundamental goal for
any long running project is ease of maintenance.** Stop trying to be clever,
just be Grug. I'm sure the code Grug wrote will still run for the next 10, 50,
and 100 years.
