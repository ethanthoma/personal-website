# A Gleam Burrito

> tldr; I made Nix builder for Gleam CLIs targeting Erlang

I've been a fan of [Gleam](https://gleam.run/) for a while now. If you're not familiar with it, Gleam 
is a friendly language with a sound type system that compiles to Erlang or JavaScript. 
It brings type safety to the Erlang ecosystem while maintaining its powerful concurrency model.

In fact, I have multiple Gleam libraries (
[effect](https://github.com/ethanthoma/effect), 
[trellis](https://github.com/ethanthoma/trellis),
[lustre_tauri](https://github.com/ethanthoma/lustre_tauri)) and have even contributed 
a little bit the community ([Gleam STD Lib](https://github.com/gleam-lang/stdlib/commit/9d76bea763732dee0358feaeae46840cb75093be), 
[spinner](https://github.com/lpil/spinner/commit/cd55490d8c1698b22de8c6916999c84c8b5cc58f)).

Beyond this, I also took a stab at writing a Gleam application, which led me down 
the path of creating a Nix builder for Gleam projects.

## Distributing a Gleam CLI

> "In reality, you have no users."
>
> — Me, explaining why I did this to someone

This post is about a [little CLI tool](https://github.com/STASER-Lab/cgq) I wrote for my lab 
to create group quizzes on [Canvas](https://www.instructure.com/canvas?utm_source=google&utm_medium=organic&utm_campaign=canvas-redirect).
The challenge wasn't writing the tool itself, but making it accessible to everyone in my lab.

I couldn't just set up a Nix development shell and call it a day. Not everyone 
uses Nix, and I needed something that anyone could download and run without 
dependencies. Ideally, I wanted a self-contained binary that would work on 
multiple platforms. 

The important bit here is that my CLI was using the Erlang target. I was using 
Erlang OTP and the Gleam HTTPC libraries, so I had to ensure my "wrapper" would 
work with the Erlang runtime.

## Erlang and Distribution

> "I wish there was native lang like Gleam. Why is [Flix](https://flix.dev/) on the JVM a;skldfja"
>
> — Also me, but wiser

For those like me who know next to nothing about Erlang, here are some details 
to be aware of: Erlang is both a language and a runtime (BEAM). This makes it 
a bit tricky. 

We can't just compile to a standalone binary like Go or Rust. 

We need to somehow package the Erlang runtime with our application. This makes
things tricky. How do you distribute an Erlang application to users who don't 
want to install Erlang?

## Enter Burrito: Wrapping Erlang with Zig

> "A box for a program that runs in a box..."
>
> — Still me

[Burrito](https://github.com/burrito-elixir/burrito) is a fantastic tool that 
lets you wrap Elixir CLI applications using [Zig](https://ziglang.org/). It 
creates standalone executables by bundling your application with an Erlang 
runtime. You simply modify your Mix project (Elixir's build tool) to include 
Burrito, add some configuration about your target platforms, and voilà; you get 
a CLI tool without external dependencies.

The keen-eyed will notice that I said Elixir and not Erlang. Also, I mentioned 
Mix, which Gleam doesn't use at all. So how do we get from a Gleam project to 
something Burrito can package?

Easy.

## The Mix(Gleam) Solution

> "A second build tool has joined the program."
>
> — Could be anyone

The solution to our problem is [MixGleam](https://github.com/gleam-lang/mix_gleam), 
which is an Elixir archive that lets Mix handle Gleam dependencies and compilation.
It serves as a bridge between Gleam and the Elixir ecosystem.

The process looked something like this:
1. Install MixGleam
1. Create a Mix project that references our Gleam project
1. Configure Mix to compile our Gleam code
1. Use Burrito to package everything up

All I had to do was write some Mix boilerplate by following the README.md, and 
I was done. I had Mix-powered Gleam builds ready to go.

The best part? It _just worked_. I could run the Burrito CLI (with a little Nix 
dev shell magic) and have it generate standalone binaries. The only real 
downside was dealing with Mix and its boilerplate.

## Nixifying Burrito

> "I can now compile my cli tool via mix via nix. All thats left is burrito. WTF am I doing..."
>
> — Me to [Vitor](https://x.com/akiyama_vitor)

One of Nix's strengths is its ability to build pretty much any language in a 
reproducible manner. I realized I could have Nix generate all the boilerplate 
for me, so I wouldn't have to think about it

Just add it to my flake and move on.

Creating the basic structure wasn't too hard. Fetch burrito-elixir from GitHub, 
generate string versions of the Mix files, and then Parse the Gleam `manifest.toml` 
to get dependencies. I've writtent more Nix code to do less.

The real challenge was configuring Mix to "trust" the dependencies provided by 
Nix. First thing was configuring Mix. I had to use rebar3 but Mix wants to fetch 
this for me...which Nix Flakes isn't a fan of. Luckily Nix has a rebar3 package 
and, with a bit of Mix doc reading, I can set it to use the local binary instead.

I then had to deal with the two sets of dependencies: the Gleam ones from the 
project and the Mix ones required by burrito-elixir. The Gleam ones came pre-compiled 
which was fun because the Mix ones did not. This means I had to have to seperate 
flows while convincing the Mix CLI that everything was correct. Fun.

I eventually figured out where to copy the pre-compiled Gleam dependencies. Luckily
the Mix dependencies were pretty easy to do. There is a [tool](https://github.com/ydlr/mix2nix) which is a generator 
for Nix expression for Mix dependencies. It even wraps each dependency in BuildMix so 
that they too are compiled. This means I can just lob them into the same directory with 
the Gleam stuff and it'll work out. It still took me a couple of days of reading 
Mix documentation and tweaking flags to get it working properly. But in the end, it did!

## Little Usage Example

> "Nix users do this to themselves."
>
> — Someone from Twitter

To use the builder in your project is straightforward:

```nix
{
  description = "My Gleam application packaged with Burrito";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    nix-gleam-burrito.url = "github:ethanthoma/nix-gleam-burrito";
  };

  outputs = { self, nixpkgs, flake-utils, nix-gleam-burrito }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system}.extend 
          nix-gleam-burrito.overlays.default;
        
        myApp = pkgs.buildGleamBurrito {
          src = ./.;  # The pname and version will be read from gleam.toml
          # The default target is Linux, override with:
          # target = "macos";
        };
      in {
        packages = {
          default = myApp;
        };
      }
    );
}
```

With this configuration, running nix build will produce a standalone executable 
in your `./result/bin/your-app-name`.

If you're interested in using this builder for your own Gleam projects, it's 
[here](https://github.com/ethanthoma/nix-gleam-burrito). Word of warning: I 
have used it for one project on one platform. No guarantees (but I will fix 
bugs if you make an issue).
