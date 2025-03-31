# A Gleam Burrito

I've been a fan of Gleam for awhile now.

In fact, I have multiple Gleam libraries (
[effect](https://github.com/ethanthoma/effect), 
[trellis](https://github.com/ethanthoma/trellis),
[lustre_tauri](https://github.com/ethanthoma/lustre_tauri)) and have 
contributed to the community ([Gleam STD Lib](https://github.com/gleam-lang/stdlib/commit/9d76bea763732dee0358feaeae46840cb75093be), 
[spinner](https://github.com/lpil/spinner/commit/cd55490d8c1698b22de8c6916999c84c8b5cc58f)).

Beyond this, I have also taken a stab at writing Gleam apps.

This post is about a [little CLI](https://github.com/STASER-Lab/cgq) I wrote for my lab to create group quizzes on [Canvas](https://www.instructure.com/canvas?utm_source=google&utm_medium=organic&utm_campaign=canvas-redirect).

## The Basics

To be honest, what the CLI is or does isn't super important. What was 
important is that the tool isn't just for me. It was for anyone in my lab to use.
That means I can't just have a Nix devshell and move on. Ideally, it's a self-contained 
binary that anyone can just download and run.

Now, the important bit here is that my CLI was using the Erlang target. I was 
using Erlang OTP and the Gleam HTTPC libraries so I had to ensure my "wrapper" 
works with Erlang.

## Burrito and MixGleam

For those like me, who know next to nothing about Erlang, here is some deets to 
be aware of. Erlang is a language and a runtime. This makes it a little tricky as 
Erlang needs this runtime to, well, run. We can't just _make_ a binary, we have to 
deal with embedding the runtime too...

Enter, [Burrito](https://github.com/burrito-elixir/burrito). Burrito is a cool 
builder that lets you wrap Elixir CLI applications with the power of Zig. You simply
modify your Mix project to include Burrito, a bunch of boilerplate stuff stating 
what targets are valid and boom, you got yourself a CLI without dependencies.

The keen eyed will notice that I said Elixir and not Erlang. Also, I mentioned Mix, 
which Gleam doesn't use at all. Luckily, we can easily solve both of these problems. 
We just need to use Mix to build our Gleam app and then have Burrito wrap it for us.

Easy.

The solution to our problem is to install [MixGleam](https://github.com/gleam-lang/mix_gleam), 
which is just an Elixir archive that lets Mix deal with Gleam deps.

All I had to do was write a bunch of Mix boilerplate by following the README.md and 
I was done. Mix-powered Gleam builds ready to go.

The best part? It _just worked_. I could run the Burrito CLI (with a little bit of Nix devshell magic) 
and have it work right away. The only real downside was having Mix at all (and the boilerplate).

## Nixifying and Dying

Want of Nix's strong suits is its ability to build pretty much every language in 
a reproducible manner. Furthermore, if I get Nix to write the boilerplate for me,
I can just not think about it
