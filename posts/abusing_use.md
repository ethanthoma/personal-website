# Abusing Use

> tldr; a simple guide on abusing gleam's `use` keyword to make useful,
> non-obvious programming monstrosities.

[Gleam](https://gleam.run/) is a simple language.

A little too simple for me.

Because I love metaprogramming. It's hella useful. I've wasted many hours
creating cursed zig comptime horrors or writing Python code that operate on type
annotations. But...gleam doesn't have any form of metaprogramming. Sure, the
author of the lang
[wrote about it](https://lpil.uk/blog/how-to-add-metaprogramming-to-gleam/) and
perhaps, once LLMs have replaced all coders, gleam will have it. Till then, I
think is probably better to use a similar lang that does have it (like my
beloved [flix](https://flix.dev/)) if you really need it.

**However**, gleam does have one amazing feature: `use`.

## WTF is `use`

The `use` syntax is a bit of syntactic sugar. What it does is sort-of unwrap
callback functions. Let's taking a common function one would use, `list.map`.

```gleam
pub fn map(list: List(a), with fun: fn(a) -> b) -> List(b) {
  map_loop(list, fun, [])
}
```

The `map` function takes in two parameters, the list it is operating over,
`list` and a mapping function, `fun`. Let's say we had a list of ints we wanted
to increment, we could

```gleam
let ints = [0, 1, 2, 3, 4]

let inc = fn(int: Int) -> Int {
    int + 1
}

echo list.map(ints, inc)
```

which prints out

```
[1, 2, 3, 4, 5]
```

In this example, we created a lambda function (which we named `inc`). Instead,
we can "unwrap" it with `use`:

```gleam
let ints = [0, 1, 2, 3, 4]

echo {
    use int <- list.map(ints, inc)
    int + 1 
}
```

Now what matters is that everything _after_ the `use` is now **inside** the
lambda function. Sort-of like an implicit scope. In otherwords, we _plop_ the
callback scope into the current scope.

## Common Uses

`use` is used in two common ways: `Result`/`Option` types and `bool.guard`.

Gleam uses `Result` for errors and `Option` for `nullables`. Often, one wants to
do something when the function _doesn't_ error. For example:

```gleam
let database: Result(String, Nil) = envoy.get("DB")

result.try(database, fn(database: String) {
    let query: Result(Dynamic, Nil) = sql.query(db, "SELECT * FROM Users;")
    result.map(query, fn(query: Dynamic) {
        let users: List(User) = decode_to_user(decode.list(query))
        echo users
    })
})
```

This is...pretty hard to read. It isn't terrible but as the nesting gets worse
and worse...so does the readability. In fact, this problem has a name:
[callback hell](http://callbackhell.com/).

`use` lets us avoid this. We can rewrite it as:

```gleam
let database: Result(String, Nil) = envoy.get("DB")

use database: String <- result.try(database)

let query: Result(Dynamic, Nil) = sql.query(db, "SELECT * FROM Users;")

use query: Dynamic <- result.map(query)

let users: List(User) = decode_to_user(decode.list(query))

echo users
```

We call this callback heaven.

Another common way for using `use` is for early returns. For example:

```gleam
let should_return_early = True

case should_return_early {
    True -> ReturnedEarly
    False -> ReturnedAsPlanned
}
```

Using `bool.guard`, we can simplify to:

```gleam
let should_return_early = True

use <- bool.guard(should_return_early, ReturnedEarly)

ReturnedAsPlanned
```

The beauty of it is that it _feels_ like an early return. And sure, it only
works with bools but there are libraries that open that up too.

## The `Case` for Ugliness

Aesthetically, gleam's weakness is the case statement.

Simply put, it's **ugly**.

Everyone seems to love gleam's pipes, `|>`, for composing function calls. To
cater for this, a lot of stdlib functions are pretty simple. In fact, a lot of
them simply wrap a case statement.

Case statements kinda suck in that they somehow feel very imperative. In
otherwords, they aren't very composable. Take a ternary statement in gleam:

```gleam
let positive_num = case num > 0 {
    True -> num
    False -> 0
}
```

Where as most languages would have this built-in:

```gleam
let num = -23
let positive_num = ifelse(num > 0, num, 0)
```

Which is supercomposable with `|>` and even `use`. We can emulate it via:

```gleam
fn ifelse(condition, true, false) {
    case condition {
        True -> true
        False -> false
    }
}
```

Which is nice! But then again, all it does is replace a case statement...

Personally, I literally avoid matching against boolean conditions because
statements just feel super janky. That's why I **abuse** use (I typically use
the [given library](https://github.com/inoas/gleam-given), but it is hardly the
_worst_ way to use `use`).

## Beginner's Guide to Abusing `use`

Let us learn about some of the simpler uses. A common pattern with `use` is
`param` and `defer`.

### param

`param` makes its appearance in a couple gleam libraries. It is primarily used
for making curried functions as Gleam doesn't do auto-currying. There used to be
helpers in the stdlib for making functions curried but they have long since been
removed...but mostly because there was no easy way to generalize over the number
of parameters (and Gleam seems to be moving away from the super functional
side). If the lang had some metaprogramming, maybe it the currying-helper
functions (and the former tuple helpers) would still be in the stdlib...

`param` looks like this:

```gleam
fn param(f: fn(a) -> b) -> fn(a) -> b {
  f
}
```

At first look, `param` seems useless; it just returns the function it's given.
But, when we combine it with `use`, we get:

```gleam
let f: fn(List(a)) -> fn(fn(a) -> b) -> List(b) = {
    use list <- param
    use map <- param

    list.map(list, map)
}
```

`use list <- param` effectively says, 'I expect a parameter named list to be
provided here later.' It turns the `use` block into a function that is waiting
for that argument. Basically, it's a curried `list.map`.

I use this all the time. Mostly because gleam doesn't let you destructure tuples
and records in function parameters. For example:

```gleam
fn some_func(#(a, b): Tuple(some_type_a, some_type_b)) { todo }
```

This would not compile. Instead you would have to do this:

```gleam
fn some_func(tuple: Tuple(some_type_a, some_type_b)) {
    let #(a, b) = tuple
    todo
}
```

Which is kind of annoying (if not redundant). We can solve this with

```gleam
let some_func = {
    use #(a, b) <- tuple
    todo
}
```

### defer

The `defer` function is meant for side-effect code. Gleam doesn't have a good
way for typing side-effects (flix my beloved) but there are times where one
needs to do something at the end of function like clean up memory or close a DB
connection.

Take for example:

```gleam
let mem = allocate_some_memory_via_ffi(bytes: 32)

write_to_mem(mem:, value: "hello", offset: 0)

print_mem(mem:)

clean_up_mem(mem:)
```

Now, all of these are side-effect calls (and probably return `Nil`) but, as odin
and zig users learnt, deallocating where you allocate is generally a good idea.
We can do this by defining a `defer`:

```gleam
fn defer(deferred: fn() -> any, current: fn() -> result) -> result {
  let result = current()
  deferred()
  result
}
```

This function takes in the deferred code (i.e. `deferred`) which we call after
the current context, `current`. We can rewrite above as

```gleam
let mem = allocate_some_memory_via_ffi(bytes: 32)
use <- defer(fn() {clean_up_mem(mem:)})

write_to_mem(mem:, value: "hello", offset: 0)

print_mem(mem:)
```

The reason we wrap the deferred code in a function is to make it **lazy**.
Otherwise, we would call it immediately.

## Unwrapping Upwards

The `use` context _wraps_ context. Wouldn't it be useful to _unwrap_?

Introducing, my, and now our, friend, `apply_with`:

```gleam
fn apply_with(to a: fn(a) -> b, with b: fn() -> a) -> b {
  a(b())
}
```

This function runs in the inner context of `use` and then calls the `deferred`
function with result.

A bit of a word-salad but you can think of it as a clever trick that lets us
unwrap our expressions. For example:

```gleam
fn some_func(route) -> Response {
    let result: Result(View, PageError) = case route {
        "home" -> home(route)
        "projects" -> projects(route)
        _ -> Error(InvalidRoute)
    }

    // this is a type error!!!
    use view: View <- result.map(result) 

    case view {
        Home -> "this my home page"
        _ -> "WIP"
    }
}
```

The `result.map` means the that the function still returns a `Result` type (a
`Result(String, PageError)`) and not a `Response`.

A typical solution would be to wrap everything:

```gleam
fn some_func(route) -> Response {
    let body: Result(String, PageError) = {
        let result: Result(View, PageError) = case route {
            "home" -> home(route)
            "projects" -> projects(route)
            _ -> Error(InvalidRoute)
        }

        use view: View <- result.map(result)

        case view {
            Home -> "this my home page"
            _ -> "WIP"
        }
    }

    case body {
        Ok(body) -> Response(200, body)
        Error(page_error) -> 
            Response(400, string.inspect(page_error))
    }
}
```

Which is...ew.

Instead, let's use our `apply_with`:

```gleam
fn some_func(route) -> Response {
    use <- apply_with({
        use body: Result(String, PageError) <- param

        case body {
            Ok(body) -> Response(200, body)
            Error(page_error) -> 
                Response(400, string.inspect(page_error))
        }
    })

    let result: Result(View, PageError) = case route {
        "home" -> home(route)
        "projects" -> projects(route)
        _ -> Error(InvalidRoute)
    }

    use view: View <- result.map(result)

    case view {
        Home -> "this my home page"
        _ -> "WIP"
    }
}
```

et voilà. We still get the niceness of using `use` with `result.map` but instead
of the function returning the mapped result, we unwrap it with our `apply_with`,
turning it into a `Response`.

Admittedly, the order in which the code runs is confusing. The `use` syntax
tends to wraps types as we go _down_ the function but our `apply_with` does the
opposite; it unwraps as we go _up_. You can mix-and-match a whole lot of these
to cause maximal psychic damage to Gleam noobies. Aka, the abuse.

## `use`-based UI

But, we can still go further.

Recently, I have been building an UI library for a Gleam mobile app (more blogs
about this to come...). The problem I had is that the only UI API (in Gleam)
that I know of is [lustre](https://github.com/lustre-labs/lustre)'s

```gleam
fn view(model) {
  let count = int.to_string(model)

  div([], [
    button([on_click(Incr)], [text(" + ")]),
    p([], [text(count)]),
    button([on_click(Decr)], [text(" - ")])
  ])
}
```

(from their readme)

And, to be honest, I hate it.

It is also how people in OCaml recommend doing webdev as well...and I hated it
then too. In fact, I setup `dune` to compile
[reasonml](https://reasonml.github.io/) with a JSX preprocessor just to **not**
do this.

Unfortunately, Gleam doesn't have anything preprocessor like (metaprogramming
rears its ugly head again). Some libraries get around this by shipping an
application that will do preprocessing for you, like turning SQL files into
Gleam types with decoders.

Instead, I decided to go full on the other way. There is no need to conform to
HTML and CSS dogma. We don't have to map one-to-one with the web's many poor
choices.

The lustre example above shows how to do 2 buttons with text in the middle. My
UI library is like so:

```gleam
fn view(model) {
    let count = int.to_string(model)

    use append, done <- list

    use <- append({
        use <- has(button(Incr))
        text(" + ")
    })

    use <- append(text(count))

    use <- append({
        use <- has(button(Decr))
        text(" - ")
    })

    done
}
```

This, this I like.

This code isn't as nested. We clearly see we are making a list and append to it
three times. We can pass expressions without having everything cramped up next
to each other in a list. The `done` value stops the user from mixing UI element
specific functions, like the `has` function, in the top level scope, making the
API safe and pretty.

Another thing is attributes. In lustre, it is a list that the user defines, the
first parameter passed to all the elements in the example. My UI library is done
like so:

```gleam
use <- has(button(ChangeView(Tower)))
use <- font_size(20)
use <- color(Background, color.green)
text("Tower")
```

This makes a button with the text "Tower", have a font size of 20, and a
background color of green. You can even apply it for things like orientation and
gap:

```gleam
use <- gap(10.0)
use <- orient(Horizontal)
use <- pad(All, 10.0)
use <- color(Background, color.blue)

use append, done <- list

use <- append({
use <- has(button(ChangeView(init())))
    use <- font_size(20)
    use <- color(Background, color.red)
    text("Rest Area")
})

use <- append({
    use <- has(button(ChangeView(Tower)))
    use <- font_size(20)
    use <- color(Background, color.green)
    text("Tower")
})

done
```

Of course, pretty APIs almost always come with the library dev cost. This is the
cost:

```gleam
pub opaque type Done(msg) {
  Done(List(Element(msg)))
}

pub fn list(
  f: fn(fn(Element(msg), fn() -> Done(msg)) -> Done(msg), Done(msg)) ->
    Done(msg),
) -> element.Element(msg) {
  let Done(elements) = {
    use first, done <- f(_, Done([]))
    let Done(rest) = done()
    Done([first, ..rest])
  }
  group(elements)
}

pub fn group(elements children: List(Element(msg))) {
  element.Element(
    tag: "group",
    attributes: attribute.new(),
    on: option.None,
    children: element.Many(children),
  )
}
```

Beautiful type signatures all around.

What makes it complex is simple. All we are doing is returning another callback
function...in the callback function. Easy.

Lets break it down real quick. The `list` function takes a single callback, `f`.
This callback has two parameters: `append` and `done`. `append` is a function
that takes in the element we want to append and a callback (so we can use
`use`).

The callback expects the user to return a `Done` type. The user can do this in
two ways. One, they use `append` again, as its return type is also `Done`. This
lets us nest our `append` calls. Or two, we simply return the `done` we were
provided by the `list` function, which is also the only way to end the `list`
`use` context. I.e. safety ensured. This is why we make `Done` opaque.

## A Forewarning (AKA Abuse Has Consequences)

The main problem with `use` is that it confuses readers.

Of course, with time and like all things, it becomes easier to parse. This has
the added bonus where you eventually start making incredibly cursed functions
that fulfill your perverse desires.

Mix in some cold, hard mutability and you can write code that seemingly does
nothing and do too much:

```gleam
let current_ids =
    option.map(object, {
        use object: render.Object <- util.param
        let id = object.id

        use _ <- function.tap(set.insert(current_ids, id))

        use <- bool.guard(cache.has_checksum(object:), Nil)

        use <- util.defer(fn() { cache.set(object:) })

        use <- bool.guard(cache.has(id:), Nil)

        render.create(object:)
    })
    |> option.unwrap(current_ids)
```

Mwah, perfection.

This has a lot of mutable state. Worse yet, there is no type system in Gleam to
represent this (or effects or regions). Perhaps one day I will figure out a nice
way to type guard it, like a region type (or even Gleam's own
[Ref library](https://github.com/lpil/javascript-mutable-reference)).

But the point is to **manage your abuse**. Too far, and you'll be caught out,
dazed and unmaintainable. Too little, and you will be crawling your way through
callback hell.

I imagine that if [Louis](https://lpil.uk/) ever saw my code, he would nuke this
feature in a heartbeat...
